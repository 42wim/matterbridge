// Bridges-as-a-Service API
// Main entry point

import Fastify from 'fastify';
import cors from '@fastify/cors';
import jwt from '@fastify/jwt';
import cookie from '@fastify/cookie';

import { authRoutes } from './routes/auth.js';
import { userRoutes } from './routes/users.js';
import { bridgeRoutes } from './routes/bridges.js';
import { serverRoutes } from './routes/servers.js';
import { adminRoutes } from './routes/admin.js';
import { webhookRoutes } from './routes/webhooks.js';

const app = Fastify({
  logger: {
    level: process.env.LOG_LEVEL || 'info',
    transport: process.env.NODE_ENV !== 'production'
      ? { target: 'pino-pretty' }
      : undefined
  }
});

// Plugins
await app.register(cors, {
  origin: process.env.CORS_ORIGIN || '*',
  credentials: true
});

await app.register(jwt, {
  secret: process.env.JWT_SECRET || 'change-me-in-production'
});

await app.register(cookie);

// Auth decorator
app.decorate('authenticate', async (request: any, reply: any) => {
  try {
    await request.jwtVerify();
  } catch (err) {
    reply.status(401).send({ error: 'Unauthorized' });
  }
});

// Routes
await app.register(authRoutes, { prefix: '/auth' });
await app.register(userRoutes, { prefix: '/users' });
await app.register(bridgeRoutes, { prefix: '/bridges' });
await app.register(serverRoutes, { prefix: '/servers' });
await app.register(adminRoutes, { prefix: '/admin' });
await app.register(webhookRoutes, { prefix: '/webhooks' });

// Health check
app.get('/health', async () => ({ status: 'ok', timestamp: new Date().toISOString() }));

// Start server
const start = async () => {
  try {
    const port = parseInt(process.env.PORT || '3000');
    await app.listen({ port, host: '0.0.0.0' });
    console.log(`🚀 BaaS API running on port ${port}`);
  } catch (err) {
    app.log.error(err);
    process.exit(1);
  }
};

start();

export { app };
