// Credential Management Service
// Secure storage and retrieval of bridge tokens and secrets

import { createCipheriv, createDecipheriv, randomBytes, scryptSync } from 'crypto';
import Redis from 'ioredis';

interface EncryptedData {
  iv: string;
  data: string;
  tag: string;
}

interface BridgeCredentials {
  bridgeType: string;
  userId: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: Date;
  metadata?: Record<string, any>;
}

export class CredentialService {
  private encryptionKey: Buffer;
  private redis: Redis;
  private algorithm = 'aes-256-gcm';

  constructor() {
    // Derive encryption key from secret
    const secret = process.env.CREDENTIAL_ENCRYPTION_KEY || 'change-me-in-production';
    this.encryptionKey = scryptSync(secret, 'salt', 32);

    // Redis for caching and session storage
    this.redis = new Redis(process.env.REDIS_URL || 'redis://localhost:6379');
  }

  // Encrypt sensitive data
  encrypt(plaintext: string): EncryptedData {
    const iv = randomBytes(16);
    const cipher = createCipheriv(this.algorithm, this.encryptionKey, iv);

    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');

    return {
      iv: iv.toString('hex'),
      data: encrypted,
      tag: (cipher as any).getAuthTag().toString('hex')
    };
  }

  // Decrypt sensitive data
  decrypt(encrypted: EncryptedData): string {
    const decipher = createDecipheriv(
      this.algorithm,
      this.encryptionKey,
      Buffer.from(encrypted.iv, 'hex')
    );
    (decipher as any).setAuthTag(Buffer.from(encrypted.tag, 'hex'));

    let decrypted = decipher.update(encrypted.data, 'hex', 'utf8');
    decrypted += decipher.final('utf8');

    return decrypted;
  }

  // Store bridge credentials securely
  async storeBridgeCredentials(credentials: BridgeCredentials): Promise<void> {
    const key = `bridge:${credentials.userId}:${credentials.bridgeType}`;

    // Encrypt tokens
    const encryptedData: Record<string, any> = {
      bridgeType: credentials.bridgeType,
      userId: credentials.userId,
      metadata: credentials.metadata
    };

    if (credentials.accessToken) {
      encryptedData.accessToken = this.encrypt(credentials.accessToken);
    }
    if (credentials.refreshToken) {
      encryptedData.refreshToken = this.encrypt(credentials.refreshToken);
    }
    if (credentials.expiresAt) {
      encryptedData.expiresAt = credentials.expiresAt.toISOString();
    }

    // Store in Redis with optional TTL
    const ttl = credentials.expiresAt
      ? Math.floor((credentials.expiresAt.getTime() - Date.now()) / 1000)
      : undefined;

    if (ttl && ttl > 0) {
      await this.redis.setex(key, ttl, JSON.stringify(encryptedData));
    } else {
      await this.redis.set(key, JSON.stringify(encryptedData));
    }
  }

  // Retrieve and decrypt bridge credentials
  async getBridgeCredentials(userId: string, bridgeType: string): Promise<BridgeCredentials | null> {
    const key = `bridge:${userId}:${bridgeType}`;
    const data = await this.redis.get(key);

    if (!data) return null;

    const encryptedData = JSON.parse(data);

    const credentials: BridgeCredentials = {
      bridgeType: encryptedData.bridgeType,
      userId: encryptedData.userId,
      metadata: encryptedData.metadata
    };

    if (encryptedData.accessToken) {
      credentials.accessToken = this.decrypt(encryptedData.accessToken);
    }
    if (encryptedData.refreshToken) {
      credentials.refreshToken = this.decrypt(encryptedData.refreshToken);
    }
    if (encryptedData.expiresAt) {
      credentials.expiresAt = new Date(encryptedData.expiresAt);
    }

    return credentials;
  }

  // Delete bridge credentials
  async deleteBridgeCredentials(userId: string, bridgeType: string): Promise<void> {
    const key = `bridge:${userId}:${bridgeType}`;
    await this.redis.del(key);
  }

  // Store Matrix access token
  async storeMatrixToken(userId: string, accessToken: string, deviceId: string): Promise<void> {
    const key = `matrix:${userId}`;
    const encrypted = this.encrypt(JSON.stringify({ accessToken, deviceId }));
    await this.redis.set(key, JSON.stringify(encrypted));
  }

  // Get Matrix access token
  async getMatrixToken(userId: string): Promise<{ accessToken: string; deviceId: string } | null> {
    const key = `matrix:${userId}`;
    const data = await this.redis.get(key);

    if (!data) return null;

    const encrypted = JSON.parse(data);
    const decrypted = this.decrypt(encrypted);
    return JSON.parse(decrypted);
  }

  // Store temporary OAuth state (for Slack OAuth flow)
  async storeOAuthState(state: string, data: Record<string, any>, ttlSeconds = 600): Promise<void> {
    const key = `oauth:${state}`;
    await this.redis.setex(key, ttlSeconds, JSON.stringify(data));
  }

  // Get and delete OAuth state (one-time use)
  async consumeOAuthState(state: string): Promise<Record<string, any> | null> {
    const key = `oauth:${state}`;
    const data = await this.redis.get(key);

    if (!data) return null;

    await this.redis.del(key);
    return JSON.parse(data);
  }

  // Store bridge login session (for multi-step auth like Telegram)
  async storeBridgeLoginSession(
    userId: string,
    bridgeType: string,
    sessionData: Record<string, any>,
    ttlSeconds = 300
  ): Promise<void> {
    const key = `login:${userId}:${bridgeType}`;
    await this.redis.setex(key, ttlSeconds, JSON.stringify(sessionData));
  }

  // Get bridge login session
  async getBridgeLoginSession(userId: string, bridgeType: string): Promise<Record<string, any> | null> {
    const key = `login:${userId}:${bridgeType}`;
    const data = await this.redis.get(key);
    return data ? JSON.parse(data) : null;
  }

  // Cleanup expired sessions
  async cleanup(): Promise<void> {
    // Redis handles TTL automatically, but we can do additional cleanup here
    console.log('Credential cleanup completed');
  }
}

// Singleton instance
export const credentials = new CredentialService();
