# Italian Train Ticketing System - AI Demo

A complete train ticket booking system built with Go backend, PostgreSQL database, and AI integration. This project demonstrates the comparison between traditional form-based booking and AI-powered conversational booking.

**Perfect for Linux Day presentations and AI demonstrations!**

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                            │
│  ┌──────────────────┐         ┌──────────────────┐        │
│  │  Traditional     │         │   AI Chat        │        │
│  │  Booking Form    │         │   Interface      │        │
│  └──────────────────┘         └──────────────────┘        │
└────────────────────┬─────────────────┬────────────────────┘
                     │                 │
                     ▼                 ▼
          ┌──────────────────────────────────────┐
          │         Go Backend (Gin)              │
          │  - REST API                           │
          │  - Train Search Service               │
          │  - Booking Service                    │
          │  - AI Integration Service             │
          └──────────┬───────────────┬────────────┘
                     │               │
                     ▼               ▼
          ┌─────────────────┐  ┌──────────────┐
          │   PostgreSQL    │  │  AI Provider │
          │   - Stations    │  │  - OpenAI    │
          │   - Trains      │  │  - Anthropic │
          │   - Schedules   │  │  - Ollama    │
          │   - Bookings    │  └──────────────┘
          └─────────────────┘
```

## Features

### Traditional Booking
- Station selection with dropdowns
- Date picker and time preferences
- Passenger count (adults, seniors, children)
- Filter by WiFi and food service
- Train search with detailed results
- Direct booking with passenger information

### AI-Powered Booking
- Natural language conversation
- Understands various date formats ("tomorrow", "next Monday")
- Smart station name matching ("Milan" → "Milano Centrale")
- Contextual booking flow
- Function calling for backend operations
- Conversational booking confirmations

### Backend Features
- RESTful API with proper error handling
- PostgreSQL database with proper indexes
- Transaction-based booking (ACID compliant)
- Seat management and availability tracking
- Passenger type discounts (adult, senior, child, infant)
- Booking reference generation (TRN-YYYY-NNNNN format)
- Fuzzy station name matching using pg_trgm
- Connection pooling and health checks
- Structured logging

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 15+ (for local development)
- (Optional) OpenAI API key or Anthropic API key

## Quick Start

### 1. Clone and Navigate

```bash
cd train-ticketing-ai
```

### 2. Choose AI Provider

#### Option A: Using Ollama (Local, Free)

```bash
# Start services
docker compose up -d

# Pull AI model (first time only)
docker exec -it train-ollama ollama pull llama2

# Wait for services to be ready (~30 seconds)
```

#### Option B: Using OpenAI

```bash
# Set API key
echo "OPENAI_API_KEY=sk-your-key-here" > backend/.env

# Update docker-compose.yml to use OpenAI
# Uncomment the OpenAI env vars and comment out Ollama

# Start services
docker compose up -d
```

#### Option C: Using Anthropic Claude

```bash
# Set API key
echo "ANTHROPIC_API_KEY=sk-ant-your-key-here" > backend/.env

# Update docker-compose.yml to use Anthropic
# Uncomment the Anthropic env vars and comment out Ollama

# Start services
docker compose up -d
```

### 3. Access the Application

Open your browser to: **http://localhost:8080**

### 4. Check Health

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "time": "2025-10-16T10:30:00Z"
}
```

## Demo Script for Presentations

### Introduction (2 minutes)

"Today I'll demonstrate a modern train booking system that showcases the difference between traditional web forms and AI-powered interfaces. The system is built with Go, PostgreSQL, and integrates with multiple AI providers."

### Traditional Booking Demo (3 minutes)

1. **Navigate to the left panel (Traditional Booking)**
2. **Show the form:**
   - Select "Milano Centrale" as origin
   - Select "Roma Termini" as destination
   - Choose tomorrow's date
   - Select "Morning" time preference
   - Set 2 adults
   - Check "WiFi Required"
3. **Click "Search Trains"**
4. **Explain the results:**
   - Train types (Frecciarossa, Intercity, Regionale)
   - Pricing based on distance and train type
   - Amenities and available seats
5. **Click "Book Now" on a train**
6. **Fill in passenger names and confirm**
7. **Show the booking confirmation with reference number**

### AI-Powered Booking Demo (5 minutes)

1. **Navigate to the right panel (AI Assistant)**
2. **Example 1 - Simple Query:**
   ```
   "I need to go from Milan to Rome tomorrow morning"
   ```
   - AI understands "Milan" = "Milano Centrale"
   - AI interprets "tomorrow morning"
   - Shows train options

3. **Example 2 - Natural Language:**
   ```
   "Show me the cheapest option with WiFi"
   ```
   - AI filters and recommends

4. **Example 3 - Booking:**
   ```
   "Book 2 adult tickets on the 7:30 train"
   ```
   - AI asks for passenger names
   - Completes booking
   - Shows confirmation with reference

5. **Example 4 - Lookup:**
   ```
   "What's the status of booking TRN-2025-00001?"
   ```
   - AI retrieves booking details

### Key Points to Emphasize

- **Traditional:** Requires specific field navigation, precise inputs
- **AI-Powered:** Natural language, contextual understanding, conversational flow
- **Both:** Use the same backend API and database
- **Real-time:** All data is live, bookings are actually created
- **Extensible:** Easy to add new features or integrate with other systems

## Database Schema

### Stations
- Italian railway stations with codes
- Fuzzy search enabled (pg_trgm extension)
- 8 major stations seeded

### Trains
- 20+ trains across three types
- Frecciarossa: High-speed, €0.15/km, WiFi + Food
- Intercity: Medium-speed, €0.10/km, WiFi only
- Regionale: Local, €0.06/km, basic service

### Schedules
- 500+ train schedules
- Covers weekdays (Monday-Friday)
- Popular routes: Milano-Roma, Milano-Venezia, Roma-Napoli
- Multiple daily departures

### Bookings
- ACID-compliant transaction handling
- Automatic seat assignment
- Passenger type discounts:
  - Adult: 100% of base price
  - Senior (65+): 80% (20% discount)
  - Child (4-17): 70% (30% discount)
  - Infant (0-3): Free, no seat

## API Endpoints

### Stations
```
GET /api/stations
```
Returns all available stations.

### Search Trains
```
POST /api/search
Content-Type: application/json

{
  "origin": "MI",
  "destination": "RM",
  "date": "2025-10-17",
  "time_preference": "morning",
  "passenger_count": 2,
  "filters": {
    "has_wifi": true,
    "has_food": false
  }
}
```

### Create Booking
```
POST /api/bookings
Content-Type: application/json

{
  "schedule_id": 123,
  "date": "2025-10-17",
  "passengers": [
    {
      "name": "John Doe",
      "passenger_type": "adult"
    },
    {
      "name": "Jane Doe",
      "passenger_type": "senior"
    }
  ]
}
```

### Get Booking
```
GET /api/bookings/:ref
```

### Cancel Booking
```
DELETE /api/bookings/:ref
```

### AI Chat
```
POST /api/ai/chat
Content-Type: application/json

{
  "session_id": "uuid-v4",
  "message": "I need to go from Milan to Rome tomorrow"
}
```

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=trainpass123
DB_NAME=traintickets

# AI Provider (choose one)
AI_PROVIDER=ollama       # Options: openai, anthropic, ollama

# Ollama (local)
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=llama2

# OpenAI
OPENAI_API_KEY=sk-...

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Server
SERVER_PORT=8080
```

## Development

### Run Locally (Without Docker)

1. **Start PostgreSQL:**
```bash
docker run -d \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=trainpass123 \
  -e POSTGRES_DB=traintickets \
  postgres:15-alpine
```

2. **Apply database schema:**
```bash
psql -h localhost -U postgres -d traintickets -f init-db.sql
```

3. **Run backend:**
```bash
cd backend
cp .env.example .env
# Edit .env with your settings
go mod download
go run main.go
```

4. **Frontend is served by backend** at http://localhost:8080

### Project Structure

```
train-ticketing-ai/
├── backend/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   │   ├── train_service.go     # Train search logic
│   │   ├── booking_service.go   # Booking management
│   │   └── ai_service.go        # AI integration
│   ├── main.go          # Application entry point
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── css/
│   │   └── style.css    # Complete styling
│   ├── js/
│   │   ├── app.js       # Utility functions
│   │   ├── traditional.js  # Traditional booking
│   │   └── ai_chat.js   # AI chat interface
│   └── index.html       # Main page
├── docker-compose.yml
├── init-db.sql          # Database schema and seed data
└── README.md
```

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep train-db

# Check logs
docker logs train-db

# Restart database
docker compose restart postgres
```

### Backend Not Starting

```bash
# Check backend logs
docker logs train-backend

# Rebuild backend
docker compose up -d --build backend
```

### Ollama Not Responding

```bash
# Check Ollama status
docker logs train-ollama

# Pull model again
docker exec -it train-ollama ollama pull llama2

# Restart Ollama
docker compose restart ollama
```

### No Trains Found

- Check that the date is not in the past
- Verify stations exist (use /api/stations)
- Check day of week (weekends have fewer trains)
- Try broader search criteria

## Testing Checklist

- [ ] Health endpoint returns 200
- [ ] Stations load in dropdowns
- [ ] Traditional search finds trains
- [ ] Traditional booking creates booking
- [ ] AI chat responds to messages
- [ ] AI chat can search trains
- [ ] AI chat can create bookings
- [ ] AI chat can retrieve bookings
- [ ] Booking reference is unique
- [ ] Seats are properly decremented
- [ ] Cancellation restores seats
- [ ] Discount pricing works correctly

## Performance Considerations

- Database connection pool: 25 max connections
- AI API timeout: 30 seconds
- Conversation history: Last 20 messages
- Frontend auto-scroll for chat
- Prepared statements for all queries
- Indexes on foreign keys and search fields

## Security Notes

This is a demo application. For production use:

- [ ] Add authentication and authorization
- [ ] Implement rate limiting
- [ ] Use HTTPS/TLS
- [ ] Sanitize all user inputs
- [ ] Add CSRF protection
- [ ] Implement proper session management
- [ ] Use secure environment variable management
- [ ] Add audit logging
- [ ] Implement PCI DSS for payments
- [ ] Add data encryption at rest

## License

MIT License - Free for educational and demonstration purposes.

## Credits

Built for Linux Day demonstrations.

- Backend: Go + Gin
- Database: PostgreSQL
- Frontend: Vanilla JavaScript
- AI: OpenAI GPT-4 / Anthropic Claude / Ollama

## Contributing

This is a demo project for presentations. Feel free to fork and modify for your own demonstrations!

## Support

For issues or questions:
1. Check the Troubleshooting section
2. Review docker-compose logs
3. Ensure all services are healthy

---

**Happy coding and great presentations!** 🚄🤖
