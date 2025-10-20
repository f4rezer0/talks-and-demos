# AI-Powered Database Interactions
## Building Intelligent Applications with Local Database Integration

---

## About This Talk

**Focus:** How to integrate AI/LLM capabilities with your local database

**What You'll Learn:**
- Architecture patterns for AI-database integration
- Function calling and tool use with LLMs
- Real-world implementation with Go + PostgreSQL + OpenAI
- Best practices and gotchas

---

## The Demo: Frecciarossa Train Booking System

**Two Booking Methods:**
1. **Traditional**: Forms, dropdowns, manual input
2. **AI Assistant**: Natural language → Database queries

**Tech Stack:**
- Backend: Go (Gin framework)
- Database: PostgreSQL
- AI: OpenAI-compatible API (RelaxAI)
- Frontend: Vanilla JavaScript

---

## The Problem Statement

**Traditional Apps:**
```
User → Form → Validation → SQL Query → Response
```

**AI-Enhanced Apps:**
```
User → Natural Language → AI → Function Call → SQL Query → Response
```

**Challenge:** How do we bridge the gap between conversational AI and structured databases?

---

## Architecture Overview

```
┌─────────────┐
│   User      │ "I need to go from Milan to Rome tomorrow"
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│   AI Service        │ Understands intent
│   (OpenAI API)      │ Extracts: origin, destination, date
└──────┬──────────────┘
       │
       ▼ Function Call
┌─────────────────────┐
│   Backend Logic     │ search_trains(origin, dest, date)
│   (Go + Gin)        │
└──────┬──────────────┘
       │
       ▼ SQL Query
┌─────────────────────┐
│   PostgreSQL DB     │ SELECT * FROM schedules...
└─────────────────────┘
```

---

## Key Concept: Function Calling (Tools)

**What is Function Calling?**

Modern LLMs can:
1. Understand when to call external functions
2. Extract parameters from natural language
3. Format function calls as structured JSON

**Example:**
```
User: "Show me trains from Milan to Rome tomorrow morning"

AI Decision: Call search_trains()
Parameters: {
  "origin": "Milano Centrale",
  "destination": "Roma Termini",
  "date": "2025-10-18",
  "time_preference": "morning"
}
```

---

## Defining Functions for the AI

**Step 1: Define Function Schema**

```go
func getFunctionDefinitions() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "type": "function",
            "function": map[string]interface{}{
                "name": "search_trains",
                "description": "Search for available train schedules",
                "parameters": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "origin": {
                            "type": "string",
                            "description": "Departure station name or code",
                        },
                        "destination": {
                            "type": "string",
                            "description": "Arrival station name or code",
                        },
                        "date": {
                            "type": "string",
                            "description": "Travel date in YYYY-MM-DD format",
                        },
                        // ... more parameters
                    },
                    "required": ["origin", "destination", "date"],
                },
            },
        },
    }
}
```

---

## The AI Service Flow

**1. User sends message to AI endpoint**
```go
POST /api/ai/chat
{
  "session_id": "uuid-here",
  "message": "I need to go from Milan to Rome tomorrow"
}
```

**2. Backend forwards to OpenAI API**
```go
messages := []Message{
    {Role: "system", Content: systemPrompt},
    {Role: "user", Content: userMessage},
}

reqBody := map[string]interface{}{
    "model":       "gpt-4",
    "messages":    messages,
    "tools":       getFunctionDefinitions(),
    "tool_choice": "auto",
}
```

---

## AI Response with Function Call

**OpenAI Returns:**
```json
{
  "choices": [{
    "message": {
      "tool_calls": [{
        "id": "call_abc123",
        "type": "function",
        "function": {
          "name": "search_trains",
          "arguments": "{\"origin\":\"Milano Centrale\",\"destination\":\"Roma Termini\",\"date\":\"2025-10-18\",\"passenger_count\":1}"
        }
      }]
    }
  }]
}
```

**Now we need to execute this function!**

---

## Executing the Function Call

```go
func executeFunction(name string, args map[string]interface{}) (interface{}, error) {
    switch name {
    case "search_trains":
        return executeSearchTrains(args)
    case "create_booking":
        return executeCreateBooking(args)
    case "get_booking_details":
        return executeGetBookingDetails(args)
    default:
        return nil, fmt.Errorf("unknown function: %s", name)
    }
}
```

**This bridges AI → Database!**

---

## Database Query Implementation

```go
func executeSearchTrains(args map[string]interface{}) (interface{}, error) {
    // Extract parameters
    origin := getString(args, "origin")
    destination := getString(args, "destination")
    date := getString(args, "date")

    // Build search request
    searchReq := models.SearchRequest{
        Origin:         mapStationCode(origin),
        Destination:    mapStationCode(destination),
        Date:           date,
        PassengerCount: getInt(args, "passenger_count"),
    }

    // Execute database query
    results, err := SearchTrains(searchReq)
    if err != nil {
        return nil, err
    }

    return results, nil
}
```

---

## The Actual SQL Query

```go
func SearchTrains(req SearchRequest) ([]TrainResult, error) {
    query := `
        SELECT
            s.id, s.departure_time, s.arrival_time,
            s.price_base, s.available_seats,
            t.number, t.type, t.has_wifi, t.has_food,
            o.name as origin_name, d.name as dest_name
        FROM schedules s
        JOIN trains t ON s.train_id = t.id
        JOIN stations o ON s.origin_id = o.id
        JOIN stations d ON s.destination_id = d.id
        WHERE o.code = $1
          AND d.code = $2
          AND s.day_of_week = $3
          AND s.available_seats >= $4
        ORDER BY s.departure_time
    `

    dayOfWeek := calculateDayOfWeek(req.Date)
    rows, err := db.Query(query, req.Origin, req.Destination,
                          dayOfWeek, req.PassengerCount)
    // ... process results
}
```

---

## Completing the Loop: Response to AI

**After executing the function, send results back to AI:**

```go
// Add function result to conversation
messages = append(messages, Message{
    Role:       "tool",
    Content:    jsonResults,
    ToolCallID: toolCallID,
})

// Call AI again with results
response := callOpenAI(messages, tools)

// AI now generates human-readable response:
// "I found 5 trains from Milan to Rome tomorrow morning.
//  The earliest departure is at 6:00 AM..."
```

---

## Session Management

**Why Sessions Matter:**

AI needs context across multiple messages!

```go
type ChatSession struct {
    ID        string
    Messages  []Message
    CreatedAt time.Time
}

var sessions = make(map[string]*ChatSession)

func getOrCreateSession(sessionID string) *ChatSession {
    if session, exists := sessions[sessionID]; exists {
        return session
    }

    session := &ChatSession{
        ID:       sessionID,
        Messages: []Message{systemMessage},
    }
    sessions[sessionID] = session
    return session
}
```

---

## Session Context Example

**Conversation Flow:**

```
User: "Show me trains from Milan to Rome"
AI:   [searches] "I found 5 trains..."

User: "Book 2 tickets on the 7:30 train"  ← Context!
AI:   [knows which trains were shown]
      [books the 7:30 train]
```

**Without sessions, AI has no memory!**

---

## Database Schema Design

**Key Tables:**

```sql
CREATE TABLE stations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    city VARCHAR(100),
    code VARCHAR(10) UNIQUE
);

CREATE TABLE trains (
    id SERIAL PRIMARY KEY,
    number VARCHAR(20),
    type VARCHAR(50),
    has_wifi BOOLEAN,
    has_food BOOLEAN
);

CREATE TABLE schedules (
    id SERIAL PRIMARY KEY,
    train_id INTEGER REFERENCES trains(id),
    origin_id INTEGER REFERENCES stations(id),
    destination_id INTEGER REFERENCES stations(id),
    departure_time TIME,
    arrival_time TIME,
    day_of_week INTEGER,  -- 0=Sunday, 1=Monday, etc.
    price_base DECIMAL(10,2),
    available_seats INTEGER
);
```

---

## Bookings Table

```sql
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    booking_ref VARCHAR(50) UNIQUE,
    schedule_id INTEGER REFERENCES schedules(id),
    booking_date DATE,
    passenger_count INTEGER,
    total_price DECIMAL(10,2),
    status VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE passengers (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER REFERENCES bookings(id),
    name VARCHAR(200),
    passenger_type VARCHAR(20),  -- adult, senior, child
    seat_number VARCHAR(10),
    price DECIMAL(10,2)
);
```

---

## Advanced: Multi-Step Function Calls

**Complex Query: "Book 2 tickets from Milan to Rome tomorrow"**

```
Step 1: AI calls search_trains()
        → Returns available trains

Step 2: AI asks user to choose
        → User: "The 7:30 one"

Step 3: AI calls create_booking()
        → Creates booking in database
        → Returns booking confirmation
```

**The AI orchestrates the entire flow!**

---

## Error Handling with AI

**Database Constraint Violation:**

```go
func executeCreateBooking(args map[string]interface{}) (interface{}, error) {
    booking, err := CreateBooking(req)
    if err != nil {
        // Return error to AI - it will explain to user!
        return nil, fmt.Errorf("booking failed: %v", err)
    }
    return booking, nil
}
```

**AI Response to User:**
```
"I'm sorry, there are no longer enough seats available
on that train. Would you like to see alternative options?"
```

**The AI handles error communication naturally!**

---

## Security Considerations

**🚨 Critical: Validate EVERYTHING**

```go
func executeCreateBooking(args map[string]interface{}) (interface{}, error) {
    scheduleID := getInt(args, "schedule_id")

    // ALWAYS validate schedule exists
    schedule, err := GetScheduleByID(scheduleID)
    if err != nil {
        return nil, fmt.Errorf("invalid schedule")
    }

    // Check seat availability
    if schedule.AvailableSeats < req.PassengerCount {
        return nil, fmt.Errorf("not enough seats")
    }

    // Use transactions!
    tx, _ := db.Begin()
    defer tx.Rollback()

    // ... create booking with proper locking

    tx.Commit()
}
```

---

## SQL Injection Prevention

**❌ NEVER do this:**
```go
// DANGEROUS!
query := fmt.Sprintf("SELECT * FROM schedules WHERE origin = '%s'", origin)
```

**✅ ALWAYS use parameterized queries:**
```go
// SAFE
query := "SELECT * FROM schedules WHERE origin = $1"
db.Query(query, origin)
```

**AI-generated parameters are still user input!**

---

## Rate Limiting & Costs

**AI API calls cost money!**

```go
type RateLimiter struct {
    requests map[string][]time.Time
}

func (rl *RateLimiter) Allow(sessionID string) bool {
    recent := rl.requests[sessionID]

    // Allow max 10 requests per minute
    cutoff := time.Now().Add(-1 * time.Minute)
    recent = filterAfter(recent, cutoff)

    if len(recent) >= 10 {
        return false
    }

    rl.requests[sessionID] = append(recent, time.Now())
    return true
}
```

---

## Monitoring & Logging

**Log everything for debugging:**

```go
func HandleAIChat(c *gin.Context) {
    log.Printf("AI chat request - Session: %s, Message: %s",
               req.SessionID, req.Message)

    // Process...

    if functionCall != nil {
        log.Printf("Executing function: %s with arguments: %v",
                   functionCall.Name, args)
    }

    log.Printf("AI response: %s", response.Message)
}
```

**Track:**
- Function calls executed
- Database queries run
- Errors encountered
- Response times

---

## Testing AI-Database Integration

**Unit Test Example:**

```go
func TestSearchTrainsFunction(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer db.Close()

    // Simulate AI function call
    args := map[string]interface{}{
        "origin":      "MI",
        "destination": "RM",
        "date":        "2025-10-18",
    }

    result, err := executeSearchTrains(args)

    assert.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

---

## Integration Testing

**Test the full AI flow:**

```go
func TestAIChatIntegration(t *testing.T) {
    // Mock OpenAI API
    mockAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Return mock function call
        json.NewEncoder(w).Encode(mockFunctionCallResponse)
    }))

    // Send chat message
    req := ChatRequest{
        SessionID: "test-123",
        Message:   "Show me trains from Milan to Rome",
    }

    resp := sendChatRequest(req)

    assert.Contains(t, resp.Message, "trains")
    assert.NotNil(t, resp.Data)
}
```

---

## Performance Optimization

**Database Indexes:**

```sql
-- Critical for fast lookups
CREATE INDEX idx_schedules_route
    ON schedules(origin_id, destination_id, day_of_week);

CREATE INDEX idx_schedules_departure
    ON schedules(departure_time);

CREATE INDEX idx_bookings_ref
    ON bookings(booking_ref);
```

**Caching:**
```go
var stationCache = make(map[string]*Station)

func mapStationCode(input string) string {
    if cached, exists := stationCache[input]; exists {
        return cached.Code
    }
    // ... query database
}
```

---

## Handling Ambiguity

**AI is great at disambiguation:**

```
User: "I need to go to Rome"

AI:   "I'd be happy to help! Which station would you like
       to depart from? We have service from:
       - Milano Centrale
       - Venezia Santa Lucia
       - Firenze Santa Maria Novella"
```

**Define this in your system prompt:**
```go
systemPrompt := `You are a helpful train booking assistant.
When information is missing, ask the user for clarification.
Available stations: Milan, Rome, Venice, Florence, Naples.`
```

---

## System Prompt Best Practices

**Good System Prompt:**

```
You are an AI assistant for Frecciarossa train bookings.

CAPABILITIES:
- Search trains between Italian cities
- Create bookings with passenger details
- Retrieve existing bookings

RULES:
- Always confirm details before booking
- Dates must be in the future
- Explain prices clearly
- Handle errors gracefully

STATIONS (use these exact names):
- Milano Centrale (MI)
- Roma Termini (RM)
- Venezia Santa Lucia (VE)
- Firenze S.M.N. (FI)
- Napoli Centrale (NA)
```

---

## Multi-Language Support

**AI handles translations naturally!**

```go
systemPrompt := `You are a multilingual assistant.
Respond in the same language the user writes in.
Supported languages: English, Italian.`
```

**User:** "Ho bisogno di andare da Milano a Roma domani"

**AI:** "Certo! Cerco i treni disponibili da Milano a Roma per domani..."

**The AI translates automatically while still calling the same functions!**

---

## Real-World Challenges

**1. Function Hallucination**
- AI might "invent" function parameters
- Solution: Strict validation + clear schemas

**2. Context Window Limits**
- Long conversations exceed token limits
- Solution: Summarize old messages, keep recent context

**3. Latency**
- AI calls + DB queries add up
- Solution: Optimize queries, cache aggressively, show loading states

**4. Costs**
- OpenAI API can get expensive
- Solution: Rate limiting, caching, use cheaper models when possible

---

## Alternative AI Providers

**This demo uses OpenAI-compatible API:**

```go
apiURL := cfg.OpenAIBaseURL + "/v1/chat/completions"
```

**Works with:**
- OpenAI (gpt-4, gpt-3.5-turbo)
- Azure OpenAI
- RelaxAI
- LocalAI (self-hosted)
- Ollama (local models)
- Anthropic Claude (different API format)

**Just change the base URL!**

---

## Local LLM Option: Ollama

**For offline/private deployments:**

```bash
# Install Ollama
curl https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama2

# Run local server
ollama serve
```

```go
// Point to local Ollama
cfg.OpenAIBaseURL = "http://localhost:11434/v1"
cfg.Model = "llama2"
```

**Pros:** Free, private, no internet needed
**Cons:** Less capable, requires powerful hardware

---

## When to Use AI-Database Integration

**✅ Good Use Cases:**
- Customer service chatbots
- Natural language search
- Complex form filling
- Data exploration tools
- Internal admin tools

**❌ Avoid for:**
- High-frequency transactional systems
- Real-time trading/critical systems
- When deterministic behavior is required
- Privacy-sensitive applications (without local models)

---

## Architecture Variants

**1. Direct Integration (This Demo)**
```
User → AI → Backend → Database
```

**2. AI as Middleware**
```
User → AI Agent → Multiple Microservices → Multiple DBs
```

**3. Hybrid Approach**
```
User → Frontend → [Traditional OR AI Path] → Database
```

---

## Extending the System

**Add More Functions:**

```go
{
    "name": "cancel_booking",
    "description": "Cancel an existing booking",
    "parameters": {
        "booking_ref": "string"
    }
}

{
    "name": "get_seat_map",
    "description": "View available seats on a train",
    "parameters": {
        "schedule_id": "integer"
    }
}

{
    "name": "apply_discount_code",
    "description": "Apply a promotional code",
    "parameters": {
        "code": "string",
        "booking_id": "integer"
    }
}
```

---

## Database Best Practices for AI Apps

**1. Use JSONB for Flexibility**
```sql
ALTER TABLE bookings ADD COLUMN metadata JSONB;

-- Store AI conversation context
UPDATE bookings
SET metadata = '{"ai_session": "uuid", "user_intent": "business_travel"}'
WHERE id = 123;
```

**2. Audit Logging**
```sql
CREATE TABLE ai_interactions (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(100),
    user_message TEXT,
    function_called VARCHAR(100),
    result JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## Monitoring AI-Database Performance

**Key Metrics:**

```go
// Prometheus metrics example
var (
    aiFunctionCalls = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_function_calls_total",
        },
        []string{"function_name", "status"},
    )

    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
        },
        []string{"query_type"},
    )
)
```

---

## Future Enhancements

**1. Streaming Responses**
```javascript
const response = await fetch('/api/ai/chat', {
    method: 'POST',
    body: JSON.stringify(request)
});

const reader = response.body.getReader();
// Stream AI responses in real-time
```

**2. Multi-Step Workflows**
- AI plans complex operations
- Executes multiple DB transactions
- Confirms with user at each step

**3. Learning from Interactions**
- Store successful query patterns
- Fine-tune models on your domain
- Improve function calling accuracy

---

## Code Repository Structure

```
train-ticketing-ai/
├── backend/
│   ├── main.go                 # Entry point
│   ├── config/                 # Configuration
│   ├── handlers/               # HTTP handlers
│   ├── services/
│   │   ├── ai_service.go       # AI integration ⭐
│   │   ├── booking_service.go  # Database operations ⭐
│   │   └── search_service.go
│   ├── models/                 # Data models
│   └── database/               # DB connection
├── frontend/
│   ├── js/
│   │   ├── ai_chat.js         # AI chat UI
│   │   └── traditional.js      # Traditional forms
│   └── css/
├── init-db.sql                 # Database schema ⭐
└── docker-compose.yml          # Infrastructure
```

---

## Key Takeaways

**1. Function Calling is the Bridge**
- LLMs understand intent
- Your code executes actions
- Database stores state

**2. Security First**
- Validate all AI-generated parameters
- Use parameterized queries
- Implement rate limiting

**3. Context Matters**
- Maintain sessions for conversations
- Provide clear system prompts
- Test edge cases thoroughly

**4. Monitor Everything**
- Log AI decisions
- Track database performance
- Watch costs closely

---

## Common Pitfalls to Avoid

**1. Trusting AI Output Blindly**
```go
// BAD
scheduleID := args["schedule_id"].(int)

// GOOD
scheduleID, ok := args["schedule_id"].(int)
if !ok || scheduleID <= 0 {
    return nil, errors.New("invalid schedule ID")
}
```

**2. Not Handling Function Failures**
- AI will try again if you return proper errors
- Provide helpful error messages

**3. Ignoring Token Limits**
- Conversations grow large
- Implement context pruning

---

## Demo Time! 🚀

**Live Demo Features:**

1. Natural language train search
   - "I need to go from Milan to Rome tomorrow morning"

2. AI-powered booking
   - "Book 2 tickets on the 7:30 train"

3. Multi-turn conversation
   - AI remembers context

4. Error handling
   - Try booking with no seats available

5. Language switching
   - English ↔️ Italian on the fly

---

## Resources & Further Reading

**OpenAI Function Calling:**
- https://platform.openai.com/docs/guides/function-calling

**Database Design:**
- PostgreSQL Documentation
- SQL Performance Best Practices

**This Project:**
- GitHub: [your-repo-url]
- Live Demo: [demo-url]

**Related Tools:**
- LangChain (Python/JS framework for LLM apps)
- Semantic Kernel (Microsoft's LLM framework)
- AutoGen (Multi-agent systems)

---

## Q&A

**Questions?**

Topics we can dive deeper into:
- Advanced function calling patterns
- Database optimization techniques
- Production deployment strategies
- Cost optimization
- Alternative AI providers
- Security hardening

**Contact:**
- GitHub: [your-github]
- Email: [your-email]
- Twitter: [your-twitter]

---

## Thank You! 🙏

**Try it yourself:**
```bash
git clone [repo-url]
cd train-ticketing-ai
docker-compose up -d
open http://localhost:8080
```

**Remember:**
- Start simple (one function)
- Test thoroughly
- Monitor in production
- Users love natural interfaces!

---
