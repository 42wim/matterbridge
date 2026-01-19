// Admin Routes
// Platform administration and monitoring

import { FastifyInstance } from 'fastify';
import { PrismaClient } from '@prisma/client';
import { synapse } from '../services/synapse.js';

const prisma = new PrismaClient();

// Admin authentication middleware
async function adminAuth(request: any, reply: any) {
  await (request as any).jwtVerify();

  // Check if user is admin (you can use a database flag or env var)
  const adminEmails = (process.env.ADMIN_EMAILS || '').split(',');
  if (!adminEmails.includes(request.user.email)) {
    return reply.status(403).send({ error: 'Admin access required' });
  }
}

export async function adminRoutes(app: FastifyInstance) {
  // All routes require admin auth
  app.addHook('preHandler', adminAuth);

  // Dashboard stats
  app.get('/stats', async () => {
    const [
      totalUsers,
      activeUsers,
      bridgeConnections,
      dedicatedServers,
      synapseStats
    ] = await Promise.all([
      prisma.user.count(),
      prisma.user.count({
        where: {
          bridgeConnections: {
            some: { status: 'CONNECTED' }
          }
        }
      }),
      prisma.bridgeConnection.groupBy({
        by: ['bridgeType', 'status'],
        _count: true
      }),
      prisma.userMatrixServer.groupBy({
        by: ['status'],
        _count: true
      }),
      synapse.getServerStats().catch(() => null)
    ]);

    return {
      users: {
        total: totalUsers,
        active: activeUsers
      },
      bridges: bridgeConnections.reduce((acc, b) => {
        if (!acc[b.bridgeType]) acc[b.bridgeType] = {};
        acc[b.bridgeType][b.status] = b._count;
        return acc;
      }, {} as Record<string, Record<string, number>>),
      servers: dedicatedServers.reduce((acc, s) => {
        acc[s.status] = s._count;
        return acc;
      }, {} as Record<string, number>),
      matrix: synapseStats
    };
  });

  // List all users
  app.get('/users', async (request) => {
    const { page = '1', limit = '50', search = '' } = request.query as any;

    const skip = (parseInt(page) - 1) * parseInt(limit);

    const [users, total] = await Promise.all([
      prisma.user.findMany({
        where: search ? {
          OR: [
            { email: { contains: search } },
            { matrixUserId: { contains: search } }
          ]
        } : undefined,
        include: {
          bridgeConnections: true,
          matrixServer: true
        },
        skip,
        take: parseInt(limit),
        orderBy: { createdAt: 'desc' }
      }),
      prisma.user.count({
        where: search ? {
          OR: [
            { email: { contains: search } },
            { matrixUserId: { contains: search } }
          ]
        } : undefined
      })
    ]);

    return {
      users: users.map(u => ({
        id: u.id,
        email: u.email,
        matrixUserId: u.matrixUserId,
        plan: u.plan,
        bridgesConnected: u.bridgeConnections.filter(b => b.status === 'CONNECTED').length,
        hasServer: !!u.matrixServer,
        createdAt: u.createdAt
      })),
      pagination: {
        page: parseInt(page),
        limit: parseInt(limit),
        total,
        pages: Math.ceil(total / parseInt(limit))
      }
    };
  });

  // Get specific user details
  app.get('/users/:userId', async (request, reply) => {
    const { userId } = request.params as any;

    const user = await prisma.user.findUnique({
      where: { id: userId },
      include: {
        bridgeConnections: true,
        matrixServer: true
      }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Get Matrix user info
    const matrixUser = await synapse.getUser(user.matrixUserId).catch(() => null);

    return {
      ...user,
      passwordHash: undefined, // Don't expose
      matrixInfo: matrixUser
    };
  });

  // Update user plan
  app.patch('/users/:userId/plan', async (request, reply) => {
    const { userId } = request.params as any;
    const { plan } = request.body as any;

    if (!['FREE', 'STARTER', 'PRO', 'ENTERPRISE'].includes(plan)) {
      return reply.status(400).send({ error: 'Invalid plan' });
    }

    const user = await prisma.user.update({
      where: { id: userId },
      data: { plan }
    });

    // Log the action
    await prisma.auditLog.create({
      data: {
        userId,
        action: 'PLAN_CHANGE',
        resource: 'user',
        resourceId: userId,
        metadata: { newPlan: plan }
      }
    });

    return { success: true, plan: user.plan };
  });

  // Deactivate user
  app.post('/users/:userId/deactivate', async (request, reply) => {
    const { userId } = request.params as any;

    const user = await prisma.user.findUnique({
      where: { id: userId }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Deactivate Matrix user
    await synapse.deactivateUser(user.matrixUserId);

    // Mark as deactivated (you might want to add a status field)
    await prisma.bridgeConnection.updateMany({
      where: { userId },
      data: { status: 'DISCONNECTED' }
    });

    await prisma.auditLog.create({
      data: {
        userId,
        action: 'USER_DEACTIVATE',
        resource: 'user',
        resourceId: userId
      }
    });

    return { success: true };
  });

  // Bridge health status
  app.get('/bridges/health', async () => {
    // Check if bridge containers are running
    const bridges = ['slack', 'telegram', 'discord'];
    const health: Record<string, any> = {};

    for (const bridge of bridges) {
      try {
        // Ping bridge health endpoint (if available)
        // Or check docker container status
        health[bridge] = {
          status: 'healthy',
          lastCheck: new Date().toISOString()
        };
      } catch (err: any) {
        health[bridge] = {
          status: 'unhealthy',
          error: err.message,
          lastCheck: new Date().toISOString()
        };
      }
    }

    return health;
  });

  // Restart bridge
  app.post('/bridges/:bridgeType/restart', async (request, reply) => {
    const { bridgeType } = request.params as any;
    const bridge = bridgeType.toLowerCase();

    if (!['slack', 'telegram', 'discord'].includes(bridge)) {
      return reply.status(400).send({ error: 'Invalid bridge type' });
    }

    // Execute docker restart
    const { exec } = await import('child_process');
    const { promisify } = await import('util');
    const execAsync = promisify(exec);

    try {
      await execAsync(`docker restart mautrix-${bridge}`);

      await prisma.auditLog.create({
        data: {
          action: 'BRIDGE_RESTART',
          resource: 'bridge',
          resourceId: bridge
        }
      });

      return { success: true, message: `${bridge} bridge restarted` };
    } catch (err: any) {
      return reply.status(500).send({ error: err.message });
    }
  });

  // Get audit logs
  app.get('/audit-logs', async (request) => {
    const { page = '1', limit = '100', userId, action } = request.query as any;
    const skip = (parseInt(page) - 1) * parseInt(limit);

    const where: any = {};
    if (userId) where.userId = userId;
    if (action) where.action = action;

    const [logs, total] = await Promise.all([
      prisma.auditLog.findMany({
        where,
        skip,
        take: parseInt(limit),
        orderBy: { createdAt: 'desc' }
      }),
      prisma.auditLog.count({ where })
    ]);

    return {
      logs,
      pagination: {
        page: parseInt(page),
        limit: parseInt(limit),
        total,
        pages: Math.ceil(total / parseInt(limit))
      }
    };
  });
}
