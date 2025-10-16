#!/bin/bash

# Train Ticketing AI - Quick Setup Script
# This script helps you get started quickly

set -e

echo "🚄 Italian Train Ticketing System - Setup"
echo "=========================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

echo "✅ Docker and Docker Compose are installed"
echo ""

# Ask which AI provider to use
echo "Which AI provider would you like to use?"
echo "1) Ollama (Local, Free) - Recommended for demos"
echo "2) OpenAI (Requires API key)"
echo "3) Anthropic Claude (Requires API key)"
echo ""
read -p "Enter your choice (1-3): " ai_choice

case $ai_choice in
    1)
        echo "Using Ollama (local AI)"
        export AI_PROVIDER="ollama"
        ;;
    2)
        echo "Using OpenAI"
        read -p "Enter your OpenAI API key: " openai_key
        export AI_PROVIDER="openai"
        export OPENAI_API_KEY="$openai_key"
        echo "OPENAI_API_KEY=$openai_key" > backend/.env
        ;;
    3)
        echo "Using Anthropic Claude"
        read -p "Enter your Anthropic API key: " anthropic_key
        export AI_PROVIDER="anthropic"
        export ANTHROPIC_API_KEY="$anthropic_key"
        echo "ANTHROPIC_API_KEY=$anthropic_key" > backend/.env
        ;;
    *)
        echo "Invalid choice. Defaulting to Ollama."
        export AI_PROVIDER="ollama"
        ;;
esac

echo ""
echo "🚀 Starting services..."
docker compose up -d

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 10

# If using Ollama, pull the model
if [ "$AI_PROVIDER" = "ollama" ]; then
    echo ""
    echo "📥 Pulling Ollama model (this may take a few minutes on first run)..."
    docker exec train-ollama ollama pull llama2
fi

echo ""
echo "🔍 Checking service health..."
max_retries=30
retry_count=0

while [ $retry_count -lt $max_retries ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Backend is healthy!"
        break
    fi
    echo "Waiting for backend... ($retry_count/$max_retries)"
    sleep 2
    retry_count=$((retry_count + 1))
done

if [ $retry_count -eq $max_retries ]; then
    echo "❌ Backend failed to start. Check logs with: docker-compose logs backend"
    exit 1
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "🌐 Access the application at: http://localhost:8080"
echo ""
echo "📋 Quick commands:"
echo "  - View logs: docker-compose logs -f"
echo "  - Stop services: docker-compose down"
echo "  - Restart: docker-compose restart"
echo ""
echo "🎯 Demo tips:"
echo "  - Left panel: Traditional form-based booking"
echo "  - Right panel: AI-powered conversational booking"
echo "  - Try: 'I need to go from Milan to Rome tomorrow morning'"
echo ""
echo "Happy presenting! 🚄🤖"
