// Webhook Routes
// Handle OAuth callbacks and bridge status updates

import { FastifyInstance } from 'fastify';
import { PrismaClient, BridgeStatus } from '@prisma/client';
import { credentials } from '../services/credentials.js';

const prisma = new PrismaClient();

export async function webhookRoutes(app: FastifyInstance) {
  // Slack OAuth callback
  app.get('/slack/oauth', async (request, reply) => {
    const { code, state, error } = request.query as any;

    if (error) {
      // Redirect to app with error
      return reply.redirect(`${process.env.APP_URL}/bridges/slack?error=${error}`);
    }

    if (!code || !state) {
      return reply.status(400).send({ error: 'Missing code or state' });
    }

    // Verify state and get user
    const oauthData = await credentials.consumeOAuthState(state);
    if (!oauthData) {
      return reply.status(400).send({ error: 'Invalid or expired state' });
    }

    try {
      // Exchange code for token
      const tokenResponse = await fetch('https://slack.com/api/oauth.v2.access', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          client_id: process.env.SLACK_CLIENT_ID || '',
          client_secret: process.env.SLACK_CLIENT_SECRET || '',
          code,
          redirect_uri: `${process.env.API_PUBLIC_URL}/webhooks/slack/oauth`
        })
      });

      const tokenData = await tokenResponse.json();

      if (!tokenData.ok) {
        throw new Error(tokenData.error || 'Failed to get token');
      }

      // Store credentials
      await credentials.storeBridgeCredentials({
        bridgeType: 'SLACK',
        userId: oauthData.userId,
        accessToken: tokenData.access_token,
        metadata: {
          teamId: tokenData.team?.id,
          teamName: tokenData.team?.name,
          scope: tokenData.scope
        }
      });

      // Update bridge connection
      await prisma.bridgeConnection.upsert({
        where: {
          userId_bridgeType: {
            userId: oauthData.userId,
            bridgeType: 'SLACK'
          }
        },
        create: {
          userId: oauthData.userId,
          bridgeType: 'SLACK',
          status: 'CONNECTED',
          remoteUsername: tokenData.authed_user?.id,
          connectedAt: new Date(),
          metadata: {
            teamId: tokenData.team?.id,
            teamName: tokenData.team?.name
          }
        },
        update: {
          status: 'CONNECTED',
          remoteUsername: tokenData.authed_user?.id,
          connectedAt: new Date(),
          metadata: {
            teamId: tokenData.team?.id,
            teamName: tokenData.team?.name
          }
        }
      });

      // Redirect to success page
      return reply.redirect(`${process.env.APP_URL}/bridges/slack?success=true`);
    } catch (err: any) {
      console.error('Slack OAuth error:', err);
      return reply.redirect(`${process.env.APP_URL}/bridges/slack?error=${encodeURIComponent(err.message)}`);
    }
  });

  // Bridge status webhook (called by monitoring or bridge bots)
  app.post('/bridge-status', async (request, reply) => {
    // Verify webhook secret
    const secret = request.headers['x-webhook-secret'];
    if (secret !== process.env.BRIDGE_WEBHOOK_SECRET) {
      return reply.status(401).send({ error: 'Invalid webhook secret' });
    }

    const { matrixUserId, bridgeType, event, data } = request.body as any;

    // Find user
    const user = await prisma.user.findUnique({
      where: { matrixUserId }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Handle different events
    switch (event) {
      case 'login_success':
        await prisma.bridgeConnection.upsert({
          where: {
            userId_bridgeType: {
              userId: user.id,
              bridgeType
            }
          },
          create: {
            userId: user.id,
            bridgeType,
            status: 'CONNECTED',
            remoteUserId: data.remoteUserId,
            remoteUsername: data.remoteUsername,
            connectedAt: new Date()
          },
          update: {
            status: 'CONNECTED',
            remoteUserId: data.remoteUserId,
            remoteUsername: data.remoteUsername,
            connectedAt: new Date(),
            lastError: null
          }
        });
        break;

      case 'login_failed':
        await prisma.bridgeConnection.update({
          where: {
            userId_bridgeType: {
              userId: user.id,
              bridgeType
            }
          },
          data: {
            status: 'ERROR',
            lastError: data.error,
            errorCount: { increment: 1 }
          }
        });
        break;

      case 'logout':
      case 'disconnected':
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
        break;

      case 'rate_limited':
        await prisma.bridgeConnection.update({
          where: {
            userId_bridgeType: {
              userId: user.id,
              bridgeType
            }
          },
          data: {
            status: 'RATE_LIMITED',
            lastError: data.message
          }
        });
        break;

      default:
        console.log(`Unknown bridge event: ${event}`);
    }

    return { success: true };
  });

  // Health check endpoint for bridges to ping
  app.get('/health', async () => {
    return {
      status: 'ok',
      timestamp: new Date().toISOString()
    };
  });

  // Matrix Appservice transaction webhook
  // This is called by Synapse when events occur
  app.put('/matrix/transactions/:txnId', async (request, reply) => {
    const hsToken = request.headers['authorization']?.replace('Bearer ', '');

    // Verify homeserver token
    if (hsToken !== process.env.MATRIX_HS_TOKEN) {
      return reply.status(401).send({ errcode: 'M_UNAUTHORIZED' });
    }

    const { events } = request.body as any;

    // Process events (for tracking bridge activity)
    for (const event of events || []) {
      // Log bridge-related events for analytics
      if (event.type === 'm.room.message' && event.sender?.includes('bot')) {
        console.log('Bridge message:', event.sender, event.room_id);
      }
    }

    return {};
  });

  // Stripe webhook (for subscription management)
  app.post('/stripe', async (request, reply) => {
    const sig = request.headers['stripe-signature'];

    // In production, verify Stripe signature
    // const event = stripe.webhooks.constructEvent(rawBody, sig, endpointSecret);

    const event = request.body as any;

    switch (event.type) {
      case 'customer.subscription.created':
      case 'customer.subscription.updated':
        // Update user plan based on subscription
        const subscription = event.data.object;
        const userId = subscription.metadata.userId;

        if (userId) {
          const planMap: Record<string, string> = {
            'price_starter': 'STARTER',
            'price_pro': 'PRO',
            'price_enterprise': 'ENTERPRISE'
          };

          const plan = planMap[subscription.items.data[0]?.price?.id] || 'FREE';

          await prisma.user.update({
            where: { id: userId },
            data: { plan: plan as any }
          });
        }
        break;

      case 'customer.subscription.deleted':
        // Downgrade to free
        const canceledSub = event.data.object;
        const canceledUserId = canceledSub.metadata.userId;

        if (canceledUserId) {
          await prisma.user.update({
            where: { id: canceledUserId },
            data: { plan: 'FREE' }
          });
        }
        break;
    }

    return { received: true };
  });
}
