// Authentication Routes
import { FastifyInstance } from 'fastify';
import { PrismaClient } from '@prisma/client';
import bcrypt from 'bcryptjs';
import { z } from 'zod';
import { synapse } from '../services/synapse.js';
import { credentials } from '../services/credentials.js';

const prisma = new PrismaClient();

// Validation schemas
const registerSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  username: z.string().min(3).max(20).regex(/^[a-z0-9_]+$/),
  displayName: z.string().optional()
});

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string()
});

export async function authRoutes(app: FastifyInstance) {
  // Register new user
  app.post('/register', async (request, reply) => {
    try {
      const body = registerSchema.parse(request.body);

      // Check if user exists
      const existing = await prisma.user.findFirst({
        where: {
          OR: [
            { email: body.email },
            { matrixUserId: `@${body.username}:${process.env.MATRIX_SERVER_NAME}` }
          ]
        }
      });

      if (existing) {
        return reply.status(400).send({ error: 'User already exists' });
      }

      // Create Matrix user first
      const matrixUser = await synapse.createUser({
        username: body.username,
        password: body.password,
        displayName: body.displayName || body.username,
        email: body.email
      });

      // Store Matrix token securely
      await credentials.storeMatrixToken(
        matrixUser.userId,
        matrixUser.accessToken,
        matrixUser.deviceId
      );

      // Create user in our database
      const passwordHash = await bcrypt.hash(body.password, 12);
      const user = await prisma.user.create({
        data: {
          email: body.email,
          passwordHash,
          matrixUserId: matrixUser.userId,
          matrixDeviceId: matrixUser.deviceId,
          displayName: body.displayName || body.username
        }
      });

      // Generate JWT
      const token = app.jwt.sign({
        sub: user.id,
        email: user.email,
        matrixUserId: user.matrixUserId
      });

      return {
        user: {
          id: user.id,
          email: user.email,
          matrixUserId: user.matrixUserId,
          displayName: user.displayName
        },
        token,
        matrix: {
          userId: matrixUser.userId,
          accessToken: matrixUser.accessToken,
          deviceId: matrixUser.deviceId,
          homeserver: process.env.MATRIX_PUBLIC_URL || process.env.SYNAPSE_URL
        }
      };
    } catch (error: any) {
      console.error('Registration error:', error);

      if (error instanceof z.ZodError) {
        return reply.status(400).send({ error: 'Validation failed', details: error.errors });
      }

      return reply.status(500).send({ error: error.message || 'Registration failed' });
    }
  });

  // Login
  app.post('/login', async (request, reply) => {
    try {
      const body = loginSchema.parse(request.body);

      const user = await prisma.user.findUnique({
        where: { email: body.email }
      });

      if (!user) {
        return reply.status(401).send({ error: 'Invalid credentials' });
      }

      const valid = await bcrypt.compare(body.password, user.passwordHash);
      if (!valid) {
        return reply.status(401).send({ error: 'Invalid credentials' });
      }

      // Get Matrix token from cache or re-login
      let matrixToken = await credentials.getMatrixToken(user.matrixUserId);

      if (!matrixToken) {
        // Token expired, need to re-login to Matrix
        const loginResponse = await fetch(
          `${process.env.SYNAPSE_URL}/_matrix/client/v3/login`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              type: 'm.login.password',
              identifier: { type: 'm.id.user', user: user.matrixUserId.split(':')[0].slice(1) },
              password: body.password,
              initial_device_display_name: 'BaaS API'
            })
          }
        );

        if (!loginResponse.ok) {
          return reply.status(401).send({ error: 'Matrix login failed' });
        }

        const loginData = await loginResponse.json();
        matrixToken = {
          accessToken: loginData.access_token,
          deviceId: loginData.device_id
        };

        // Cache the new token
        await credentials.storeMatrixToken(
          user.matrixUserId,
          matrixToken.accessToken,
          matrixToken.deviceId
        );
      }

      // Generate JWT
      const token = app.jwt.sign({
        sub: user.id,
        email: user.email,
        matrixUserId: user.matrixUserId
      });

      return {
        user: {
          id: user.id,
          email: user.email,
          matrixUserId: user.matrixUserId,
          displayName: user.displayName,
          plan: user.plan
        },
        token,
        matrix: {
          userId: user.matrixUserId,
          accessToken: matrixToken.accessToken,
          deviceId: matrixToken.deviceId,
          homeserver: process.env.MATRIX_PUBLIC_URL || process.env.SYNAPSE_URL
        }
      };
    } catch (error: any) {
      console.error('Login error:', error);

      if (error instanceof z.ZodError) {
        return reply.status(400).send({ error: 'Validation failed', details: error.errors });
      }

      return reply.status(500).send({ error: 'Login failed' });
    }
  });

  // Refresh Matrix token
  app.post('/refresh-matrix', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    const matrixToken = await credentials.getMatrixToken(user.matrixUserId);

    if (!matrixToken) {
      return reply.status(401).send({ error: 'Matrix session expired, please login again' });
    }

    return {
      matrix: {
        userId: user.matrixUserId,
        accessToken: matrixToken.accessToken,
        deviceId: matrixToken.deviceId,
        homeserver: process.env.MATRIX_PUBLIC_URL || process.env.SYNAPSE_URL
      }
    };
  });

  // Get current user
  app.get('/me', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: {
        bridgeConnections: true
      }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    return {
      id: user.id,
      email: user.email,
      matrixUserId: user.matrixUserId,
      displayName: user.displayName,
      plan: user.plan,
      bridges: user.bridgeConnections.map(b => ({
        type: b.bridgeType,
        status: b.status,
        remoteUsername: b.remoteUsername,
        connectedAt: b.connectedAt
      })),
      createdAt: user.createdAt
    };
  });

  // Logout (invalidate tokens)
  app.post('/logout', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub }
    });

    if (user) {
      // Delete cached Matrix token
      const redis = (await import('ioredis')).default;
      const client = new redis(process.env.REDIS_URL);
      await client.del(`matrix:${user.matrixUserId}`);
      await client.quit();
    }

    return { success: true };
  });
}
