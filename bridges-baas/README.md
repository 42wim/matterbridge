# Bridges-as-a-Service (BaaS)

A complete platform for deploying Matrix bridges (Slack, Telegram, Discord) as a service for your users.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Your Users                               │
│                    (Mobile App / Web App)                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Nginx                                   │
│            (SSL termination, rate limiting)                      │
└─────────────────────────────────────────────────────────────────┘
                    │                    │
                    ▼                    ▼
┌─────────────────────────┐    ┌─────────────────────────┐
│      BaaS API           │    │        Synapse          │
│    (TypeScript)         │    │   (Matrix Homeserver)   │
│                         │    │                         │
│  • User management      │    │  • Matrix protocol      │
│  • Auth (JWT)           │    │  • Room management      │
│  • Bridge orchestration │    │  • Federation           │
│  • Billing/plans        │    │  • Appservice host      │
└─────────────────────────┘    └─────────────────────────┘
            │                              │
            └──────────────┬───────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│ mautrix-    │   │ mautrix-    │   │ mautrix-    │
│ slack       │   │ telegram    │   │ discord     │
└─────────────┘   └─────────────┘   └─────────────┘
         │                 │                 │
         ▼                 ▼                 ▼
      Slack            Telegram          Discord
```

## Quick Start

### Prerequisites

- Docker & Docker Compose
- A domain with DNS configured
- SSL certificates (Let's Encrypt recommended)
- Telegram API credentials (https://my.telegram.org/apps)

### 1. Clone and Configure

```bash
cd bridges-baas
cp .env.example .env
```

Edit `.env` with your settings:
- Set your domain name
- Generate secrets: `openssl rand -hex 32`
- Add Telegram API credentials

### 2. Generate Bridge Tokens

```bash
# Generate unique tokens for each bridge
for token in {1..6}; do
  echo "TOKEN_$token: $(openssl rand -hex 32)"
done
```

Update these tokens in:
- `bridges/slack/config.yaml` (as_token, hs_token)
- `bridges/telegram/config.yaml` (as_token, hs_token)
- `bridges/discord/config.yaml` (as_token, hs_token)
- `appservices/*.yaml` (matching tokens)

### 3. Update Domain Names

Replace `YOURDOMAIN.COM` in all config files:
```bash
find . -type f -name "*.yaml" -exec sed -i 's/YOURDOMAIN\.COM/yourdomain.com/g' {} +
find . -type f -name "*.conf" -exec sed -i 's/yourdomain\.com/yourdomain.com/g' {} +
```

### 4. Start Infrastructure

```bash
# Start core services
docker-compose up -d postgres redis

# Wait for postgres to be ready
sleep 10

# Start Synapse (generates signing keys on first run)
docker-compose up -d synapse

# Create admin user
docker exec -it bridges-baas-synapse-1 \
  register_new_matrix_user -a -c /data/homeserver.yaml http://localhost:8008

# Start bridges
docker-compose up -d mautrix-slack mautrix-telegram mautrix-discord

# Start API
docker-compose up -d api nginx
```

### 5. Initialize Database

```bash
# Run Prisma migrations
docker exec -it bridges-baas-api-1 npx prisma migrate deploy
```

### 6. Get Admin Token

Login to Element (or any Matrix client) as your admin user, then:
- Settings → Help & About → Access Token
- Add this to your `.env` as `SYNAPSE_ADMIN_TOKEN`
- Restart the API: `docker-compose restart api`

## API Endpoints

### Authentication

```bash
# Register new user
curl -X POST https://api.yourdomain.com/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "username": "newuser"
  }'

# Login
curl -X POST https://api.yourdomain.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### Bridge Management

```bash
# Get available bridges
curl https://api.yourdomain.com/bridges \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Connect to Slack
curl -X POST https://api.yourdomain.com/bridges/slack/connect \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Check bridge status
curl https://api.yourdomain.com/bridges/slack/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## User Bridge Flow

1. **User registers** via API → Creates Matrix account
2. **User requests bridge** → API returns instructions
3. **User messages bridge bot** in Matrix → `!slack login`
4. **Bridge bot authenticates** → OAuth or token flow
5. **Webhook updates status** → User is connected
6. **Messages flow** → Slack ↔ Matrix ↔ User's app

## Mobile App Integration

```typescript
import { createClient } from 'matrix-js-sdk';

// 1. Register/Login via your API
const { matrix } = await api.post('/auth/login', credentials);

// 2. Connect to Matrix
const client = createClient({
  baseUrl: matrix.homeserver,
  accessToken: matrix.accessToken,
  userId: matrix.userId
});

// 3. Start receiving messages
await client.startClient();

client.on('Room.timeline', (event, room) => {
  console.log(`[${room.name}] ${event.getContent().body}`);
});

// 4. Send message (goes to bridged platform)
await client.sendMessage(roomId, {
  msgtype: 'm.text',
  body: 'Hello from the app!'
});
```

## Bridge Commands

### Slack
```
!slack login          # Start OAuth flow
!slack logout         # Disconnect
!slack ping           # Test connection
```

### Telegram
```
!tg login +1234567890  # Login with phone
!tg code 12345         # Enter verification code
!tg 2fa password       # Enter 2FA password (if enabled)
!tg logout             # Disconnect
```

### Discord
```
!discord login         # Show QR code
!discord login-token   # Login with token
!discord logout        # Disconnect
```

## Monitoring

### Check Bridge Health

```bash
# View logs
docker-compose logs -f mautrix-slack
docker-compose logs -f mautrix-telegram
docker-compose logs -f mautrix-discord

# Check container status
docker-compose ps
```

### Admin Dashboard

Access admin endpoints with admin JWT:
```bash
# Get platform stats
curl https://api.yourdomain.com/admin/stats \
  -H "Authorization: Bearer ADMIN_JWT"

# List all users
curl https://api.yourdomain.com/admin/users \
  -H "Authorization: Bearer ADMIN_JWT"
```

## Scaling

### For 100+ Users
- Move PostgreSQL to managed service (RDS, Cloud SQL)
- Add Redis cluster
- Use Kubernetes for bridge orchestration

### For 1000+ Users
- Deploy Synapse workers
- Add CDN for media
- Consider per-user bridge isolation

## Troubleshooting

### Bridge won't start
```bash
# Check logs
docker-compose logs mautrix-slack

# Common issues:
# - Tokens don't match between bridge config and appservice registration
# - Database connection failed
# - Synapse not reachable
```

### Users can't connect
```bash
# Verify appservice is registered
docker exec -it bridges-baas-synapse-1 cat /data/homeserver.yaml | grep app_service

# Check bridge is responding
docker exec -it bridges-baas-mautrix-slack-1 wget -qO- http://localhost:29335/_matrix/mau/live
```

### Messages not bridging
```bash
# Check bridge database
docker exec -it bridges-baas-postgres-1 psql -U matrix -d mautrix_slack -c "SELECT * FROM portal LIMIT 5;"

# Verify user is logged in
# Send `!slack ping` to the bridge bot
```

## Security Considerations

1. **Rotate tokens regularly** - Bridge tokens, JWT secrets
2. **Encrypt credentials at rest** - The API encrypts stored tokens
3. **Use separate databases** - Each bridge has its own database
4. **Rate limit API** - Nginx config includes rate limiting
5. **Audit logging** - All admin actions are logged

## License

MIT
