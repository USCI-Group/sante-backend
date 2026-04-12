#!/bin/bash

# Configuration
SERVER="192.168.28.76" # Staging
PORT="4000"
USER="vicknesh"
DEST_DIR="~/sante-backend"

echo "🚀 Starting Santé Backend Deployment to $SERVER..."

# 1. Check if .env exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found! Please create it from .env.sante.template first."
    exit 1
fi

# 2. Sync files to server
echo "📦 Transferring files to server..."
# We exclude node_modules, .git, etc.
rsync -avz -e "ssh -p $PORT" --exclude '.git' --exclude '.encore' --exclude '.output' . $USER@$SERVER:$DEST_DIR

# 3. Remote Build and Run
echo "🏗️ Building and starting containers on server..."
ssh -p $PORT $USER@$SERVER << EOF
    cd $DEST_DIR
    # Build the docker image
    docker build -t sante-backend:latest .
    # Restart containers
    docker-compose down
    docker-compose up -d
    echo "✅ Deployment complete! Checking logs..."
    docker-compose logs --tail=20 backend
EOF
