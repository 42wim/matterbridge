-- Initialize databases for Synapse and all bridges
-- Each component gets its own database for isolation

-- Synapse database (main Matrix server)
CREATE DATABASE synapse;

-- Bridge databases
CREATE DATABASE mautrix_slack;
CREATE DATABASE mautrix_telegram;
CREATE DATABASE mautrix_discord;

-- BaaS API database (user management, billing, etc.)
CREATE DATABASE baas;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE synapse TO matrix;
GRANT ALL PRIVILEGES ON DATABASE mautrix_slack TO matrix;
GRANT ALL PRIVILEGES ON DATABASE mautrix_telegram TO matrix;
GRANT ALL PRIVILEGES ON DATABASE mautrix_discord TO matrix;
GRANT ALL PRIVILEGES ON DATABASE baas TO matrix;
