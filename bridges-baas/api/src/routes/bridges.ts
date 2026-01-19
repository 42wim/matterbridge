// Bridge Management Routes
// Connect/disconnect users to Slack, Telegram, Discord

import { FastifyInstance } from 'fastify';
import { PrismaClient, BridgeType, BridgeStatus } from '@prisma/client';
import { z } from 'zod';
import { synapse } from '../services/synapse.js';
import { credentials } from '../services/credentials.js';

const prisma = new PrismaClient();

// Bridge bot Matrix IDs
const BRIDGE_BOTS: Record<string, string> = {
  SLACK: `@slackbot:${process.env.MATRIX_SERVER_NAME}`,
  TELEGRAM: `@telegrambot:${process.env.MATRIX_SERVER_NAME}`,
  DISCORD: `@discordbot:${process.env.MATRIX_SERVER_NAME}`
};

// Bridge login commands
const BRIDGE_LOGIN_COMMANDS: Record<string, string> = {
  SLACK: '!slack login',
  TELEGRAM: '!tg login',
  DISCORD: '!discord login'
};

export async function bridgeRoutes(app: FastifyInstance) {
  // Get available bridges for user
  app.get('/', {
    preHandler: [(app as any).authenticate]
  }, async (request: any) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { bridgeConnections: true }
    });

    if (!user) throw new Error('User not found');

    // Define available bridges based on plan
    const allBridges = ['SLACK', 'TELEGRAM', 'DISCORD'];
    const planLimits: Record<string, number> = {
      FREE: 1,
      STARTER: 3,
      PRO: 10,
      ENTERPRISE: 100
    };

    const limit = planLimits[user.plan] || 1;
    const connected = user.bridgeConnections.filter(b => b.status === 'CONNECTED').length;

    return {
      available: allBridges.map(bridge => ({
        type: bridge,
        name: bridge.charAt(0) + bridge.slice(1).toLowerCase(),
        botId: BRIDGE_BOTS[bridge],
        connected: user.bridgeConnections.some(
          b => b.bridgeType === bridge && b.status === 'CONNECTED'
        )
      })),
      plan: user.plan,
      limit,
      connected,
      canConnectMore: connected < limit
    };
  });

  // Initialize bridge connection
  app.post('/:bridgeType/connect', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const bridgeType = (request.params as any).bridgeType.toUpperCase() as BridgeType;

    if (!['SLACK', 'TELEGRAM', 'DISCORD'].includes(bridgeType)) {
      return reply.status(400).send({ error: 'Invalid bridge type' });
    }

    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { bridgeConnections: true }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Check plan limits
    const planLimits: Record<string, number> = {
      FREE: 1,
      STARTER: 3,
      PRO: 10,
      ENTERPRISE: 100
    };
    const limit = planLimits[user.plan] || 1;
    const connected = user.bridgeConnections.filter(b => b.status === 'CONNECTED').length;

    if (connected >= limit) {
      return reply.status(403).send({
        error: 'Bridge limit reached',
        message: `Your ${user.plan} plan allows ${limit} bridge(s). Please upgrade to connect more.`
      });
    }

    // Check if already connected
    const existing = user.bridgeConnections.find(b => b.bridgeType === bridgeType);
    if (existing?.status === 'CONNECTED') {
      return reply.status(400).send({ error: 'Bridge already connected' });
    }

    // Get user's Matrix token
    const matrixToken = await credentials.getMatrixToken(user.matrixUserId);
    if (!matrixToken) {
      return reply.status(401).send({ error: 'Matrix session expired' });
    }

    // Create or update bridge connection record
    const connection = await prisma.bridgeConnection.upsert({
      where: {
        userId_bridgeType: {
          userId: user.id,
          bridgeType
        }
      },
      create: {
        userId: user.id,
        bridgeType,
        status: 'PENDING'
      },
      update: {
        status: 'PENDING',
        lastError: null,
        errorCount: 0
      }
    });

    // Get or create DM room with bridge bot
    const botId = BRIDGE_BOTS[bridgeType];

    // Return instructions for the mobile app
    return {
      connectionId: connection.id,
      bridgeType,
      status: 'PENDING',
      instructions: {
        step1: `Open a chat with ${botId}`,
        step2: `Send the command: ${BRIDGE_LOGIN_COMMANDS[bridgeType]}`,
        step3: 'Follow the prompts to authenticate',
        botId,
        loginCommand: BRIDGE_LOGIN_COMMANDS[bridgeType]
      },
      // For Slack, we can provide OAuth URL
      ...(bridgeType === 'SLACK' && {
        oauthUrl: await generateSlackOAuthUrl(user.id)
      })
    };
  });

  // Check bridge connection status
  app.get('/:bridgeType/status', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const bridgeType = (request.params as any).bridgeType.toUpperCase() as BridgeType;

    const connection = await prisma.bridgeConnection.findUnique({
      where: {
        userId_bridgeType: {
          userId: request.user.sub,
          bridgeType
        }
      }
    });

    if (!connection) {
      return { status: 'NOT_CONNECTED', bridgeType };
    }

    return {
      bridgeType,
      status: connection.status,
      remoteUsername: connection.remoteUsername,
      connectedAt: connection.connectedAt,
      lastError: connection.lastError
    };
  });

  // Disconnect bridge
  app.post('/:bridgeType/disconnect', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const bridgeType = (request.params as any).bridgeType.toUpperCase() as BridgeType;

    const user = await prisma.user.findUnique({
      where: { id: request.user.sub }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Update connection status
    await prisma.bridgeConnection.update({
      where: {
        userId_bridgeType: {
          userId: user.id,
          bridgeType
        }
      },
      data: {
        status: 'DISCONNECTED',
        disconnectedAt: new Date()
      }
    });

    // Delete stored credentials
    await credentials.deleteBridgeCredentials(user.id, bridgeType);

    // Note: The actual logout from the bridge happens via Matrix
    // User should send !slack logout, !tg logout, or !discord logout

    return {
      success: true,
      message: `To complete disconnection, send the logout command to the bridge bot`,
      logoutCommand: bridgeType === 'SLACK' ? '!slack logout'
        : bridgeType === 'TELEGRAM' ? '!tg logout'
        : '!discord logout'
    };
  });

  // Webhook for bridge status updates (called by bridge bots or monitoring)
  app.post('/webhook/status', async (request, reply) => {
    // Verify webhook secret
    const secret = request.headers['x-webhook-secret'];
    if (secret !== process.env.BRIDGE_WEBHOOK_SECRET) {
      return reply.status(401).send({ error: 'Invalid webhook secret' });
    }

    const schema = z.object({
      matrixUserId: z.string(),
      bridgeType: z.enum(['SLACK', 'TELEGRAM', 'DISCORD']),
      status: z.enum(['CONNECTED', 'DISCONNECTED', 'ERROR']),
      remoteUserId: z.string().optional(),
      remoteUsername: z.string().optional(),
      error: z.string().optional()
    });

    const body = schema.parse(request.body);

    // Find user by Matrix ID
    const user = await prisma.user.findUnique({
      where: { matrixUserId: body.matrixUserId }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Update bridge connection
    await prisma.bridgeConnection.upsert({
      where: {
        userId_bridgeType: {
          userId: user.id,
          bridgeType: body.bridgeType as BridgeType
        }
      },
      create: {
        userId: user.id,
        bridgeType: body.bridgeType as BridgeType,
        status: body.status as BridgeStatus,
        remoteUserId: body.remoteUserId,
        remoteUsername: body.remoteUsername,
        connectedAt: body.status === 'CONNECTED' ? new Date() : undefined,
        lastError: body.error
      },
      update: {
        status: body.status as BridgeStatus,
        remoteUserId: body.remoteUserId,
        remoteUsername: body.remoteUsername,
        connectedAt: body.status === 'CONNECTED' ? new Date() : undefined,
        disconnectedAt: body.status === 'DISCONNECTED' ? new Date() : undefined,
        lastError: body.error,
        errorCount: body.status === 'ERROR' ? { increment: 1 } : undefined
      }
    });

    return { success: true };
  });

  // Telegram-specific: Submit phone number
  app.post('/telegram/phone', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const schema = z.object({
      phoneNumber: z.string().regex(/^\+[1-9]\d{1,14}$/)
    });

    const { phoneNumber } = schema.parse(request.body);

    // Store in session for the bridge bot to use
    await credentials.storeBridgeLoginSession(
      request.user.sub,
      'TELEGRAM',
      { phoneNumber, step: 'AWAITING_CODE' }
    );

    return {
      success: true,
      message: 'Phone number stored. Send the login command to the Telegram bridge bot.',
      nextStep: 'Send !tg login ' + phoneNumber + ' to @telegrambot'
    };
  });

  // Telegram-specific: Submit verification code
  app.post('/telegram/code', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const schema = z.object({
      code: z.string().length(5)
    });

    const { code } = schema.parse(request.body);

    const session = await credentials.getBridgeLoginSession(
      request.user.sub,
      'TELEGRAM'
    );

    if (!session) {
      return reply.status(400).send({ error: 'No pending login session' });
    }

    return {
      success: true,
      message: 'Code received. Send it to the Telegram bridge bot.',
      nextStep: 'Send !tg code ' + code + ' to @telegrambot'
    };
  });
}

// Generate Slack OAuth URL
async function generateSlackOAuthUrl(userId: string): Promise<string> {
  const state = `slack_${userId}_${Date.now()}`;

  // Store state for verification
  await credentials.storeOAuthState(state, { userId, bridgeType: 'SLACK' });

  const params = new URLSearchParams({
    client_id: process.env.SLACK_CLIENT_ID || '',
    scope: 'users:read,channels:read,channels:history,chat:write',
    redirect_uri: `${process.env.API_PUBLIC_URL}/webhooks/slack/oauth`,
    state
  });

  return `https://slack.com/oauth/v2/authorize?${params}`;
}
