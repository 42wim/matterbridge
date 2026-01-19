// Synapse Admin API Service
// Manages Matrix users and server administration

interface SynapseConfig {
  url: string;
  adminToken: string;
  serverName: string;
  registrationSecret?: string;
}

interface CreateUserOptions {
  username: string;
  password: string;
  displayName?: string;
  admin?: boolean;
  email?: string;
}

interface MatrixUser {
  userId: string;
  accessToken: string;
  deviceId: string;
}

export class SynapseService {
  private config: SynapseConfig;

  constructor(config: SynapseConfig) {
    this.config = config;
  }

  // Create a new Matrix user via Admin API
  async createUser(options: CreateUserOptions): Promise<MatrixUser> {
    const userId = `@${options.username}:${this.config.serverName}`;

    // Create user via Synapse Admin API
    const createResponse = await fetch(
      `${this.config.url}/_synapse/admin/v2/users/${encodeURIComponent(userId)}`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${this.config.adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          password: options.password,
          displayname: options.displayName || options.username,
          admin: options.admin || false,
          deactivated: false,
          threepids: options.email
            ? [{ medium: 'email', address: options.email }]
            : []
        })
      }
    );

    if (!createResponse.ok) {
      const error = await createResponse.text();
      throw new Error(`Failed to create Matrix user: ${error}`);
    }

    // Login to get access token
    const loginResponse = await fetch(
      `${this.config.url}/_matrix/client/v3/login`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'm.login.password',
          identifier: {
            type: 'm.id.user',
            user: options.username
          },
          password: options.password,
          initial_device_display_name: 'BaaS API'
        })
      }
    );

    if (!loginResponse.ok) {
      throw new Error('Failed to login as new user');
    }

    const loginData = await loginResponse.json();

    return {
      userId,
      accessToken: loginData.access_token,
      deviceId: loginData.device_id
    };
  }

  // Get user info
  async getUser(userId: string): Promise<any> {
    const response = await fetch(
      `${this.config.url}/_synapse/admin/v2/users/${encodeURIComponent(userId)}`,
      {
        headers: { 'Authorization': `Bearer ${this.config.adminToken}` }
      }
    );

    if (!response.ok) {
      if (response.status === 404) return null;
      throw new Error('Failed to get user info');
    }

    return response.json();
  }

  // Deactivate user
  async deactivateUser(userId: string): Promise<void> {
    const response = await fetch(
      `${this.config.url}/_synapse/admin/v1/deactivate/${encodeURIComponent(userId)}`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.config.adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ erase: false })
      }
    );

    if (!response.ok) {
      throw new Error('Failed to deactivate user');
    }
  }

  // Get user's joined rooms
  async getUserRooms(userId: string): Promise<string[]> {
    const response = await fetch(
      `${this.config.url}/_synapse/admin/v1/users/${encodeURIComponent(userId)}/joined_rooms`,
      {
        headers: { 'Authorization': `Bearer ${this.config.adminToken}` }
      }
    );

    if (!response.ok) {
      throw new Error('Failed to get user rooms');
    }

    const data = await response.json();
    return data.joined_rooms;
  }

  // Create a DM room between user and bridge bot
  async createBridgeDM(userId: string, bridgeBot: string): Promise<string> {
    // Use the user's token to create room
    // This ensures the user is the creator
    const response = await fetch(
      `${this.config.url}/_matrix/client/v3/createRoom`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.config.adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          preset: 'trusted_private_chat',
          is_direct: true,
          invite: [bridgeBot],
          initial_state: [
            {
              type: 'm.room.guest_access',
              state_key: '',
              content: { guest_access: 'forbidden' }
            }
          ]
        })
      }
    );

    if (!response.ok) {
      throw new Error('Failed to create bridge DM room');
    }

    const data = await response.json();
    return data.room_id;
  }

  // Send a message as the admin bot (for bridge commands)
  async sendMessage(roomId: string, message: string, accessToken: string): Promise<void> {
    const txnId = `baas_${Date.now()}_${Math.random().toString(36).slice(2)}`;

    const response = await fetch(
      `${this.config.url}/_matrix/client/v3/rooms/${encodeURIComponent(roomId)}/send/m.room.message/${txnId}`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          msgtype: 'm.text',
          body: message
        })
      }
    );

    if (!response.ok) {
      throw new Error('Failed to send message');
    }
  }

  // Reset user password
  async resetPassword(userId: string, newPassword: string): Promise<void> {
    const response = await fetch(
      `${this.config.url}/_synapse/admin/v1/reset_password/${encodeURIComponent(userId)}`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.config.adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          new_password: newPassword,
          logout_devices: true
        })
      }
    );

    if (!response.ok) {
      throw new Error('Failed to reset password');
    }
  }

  // Get server stats
  async getServerStats(): Promise<any> {
    const [users, rooms] = await Promise.all([
      fetch(`${this.config.url}/_synapse/admin/v2/users?limit=1`, {
        headers: { 'Authorization': `Bearer ${this.config.adminToken}` }
      }).then(r => r.json()),
      fetch(`${this.config.url}/_synapse/admin/v1/rooms?limit=1`, {
        headers: { 'Authorization': `Bearer ${this.config.adminToken}` }
      }).then(r => r.json())
    ]);

    return {
      totalUsers: users.total,
      totalRooms: rooms.total_rooms
    };
  }
}

// Singleton instance
export const synapse = new SynapseService({
  url: process.env.SYNAPSE_URL || 'http://synapse:8008',
  adminToken: process.env.SYNAPSE_ADMIN_TOKEN || '',
  serverName: process.env.MATRIX_SERVER_NAME || 'localhost'
});
