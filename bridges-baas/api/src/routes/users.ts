// User Management Routes
import { FastifyInstance } from 'fastify';
import { PrismaClient } from '@prisma/client';
import { z } from 'zod';
import bcrypt from 'bcryptjs';
import { synapse } from '../services/synapse.js';

const prisma = new PrismaClient();

export async function userRoutes(app: FastifyInstance) {
  // Update user profile
  app.patch('/profile', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const schema = z.object({
      displayName: z.string().min(1).max(50).optional(),
      avatarUrl: z.string().url().optional()
    });

    const body = schema.parse(request.body);

    const user = await prisma.user.update({
      where: { id: request.user.sub },
      data: {
        displayName: body.displayName,
        avatarUrl: body.avatarUrl
      }
    });

    return {
      id: user.id,
      email: user.email,
      displayName: user.displayName,
      avatarUrl: user.avatarUrl
    };
  });

  // Change password
  app.post('/change-password', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const schema = z.object({
      currentPassword: z.string(),
      newPassword: z.string().min(8)
    });

    const { currentPassword, newPassword } = schema.parse(request.body);

    const user = await prisma.user.findUnique({
      where: { id: request.user.sub }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Verify current password
    const valid = await bcrypt.compare(currentPassword, user.passwordHash);
    if (!valid) {
      return reply.status(401).send({ error: 'Current password is incorrect' });
    }

    // Update password in database
    const newHash = await bcrypt.hash(newPassword, 12);
    await prisma.user.update({
      where: { id: user.id },
      data: { passwordHash: newHash }
    });

    // Update Matrix password
    await synapse.resetPassword(user.matrixUserId, newPassword);

    return { success: true };
  });

  // Get user's bridge connections
  app.get('/bridges', {
    preHandler: [(app as any).authenticate]
  }, async (request: any) => {
    const connections = await prisma.bridgeConnection.findMany({
      where: { userId: request.user.sub }
    });

    return connections.map(c => ({
      bridgeType: c.bridgeType,
      status: c.status,
      remoteUsername: c.remoteUsername,
      connectedAt: c.connectedAt,
      lastError: c.lastError
    }));
  });

  // Upgrade plan (integrate with your payment system)
  app.post('/upgrade', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const schema = z.object({
      plan: z.enum(['STARTER', 'PRO', 'ENTERPRISE']),
      paymentToken: z.string().optional() // From Stripe/etc
    });

    const { plan, paymentToken } = schema.parse(request.body);

    // TODO: Integrate with Stripe/payment processor
    // For now, just update the plan
    const user = await prisma.user.update({
      where: { id: request.user.sub },
      data: { plan }
    });

    return {
      success: true,
      plan: user.plan,
      message: `Upgraded to ${plan} plan`
    };
  });

  // Delete account
  app.delete('/account', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Deactivate Matrix user
    try {
      await synapse.deactivateUser(user.matrixUserId);
    } catch (err) {
      console.error('Failed to deactivate Matrix user:', err);
    }

    // Delete from our database (cascades to bridge connections)
    await prisma.user.delete({
      where: { id: user.id }
    });

    return { success: true };
  });
}
