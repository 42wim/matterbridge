#!/bin/bash
# Bridges-as-a-Service Setup Script
# Run this after configuring .env

set -e

echo "🚀 Bridges-as-a-Service Setup"
echo "=============================="

# Check for .env file
if [ ! -f .env ]; then
    echo "❌ .env file not found. Copy .env.example to .env and configure it first."
    exit 1
fi

# Load environment
source .env

# Generate tokens if not set
generate_token() {
    openssl rand -hex 32
}

echo ""
echo "📝 Step 1: Generating tokens..."
echo ""

# Generate and display tokens for manual configuration
echo "Copy these tokens to your bridge configs:"
echo ""
echo "Slack Bridge (bridges/slack/config.yaml + appservices/slack.yaml):"
echo "  as_token: $(generate_token)"
echo "  hs_token: $(generate_token)"
echo ""
echo "Telegram Bridge (bridges/telegram/config.yaml + appservices/telegram.yaml):"
echo "  as_token: $(generate_token)"
echo "  hs_token: $(generate_token)"
echo ""
echo "Discord Bridge (bridges/discord/config.yaml + appservices/discord.yaml):"
echo "  as_token: $(generate_token)"
echo "  hs_token: $(generate_token)"
echo ""

read -p "Press Enter after updating the configs..."

echo ""
echo "🐘 Step 2: Starting PostgreSQL and Redis..."
docker-compose up -d postgres redis
sleep 10

echo ""
echo "🏠 Step 3: Starting Synapse..."
docker-compose up -d synapse
sleep 15

echo ""
echo "👤 Step 4: Creating admin user..."
echo "Follow the prompts to create an admin account:"
docker exec -it bridges-baas-synapse-1 \
    register_new_matrix_user -a -c /data/homeserver.yaml http://localhost:8008

echo ""
echo "🌉 Step 5: Starting bridges..."
docker-compose up -d mautrix-slack mautrix-telegram mautrix-discord
sleep 10

echo ""
echo "🔧 Step 6: Building and starting API..."
docker-compose up -d --build api
sleep 5

echo ""
echo "🔄 Step 7: Running database migrations..."
docker exec bridges-baas-api-1 npx prisma migrate deploy

echo ""
echo "🌐 Step 8: Starting Nginx..."
docker-compose up -d nginx

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Login to Element as your admin user"
echo "2. Go to Settings → Help & About → Access Token"
echo "3. Add this token to .env as SYNAPSE_ADMIN_TOKEN"
echo "4. Run: docker-compose restart api"
echo ""
echo "Your services:"
echo "  Matrix:  https://matrix.${BASE_DOMAIN}"
echo "  API:     https://api.${BASE_DOMAIN}"
echo ""
echo "Test the API:"
echo "  curl https://api.${BASE_DOMAIN}/health"
