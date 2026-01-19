// Per-User Matrix Server Management Routes
// Premium feature: Deploy dedicated Matrix servers for users

import { FastifyInstance } from 'fastify';
import { PrismaClient, ServerStatus } from '@prisma/client';
import { z } from 'zod';
import { randomBytes } from 'crypto';

const prisma = new PrismaClient();

// Docker/Kubernetes client would be initialized here
// import * as k8s from '@kubernetes/client-node';
// const kc = new k8s.KubeConfig();
// kc.loadFromDefault();

interface ServerConfig {
  serverName: string;
  userId: string;
  adminToken: string;
  registrationSecret: string;
}

export async function serverRoutes(app: FastifyInstance) {
  // Check if user can have their own server
  app.get('/eligibility', {
    preHandler: [(app as any).authenticate]
  }, async (request: any) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user) throw new Error('User not found');

    const eligible = ['PRO', 'ENTERPRISE'].includes(user.plan);

    return {
      eligible,
      currentPlan: user.plan,
      hasServer: !!user.matrixServer,
      serverStatus: user.matrixServer?.status,
      requiredPlan: 'PRO'
    };
  });

  // Request a dedicated Matrix server
  app.post('/provision', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const schema = z.object({
      subdomain: z.string()
        .min(3)
        .max(20)
        .regex(/^[a-z0-9-]+$/, 'Only lowercase letters, numbers, and hyphens')
    });

    const { subdomain } = schema.parse(request.body);

    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user) {
      return reply.status(404).send({ error: 'User not found' });
    }

    // Check plan eligibility
    if (!['PRO', 'ENTERPRISE'].includes(user.plan)) {
      return reply.status(403).send({
        error: 'Dedicated servers require PRO or ENTERPRISE plan',
        currentPlan: user.plan
      });
    }

    // Check if user already has a server
    if (user.matrixServer) {
      return reply.status(400).send({
        error: 'You already have a dedicated server',
        serverName: user.matrixServer.serverName,
        status: user.matrixServer.status
      });
    }

    // Check if subdomain is available
    const serverName = `${subdomain}.${process.env.BASE_DOMAIN || 'matrix.example.com'}`;
    const existing = await prisma.userMatrixServer.findUnique({
      where: { serverName }
    });

    if (existing) {
      return reply.status(400).send({ error: 'Subdomain already taken' });
    }

    // Generate credentials
    const adminToken = randomBytes(32).toString('hex');
    const registrationSecret = randomBytes(32).toString('hex');

    // Create server record
    const server = await prisma.userMatrixServer.create({
      data: {
        userId: user.id,
        serverName,
        status: 'PROVISIONING',
        adminToken, // Should be encrypted in production
        registrationSecret
      }
    });

    // Start async provisioning
    provisionServer({
      serverName,
      userId: user.id,
      adminToken,
      registrationSecret
    }).catch(err => {
      console.error('Server provisioning failed:', err);
      prisma.userMatrixServer.update({
        where: { id: server.id },
        data: { status: 'ERROR' }
      });
    });

    return {
      serverId: server.id,
      serverName,
      status: 'PROVISIONING',
      message: 'Your Matrix server is being provisioned. This usually takes 2-5 minutes.',
      estimatedTime: '2-5 minutes'
    };
  });

  // Get server status
  app.get('/status', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user?.matrixServer) {
      return reply.status(404).send({ error: 'No dedicated server found' });
    }

    const server = user.matrixServer;

    return {
      serverName: server.serverName,
      status: server.status,
      externalUrl: server.externalUrl,
      resources: {
        cpu: server.cpuLimit,
        memory: server.memoryLimit,
        storage: server.storageLimit
      },
      createdAt: server.createdAt
    };
  });

  // Stop server (to save resources)
  app.post('/stop', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user?.matrixServer) {
      return reply.status(404).send({ error: 'No dedicated server found' });
    }

    // Update status
    await prisma.userMatrixServer.update({
      where: { id: user.matrixServer.id },
      data: { status: 'STOPPING' }
    });

    // Stop the container/pod
    await stopServer(user.matrixServer.serverName);

    await prisma.userMatrixServer.update({
      where: { id: user.matrixServer.id },
      data: { status: 'STOPPED' }
    });

    return { success: true, status: 'STOPPED' };
  });

  // Start server
  app.post('/start', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user?.matrixServer) {
      return reply.status(404).send({ error: 'No dedicated server found' });
    }

    if (user.matrixServer.status !== 'STOPPED') {
      return reply.status(400).send({ error: 'Server is not stopped' });
    }

    // Update status
    await prisma.userMatrixServer.update({
      where: { id: user.matrixServer.id },
      data: { status: 'STARTING' }
    });

    // Start the container/pod
    await startServer(user.matrixServer.serverName);

    await prisma.userMatrixServer.update({
      where: { id: user.matrixServer.id },
      data: { status: 'RUNNING' }
    });

    return {
      success: true,
      status: 'RUNNING',
      externalUrl: user.matrixServer.externalUrl
    };
  });

  // Delete server
  app.delete('/', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user?.matrixServer) {
      return reply.status(404).send({ error: 'No dedicated server found' });
    }

    // Delete the infrastructure
    await deleteServer(user.matrixServer.serverName);

    // Delete the record
    await prisma.userMatrixServer.delete({
      where: { id: user.matrixServer.id }
    });

    return { success: true, message: 'Server deleted' };
  });

  // Get server admin credentials
  app.get('/credentials', {
    preHandler: [(app as any).authenticate]
  }, async (request: any, reply) => {
    const user = await prisma.user.findUnique({
      where: { id: request.user.sub },
      include: { matrixServer: true }
    });

    if (!user?.matrixServer) {
      return reply.status(404).send({ error: 'No dedicated server found' });
    }

    if (user.matrixServer.status !== 'RUNNING') {
      return reply.status(400).send({ error: 'Server is not running' });
    }

    return {
      serverName: user.matrixServer.serverName,
      externalUrl: user.matrixServer.externalUrl,
      adminUserId: `@admin:${user.matrixServer.serverName}`,
      // Only return these once, user should store them securely
      adminToken: user.matrixServer.adminToken,
      registrationSecret: user.matrixServer.registrationSecret
    };
  });
}

// Server provisioning functions (implement based on your infrastructure)

async function provisionServer(config: ServerConfig): Promise<void> {
  console.log(`Provisioning server: ${config.serverName}`);

  // Option 1: Docker Compose per user
  // await provisionWithDocker(config);

  // Option 2: Kubernetes
  // await provisionWithK8s(config);

  // For now, simulate provisioning
  await new Promise(resolve => setTimeout(resolve, 5000));

  // Update status to RUNNING
  await prisma.userMatrixServer.update({
    where: { serverName: config.serverName },
    data: {
      status: 'RUNNING',
      internalUrl: `http://synapse-${config.serverName.replace(/\./g, '-')}:8008`,
      externalUrl: `https://${config.serverName}`
    }
  });

  console.log(`Server provisioned: ${config.serverName}`);
}

async function provisionWithDocker(config: ServerConfig): Promise<void> {
  const { exec } = await import('child_process');
  const { promisify } = await import('util');
  const execAsync = promisify(exec);

  // Generate docker-compose for this user
  const composeFile = generateUserDockerCompose(config);

  // Write compose file
  const fs = await import('fs/promises');
  const userDir = `/opt/baas/users/${config.serverName}`;
  await fs.mkdir(userDir, { recursive: true });
  await fs.writeFile(`${userDir}/docker-compose.yml`, composeFile);

  // Generate Synapse config
  const synapseConfig = generateUserSynapseConfig(config);
  await fs.writeFile(`${userDir}/homeserver.yaml`, synapseConfig);

  // Start the stack
  await execAsync(`docker-compose -f ${userDir}/docker-compose.yml up -d`);
}

function generateUserDockerCompose(config: ServerConfig): string {
  return `
version: "3.8"
services:
  synapse-${config.serverName.replace(/\./g, '-')}:
    image: matrixdotorg/synapse:latest
    restart: unless-stopped
    environment:
      SYNAPSE_SERVER_NAME: ${config.serverName}
      SYNAPSE_CONFIG_PATH: /data/homeserver.yaml
    volumes:
      - ./homeserver.yaml:/data/homeserver.yaml:ro
      - synapse_data:/data
    networks:
      - baas-network
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.${config.serverName}.rule=Host(\`${config.serverName}\`)"
      - "traefik.http.routers.${config.serverName}.tls=true"

volumes:
  synapse_data:

networks:
  baas-network:
    external: true
`;
}

function generateUserSynapseConfig(config: ServerConfig): string {
  return `
server_name: "${config.serverName}"
pid_file: /data/homeserver.pid
public_baseurl: "https://${config.serverName}/"

listeners:
  - port: 8008
    tls: false
    type: http
    x_forwarded: true
    bind_addresses: ['0.0.0.0']
    resources:
      - names: [client, federation]
        compress: false

database:
  name: sqlite3
  args:
    database: /data/homeserver.db

log_config: "/data/log.config"
media_store_path: /data/media_store
registration_shared_secret: "${config.registrationSecret}"
enable_registration: false
report_stats: false
signing_key_path: "/data/signing.key"

# Connect to main BaaS bridges
app_service_config_files: []
`;
}

async function stopServer(serverName: string): Promise<void> {
  console.log(`Stopping server: ${serverName}`);
  // Docker: docker-compose -f /opt/baas/users/${serverName}/docker-compose.yml stop
  // K8s: kubectl scale deployment synapse-${serverName} --replicas=0
}

async function startServer(serverName: string): Promise<void> {
  console.log(`Starting server: ${serverName}`);
  // Docker: docker-compose -f /opt/baas/users/${serverName}/docker-compose.yml start
  // K8s: kubectl scale deployment synapse-${serverName} --replicas=1
}

async function deleteServer(serverName: string): Promise<void> {
  console.log(`Deleting server: ${serverName}`);
  // Docker: docker-compose -f /opt/baas/users/${serverName}/docker-compose.yml down -v
  // K8s: kubectl delete deployment,service,pvc synapse-${serverName}
}
