package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"train-ticketing-ai/config"
	"train-ticketing-ai/database"
	"train-ticketing-ai/models"
)

var cfg *config.Config

// InitAIService initializes the AI service with configuration
func InitAIService(config *config.Config) {
	cfg = config
}

// ProcessMessage processes a user message and returns AI response
func ProcessMessage(sessionID, message string) (*models.ChatResponse, error) {
	db := database.GetDB()

	// Save user message
	_, err := db.Exec(`
		INSERT INTO conversation_history (session_id, role, message)
		VALUES ($1, 'user', $2)
	`, sessionID, message)
	if err != nil {
		log.Printf("Failed to save user message: %v", err)
	}

	// Load conversation history
	history, err := loadConversationHistory(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation history: %w", err)
	}

	// Build system prompt
	systemPrompt := buildSystemPrompt()

	// Call AI provider
	var response *models.ChatResponse
	switch cfg.AIProvider {
	case "openai":
		response, err = callOpenAI(systemPrompt, history, message)
	case "anthropic":
		response, err = callAnthropic(systemPrompt, history, message)
	case "ollama":
		response, err = callOllama(systemPrompt, history, message)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.AIProvider)
	}

	if err != nil {
		return nil, fmt.Errorf("AI provider error: %w", err)
	}

	// Save assistant response
	functionCallJSON, _ := json.Marshal(response.FunctionCall)
	_, err = db.Exec(`
		INSERT INTO conversation_history (session_id, role, message, function_call)
		VALUES ($1, 'assistant', $2, $3)
	`, sessionID, response.Message, functionCallJSON)
	if err != nil {
		log.Printf("Failed to save assistant message: %v", err)
	}

	return response, nil
}

// loadConversationHistory loads recent conversation history
func loadConversationHistory(sessionID string) ([]models.ConversationHistory, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, session_id, role, message, function_call, timestamp
		FROM conversation_history
		WHERE session_id = $1
		ORDER BY timestamp DESC
		LIMIT 20
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ConversationHistory
	for rows.Next() {
		var h models.ConversationHistory
		var functionCall sql.NullString

		err := rows.Scan(&h.ID, &h.SessionID, &h.Role, &h.Message, &functionCall, &h.Timestamp)
		if err != nil {
			return nil, err
		}

		if functionCall.Valid {
			h.FunctionCall = json.RawMessage(functionCall.String)
		}

		history = append(history, h)
	}

	// Reverse to chronological order
	for i := 0; i < len(history)/2; i++ {
		j := len(history) - i - 1
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// buildSystemPrompt creates the system prompt with available functions
func buildSystemPrompt() string {
	stations, _ := GetAllStations()
	stationList := make([]string, len(stations))
	for i, s := range stations {
		stationList[i] = fmt.Sprintf("%s (%s)", s.Name, s.Code)
	}

	currentDate := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`You are a helpful AI assistant for an Italian train booking system.

Current date: %s

Available stations:
%s

Train types and pricing:
- Frecciarossa (FR): High-speed trains, €0.15/km, has WiFi and food service
- Intercity (IC): Medium-speed trains, €0.10/km, has WiFi
- Regionale (RG): Regional trains, €0.06/km, basic service

Passenger types and discounts:
- Adult: Full price (100%%)
- Senior (65+): 20%% discount (80%% of base price)
- Child (4-17): 30%% discount (70%% of base price)
- Infant (0-3): Free, no seat assigned

Your role:
1. Help users search for trains between stations
2. Provide train recommendations based on their preferences
3. Assist with booking tickets
4. Answer questions about bookings
5. Be friendly, conversational, and helpful

Important guidelines:
- Always ask for clarification if information is missing
- Confirm booking details before creating a booking
- Use natural language and be conversational
- When users mention city names (like "Milan" or "Rome"), match them to station names
- Provide helpful suggestions and alternatives

Available functions you can call:
- search_trains: Search for available trains
- create_booking: Create a new booking
- get_booking_details: Retrieve booking information
- cancel_booking: Cancel an existing booking

Always use functions when appropriate to provide accurate information.`, currentDate, strings.Join(stationList, ", "))
}

// Function definitions for AI
func getFunctionDefinitions() interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "search_trains",
				"description": "Search for available train schedules between two stations on a specific date",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"origin": map[string]interface{}{
							"type":        "string",
							"description": "Departure station name or code (e.g., 'Milano Centrale' or 'MI')",
						},
						"destination": map[string]interface{}{
							"type":        "string",
							"description": "Arrival station name or code (e.g., 'Roma Termini' or 'RM')",
						},
						"date": map[string]interface{}{
							"type":        "string",
							"description": "Travel date in YYYY-MM-DD format",
						},
						"time_preference": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"morning", "afternoon", "evening", "any"},
							"description": "Preferred time of day for travel",
						},
						"passenger_count": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of passengers",
							"minimum":     1,
						},
					},
					"required": []string{"origin", "destination", "date"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "create_booking",
				"description": "Create a train ticket booking for passengers",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"schedule_id": map[string]interface{}{
							"type":        "integer",
							"description": "The schedule ID from search results",
						},
						"date": map[string]interface{}{
							"type":        "string",
							"description": "Travel date in YYYY-MM-DD format",
						},
						"passengers": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type":        "string",
										"description": "Passenger full name",
									},
									"passenger_type": map[string]interface{}{
										"type":        "string",
										"enum":        []string{"adult", "senior", "child", "infant"},
										"description": "Type of passenger",
									},
								},
								"required": []string{"name", "passenger_type"},
							},
							"description": "List of passengers for this booking",
						},
					},
					"required": []string{"schedule_id", "date", "passengers"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_booking_details",
				"description": "Retrieve details of an existing booking by reference number",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"booking_ref": map[string]interface{}{
							"type":        "string",
							"description": "Booking reference number (e.g., 'TRN-2025-00001')",
						},
					},
					"required": []string{"booking_ref"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "cancel_booking",
				"description": "Cancel an existing booking",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"booking_ref": map[string]interface{}{
							"type":        "string",
							"description": "Booking reference number to cancel",
						},
					},
					"required": []string{"booking_ref"},
				},
			},
		},
	}
}

// executeFunction executes a function call and returns the result
func executeFunction(functionName string, arguments map[string]interface{}) (interface{}, error) {
	log.Printf("Executing function: %s with arguments: %v", functionName, arguments)

	switch functionName {
	case "search_trains":
		return executSearchTrains(arguments)
	case "create_booking":
		return executeCreateBooking(arguments)
	case "get_booking_details":
		return executeGetBookingDetails(arguments)
	case "cancel_booking":
		return executeCancelBooking(arguments)
	default:
		return nil, fmt.Errorf("unknown function: %s", functionName)
	}
}

func executSearchTrains(args map[string]interface{}) (interface{}, error) {
	req := models.SearchRequest{
		Origin:         getString(args, "origin"),
		Destination:    getString(args, "destination"),
		Date:           getString(args, "date"),
		TimePreference: getString(args, "time_preference"),
		PassengerCount: getInt(args, "passenger_count"),
		Filters:        make(map[string]interface{}),
	}

	if req.PassengerCount == 0 {
		req.PassengerCount = 1
	}

	results, err := SearchTrains(req)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func executeCreateBooking(args map[string]interface{}) (interface{}, error) {
	scheduleID := getInt(args, "schedule_id")
	date := getString(args, "date")
	passengersRaw := args["passengers"]

	passengers := []models.PassengerCreateRequest{}
	if pList, ok := passengersRaw.([]interface{}); ok {
		for _, p := range pList {
			if pMap, ok := p.(map[string]interface{}); ok {
				passengers = append(passengers, models.PassengerCreateRequest{
					Name:          getString(pMap, "name"),
					PassengerType: getString(pMap, "passenger_type"),
				})
			}
		}
	}

	req := models.BookingRequest{
		ScheduleID: scheduleID,
		Date:       date,
		Passengers: passengers,
	}

	booking, err := CreateBooking(req)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func executeGetBookingDetails(args map[string]interface{}) (interface{}, error) {
	bookingRef := getString(args, "booking_ref")
	return GetBooking(bookingRef)
}

func executeCancelBooking(args map[string]interface{}) (interface{}, error) {
	bookingRef := getString(args, "booking_ref")
	err := CancelBooking(bookingRef)
	if err != nil {
		return nil, err
	}
	return map[string]string{"status": "cancelled", "booking_ref": bookingRef}, nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

// callOllama calls the Ollama API
func callOllama(systemPrompt string, history []models.ConversationHistory, userMessage string) (*models.ChatResponse, error) {
	// Build messages
	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	for _, h := range history {
		messages = append(messages, map[string]interface{}{
			"role":    h.Role,
			"content": h.Message,
		})
	}

	// Ollama doesn't support function calling natively, so we include function definitions in the system prompt
	// and parse the response for function calls
	enhancedSystemPrompt := systemPrompt + "\n\nIf you need to call a function, respond with JSON in this format: {\"function_call\": {\"name\": \"function_name\", \"arguments\": {...}}}"

	messages[0]["content"] = enhancedSystemPrompt

	reqBody := map[string]interface{}{
		"model":    cfg.OllamaModel,
		"messages": messages,
		"stream":   false,
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(cfg.OllamaURL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	message := ""
	if msg, ok := result["message"].(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			message = content
		}
	}

	// Try to parse function call from message
	// Simple heuristic: look for JSON-like function call
	if strings.Contains(message, "\"function_call\"") {
		// Try to extract and execute function call
		// For demo purposes, we'll use a simple approach
		// In production, use more robust JSON parsing
	}

	return &models.ChatResponse{
		Success: true,
		Message: message,
	}, nil
}

// callOpenAI calls the OpenAI API with function calling
func callOpenAI(systemPrompt string, history []models.ConversationHistory, userMessage string) (*models.ChatResponse, error) {
	// Build messages
	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	for _, h := range history {
		messages = append(messages, map[string]interface{}{
			"role":    h.Role,
			"content": h.Message,
		})
	}

	reqBody := map[string]interface{}{
		"model":       "Llama-4-Maverick-17B-128E",
		"messages":    messages,
		"tools":       getFunctionDefinitions(),
		"tool_choice": "auto",
	}

	body, _ := json.Marshal(reqBody)
	apiURL := cfg.OpenAIBaseURL + "/v1/chat/completions"
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openAI API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})

	response := &models.ChatResponse{Success: true}

	// Check for function call
	if toolCalls, ok := message["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
		toolCall := toolCalls[0].(map[string]interface{})
		function := toolCall["function"].(map[string]interface{})

		functionName := function["name"].(string)
		var arguments map[string]interface{}
		json.Unmarshal([]byte(function["arguments"].(string)), &arguments)

		// Execute function
		result, err := executeFunction(functionName, arguments)
		if err != nil {
			response.Message = fmt.Sprintf("Error executing function: %v", err)
		} else {
			response.FunctionCall = map[string]interface{}{
				"name":      functionName,
				"arguments": arguments,
			}
			response.Data = result
			response.Message = formatFunctionResult(functionName, result)
		}
	} else {
		response.Message = message["content"].(string)
	}

	return response, nil
}

// callAnthropic calls the Anthropic Claude API
func callAnthropic(systemPrompt string, history []models.ConversationHistory, userMessage string) (*models.ChatResponse, error) {
	// Similar to OpenAI but with Anthropic's API format
	// Anthropic uses a different function calling format
	// For brevity, implementing a simplified version

	messages := []map[string]interface{}{}
	for _, h := range history {
		if h.Role != "system" {
			messages = append(messages, map[string]interface{}{
				"role":    h.Role,
				"content": h.Message,
			})
		}
	}

	reqBody := map[string]interface{}{
		"model":      "claude-3-sonnet-20240229",
		"max_tokens": 2048,
		"system":     systemPrompt,
		"messages":   messages,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := result["content"].([]interface{})
	if len(content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}

	textContent := content[0].(map[string]interface{})
	message := textContent["text"].(string)

	return &models.ChatResponse{
		Success: true,
		Message: message,
	}, nil
}

// formatFunctionResult formats function results into natural language
func formatFunctionResult(functionName string, result interface{}) string {
	switch functionName {
	case "search_trains":
		trains := result.([]models.SearchResponse)
		if len(trains) == 0 {
			return "I couldn't find any trains matching your criteria."
		}
		msg := fmt.Sprintf("I found %d available trains:\n\n", len(trains))
		for i, train := range trains {
			msg += fmt.Sprintf("%d. Train %s (%s)\n", i+1, train.Schedule.Train.Number, train.Schedule.Train.Type)
			msg += fmt.Sprintf("   Departure: %s from %s\n", train.DepartureTime, train.Schedule.Origin.Name)
			msg += fmt.Sprintf("   Arrival: %s at %s\n", train.ArrivalTime, train.Schedule.Destination.Name)
			msg += fmt.Sprintf("   Duration: %s\n", train.Duration)
			msg += fmt.Sprintf("   Price: €%.2f per person (Total: €%.2f)\n", train.PricePerPerson, train.TotalPrice)
			msg += fmt.Sprintf("   Amenities: WiFi: %v, Food: %v\n", train.Schedule.Train.HasWifi, train.Schedule.Train.HasFood)
			msg += fmt.Sprintf("   Schedule ID: %d\n\n", train.Schedule.ID)
		}
		return msg

	case "create_booking":
		booking := result.(*models.Booking)
		msg := fmt.Sprintf("Booking confirmed! 🎉\n\n")
		msg += fmt.Sprintf("Booking Reference: %s\n", booking.BookingRef)
		msg += fmt.Sprintf("Train: %s from %s to %s\n", booking.Schedule.Train.Number,
			booking.Schedule.Origin.Name, booking.Schedule.Destination.Name)
		msg += fmt.Sprintf("Date: %s\n", booking.BookingDate.Format("2006-01-02"))
		msg += fmt.Sprintf("Total Price: €%.2f\n", booking.TotalPrice)
		msg += fmt.Sprintf("Passengers: %d\n", booking.PassengerCount)
		return msg

	case "get_booking_details":
		booking := result.(*models.Booking)
		msg := fmt.Sprintf("Booking Details:\n\n")
		msg += fmt.Sprintf("Reference: %s\n", booking.BookingRef)
		msg += fmt.Sprintf("Status: %s\n", booking.Status)
		msg += fmt.Sprintf("Train: %s from %s to %s\n", booking.Schedule.Train.Number,
			booking.Schedule.Origin.Name, booking.Schedule.Destination.Name)
		msg += fmt.Sprintf("Date: %s\n", booking.BookingDate.Format("2006-01-02"))
		msg += fmt.Sprintf("Total Price: €%.2f\n", booking.TotalPrice)
		return msg

	case "cancel_booking":
		data := result.(map[string]string)
		return fmt.Sprintf("Booking %s has been cancelled successfully.", data["booking_ref"])

	default:
		return "Function executed successfully."
	}
}
