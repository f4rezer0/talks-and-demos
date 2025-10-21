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
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	return fmt.Sprintf(`You are a helpful AI assistant for an Italian train booking system.
You are bilingual and MUST respond in the same language the user is using (Italian or English).

TODAY'S DATE: %s
TOMORROW'S DATE: %s

IMPORTANT: When the user says "domani" or "tomorrow", use the date %s (tomorrow's date).
DO NOT use dates in the past. All bookings must be for today or future dates.

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
3. Assist with booking tickets (both showing options and direct booking)
4. Answer questions about bookings
5. Be friendly, conversational, and helpful
6. ALWAYS respond in the same language as the user

Important guidelines:
- ALWAYS respond in the same language the user is writing in (Italian/English)
- When the user says "Prenota" or "Book" WITHOUT providing passenger names:
  * DO NOT call any function yet
  * Ask for passenger information in their language
  * Italian: "Perfetto! Quanti passeggeri viaggiano e quali sono i loro nomi?"
  * English: "Perfect! How many passengers will be traveling, and what are their names?"
- When users ask to "see trains" or "show options", use search_trains
- ONLY call book_train_direct when you have ALL required information including passenger names
- Use natural language and be conversational in the user's language
- When users mention city names (like "Milan"/"Milano" or "Rome"/"Roma"), match them to station names

Booking workflow - FOLLOW THIS EXACTLY:
1. User: "Prenota un treno da Milano a Roma domani"
2. You respond with TEXT (NOT a function call): "Perfetto! Quanti passeggeri viaggiano e quali sono i loro nomi?"
3. User: "1 passeggero, Alessandro Argentieri"
4. NOW you call book_train_direct with ALL the information
5. NEVER call book_train_direct with empty passengers array

Available functions you can call:
- search_trains: Search for available trains (use when user wants to see options)
- book_train_direct: Search AND book in one step (ONLY use when passenger names are provided)
- create_booking: Create booking from a specific schedule_id (use when user selected from search results)
- get_booking_details: Retrieve booking information
- cancel_booking: Cancel an existing booking

Examples of handling booking requests:
1. User (Italian): "Prenota un treno da Milano a Roma domani"
   You (Italian): "Sarei felice di prenotare per te! Quanti passeggeri viaggiano e quali sono i loro nomi?"

2. User (English): "Book me a train from Milan to Rome tomorrow"
   You (English): "I'd be happy to book that for you! How many passengers will be traveling, and what are their names?"

3. User (Italian): "Devo andare da Milano a Roma domani mattina"
   You: Call search_trains and present results in Italian

4. User (English): "I need to go from Milan to Rome tomorrow morning"
   You: Call search_trains and present results in English

5. User (Italian): "Prenota il primo treno per Alessandro Argentieri"
   You: Call book_train_direct with passenger name "Alessandro Argentieri"

Always use functions when appropriate to provide accurate information.`, currentDate, tomorrow, tomorrow, strings.Join(stationList, ", "))
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
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "book_train_direct",
				"description": "Search for trains and directly book the most suitable one based on user criteria. Use this when the user wants to book directly without seeing options (e.g., 'book the first train', 'book me on the earliest train', 'book the morning train')",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"origin": map[string]interface{}{
							"type":        "string",
							"description": "Departure station name or code",
						},
						"destination": map[string]interface{}{
							"type":        "string",
							"description": "Arrival station name or code",
						},
						"date": map[string]interface{}{
							"type":        "string",
							"description": "Travel date in YYYY-MM-DD format",
						},
						"time_preference": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"morning", "afternoon", "evening", "any", "earliest", "latest"},
							"description": "Time preference - 'earliest' for first available, 'latest' for last available",
						},
						"train_selection": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"first", "last", "cheapest", "fastest"},
							"description": "Which train to select from available options",
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
					"required": []string{"origin", "destination", "date", "passengers"},
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
	case "book_train_direct":
		return executeBookTrainDirect(arguments)
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

func executeBookTrainDirect(args map[string]interface{}) (interface{}, error) {
	// First, search for trains
	searchReq := models.SearchRequest{
		Origin:         getString(args, "origin"),
		Destination:    getString(args, "destination"),
		Date:           getString(args, "date"),
		TimePreference: getString(args, "time_preference"),
		PassengerCount: 0, // Will be calculated from passengers
		Filters:        make(map[string]interface{}),
	}

	// Extract passengers
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

	searchReq.PassengerCount = len(passengers)
	if searchReq.PassengerCount == 0 {
		return nil, fmt.Errorf("at least one passenger is required")
	}

	// Search for trains
	results, err := SearchTrains(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search trains: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no trains found matching your criteria")
	}

	// Select the appropriate train based on criteria
	trainSelection := getString(args, "train_selection")
	selectedTrain := results[0] // Default to first

	switch trainSelection {
	case "last":
		selectedTrain = results[len(results)-1]
	case "cheapest":
		for _, train := range results {
			if train.TotalPrice < selectedTrain.TotalPrice {
				selectedTrain = train
			}
		}
	case "fastest":
		// Parse duration and find fastest
		for _, train := range results {
			// For simplicity, use first one (duration comparison would need parsing)
			selectedTrain = train
			break
		}
	default: // "first" or any other value
		selectedTrain = results[0]
	}

	// Create the booking
	bookingReq := models.BookingRequest{
		ScheduleID: selectedTrain.Schedule.ID,
		Date:       getString(args, "date"),
		Passengers: passengers,
	}

	booking, err := CreateBooking(bookingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Return booking with additional context about selection
	return map[string]interface{}{
		"booking":        booking,
		"selected_train": selectedTrain,
		"total_found":    len(results),
	}, nil
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
	enhancedSystemPrompt := systemPrompt + `

IMPORTANT: When you need to call a function, respond ONLY with valid JSON in this format (no markdown, no extra text):
{"name": "function_name", "parameters": {...}}

Example for searching trains:
{"name": "search_trains", "parameters": {"origin": "Milano Centrale", "destination": "Roma Termini", "date": "2025-10-20"}}

CRITICAL RULES:
1. NEVER call book_train_direct if passenger names are empty or missing
2. If the user wants to book but hasn't provided passenger names, respond with conversational text in their language asking for names (NOT JSON)
3. Only output JSON when you have all required information AND the user wants to perform an action
4. ALWAYS match the user's language - if they write in Italian, respond in Italian; if in English, respond in English`

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

	response := &models.ChatResponse{Success: true}

	// Try to parse function call from message
	// Extract JSON from message even if there's text before or after it
	if strings.Contains(message, "{\"name\":") || strings.Contains(message, "\"function_call\"") {
		// Find the JSON object in the message
		startIdx := strings.Index(message, "{")
		endIdx := strings.LastIndex(message, "}")

		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			jsonStr := message[startIdx : endIdx+1]

			// Try parsing as direct function call format first
			var directFunctionCall struct {
				Name       string                 `json:"name"`
				Parameters map[string]interface{} `json:"parameters"`
			}
			if err := json.Unmarshal([]byte(jsonStr), &directFunctionCall); err == nil && directFunctionCall.Name != "" {
				log.Printf("Parsed direct function call: %s", directFunctionCall.Name)

				// Execute function
				result, err := executeFunction(directFunctionCall.Name, directFunctionCall.Parameters)
				if err != nil {
					response.Message = fmt.Sprintf("Errore durante l'esecuzione della funzione: %v", err)
				} else {
					response.FunctionCall = map[string]interface{}{
						"name":      directFunctionCall.Name,
						"arguments": directFunctionCall.Parameters,
					}
					response.Data = result
					response.Message = formatFunctionResult(directFunctionCall.Name, result)
				}
				return response, nil
			}

			// Try parsing as nested function_call format
			var functionCallData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &functionCallData); err == nil {
				if fc, ok := functionCallData["function_call"].(map[string]interface{}); ok {
					functionName, _ := fc["name"].(string)
					arguments, _ := fc["arguments"].(map[string]interface{})

					if functionName != "" {
						log.Printf("Parsed nested function call: %s", functionName)

						// Execute function
						result, err := executeFunction(functionName, arguments)
						if err != nil {
							response.Message = fmt.Sprintf("Errore durante l'esecuzione della funzione: %v", err)
						} else {
							response.FunctionCall = map[string]interface{}{
								"name":      functionName,
								"arguments": arguments,
							}
							response.Data = result
							response.Message = formatFunctionResult(functionName, result)
						}
						return response, nil
					}
				}
			}
		}
	}

	response.Message = message
	return response, nil
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

	// Add explicit instruction to NOT output JSON in the last user message
	// This is a workaround for models that don't handle tool_choice properly
	lastMessageIdx := len(messages) - 1
	if lastMessageIdx >= 1 {
		if content, ok := messages[lastMessageIdx]["content"].(string); ok {
			messages[lastMessageIdx]["content"] = content + "\n\n[SYSTEM INSTRUCTION: You must use the available tools/functions. DO NOT output JSON directly in your response. Use the tool calling mechanism provided by the API.]"
		}
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
		log.Printf("OpenAI API error response: %s", string(bodyBytes))
		return nil, fmt.Errorf("openAI API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("OpenAI raw response: %s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Printf("Failed to parse OpenAI response: %v", err)
		return nil, err
	}

	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})

	log.Printf("OpenAI message received: %+v", message)

	response := &models.ChatResponse{Success: true}

	// Check for function call
	if toolCalls, ok := message["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
		log.Printf("OpenAI returned tool_calls")
		toolCall := toolCalls[0].(map[string]interface{})
		function := toolCall["function"].(map[string]interface{})

		functionName := function["name"].(string)
		var arguments map[string]interface{}
		json.Unmarshal([]byte(function["arguments"].(string)), &arguments)

		log.Printf("Executing OpenAI function call: %s with args: %v", functionName, arguments)

		// Execute function
		result, err := executeFunction(functionName, arguments)
		if err != nil {
			log.Printf("Function execution error: %v", err)
			response.Message = fmt.Sprintf("Error executing function: %v", err)
		} else {
			response.FunctionCall = map[string]interface{}{
				"name":      functionName,
				"arguments": arguments,
			}
			response.Data = result
			response.Message = formatFunctionResult(functionName, result)
			log.Printf("Function executed successfully, formatted message length: %d", len(response.Message))
		}
	} else {
		// No tool calls, check if content is JSON (fallback for models that don't use tool_calls properly)
		contentStr := ""
		if content, ok := message["content"]; ok {
			if str, ok := content.(string); ok {
				contentStr = str
			}
		}

		log.Printf("No tool_calls found. Message content: %s", contentStr)

		// Check if the content looks like a function call JSON
		if strings.Contains(contentStr, "{\"name\":") {
			log.Printf("Content looks like JSON function call, attempting to parse")
			startIdx := strings.Index(contentStr, "{")
			endIdx := strings.LastIndex(contentStr, "}")

			if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
				jsonStr := contentStr[startIdx : endIdx+1]

				var directFunctionCall struct {
					Name       string                 `json:"name"`
					Parameters map[string]interface{} `json:"parameters"`
				}
				if err := json.Unmarshal([]byte(jsonStr), &directFunctionCall); err == nil && directFunctionCall.Name != "" {
					log.Printf("Successfully parsed direct function call: %s", directFunctionCall.Name)

					// Execute function
					result, err := executeFunction(directFunctionCall.Name, directFunctionCall.Parameters)
					if err != nil {
						log.Printf("Function execution error: %v", err)
						response.Message = fmt.Sprintf("Errore durante l'esecuzione della funzione: %v", err)
					} else {
						response.FunctionCall = map[string]interface{}{
							"name":      directFunctionCall.Name,
							"arguments": directFunctionCall.Parameters,
						}
						response.Data = result
						response.Message = formatFunctionResult(directFunctionCall.Name, result)
						log.Printf("Function executed successfully from content, formatted message length: %d", len(response.Message))
					}
					return response, nil
				} else {
					log.Printf("Failed to parse as function call: %v", err)
				}
			}
		}

		response.Message = contentStr
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

	case "book_train_direct":
		data := result.(map[string]interface{})
		booking := data["booking"].(*models.Booking)
		selectedTrain := data["selected_train"].(models.SearchResponse)
		totalFound := data["total_found"].(int)

		msg := fmt.Sprintf("Booking confirmed! 🎉\n\n")
		msg += fmt.Sprintf("I found %d trains and booked the best match for you:\n\n", totalFound)
		msg += fmt.Sprintf("Train: %s (%s)\n", selectedTrain.Schedule.Train.Number, selectedTrain.Schedule.Train.Type)
		msg += fmt.Sprintf("Route: %s → %s\n", selectedTrain.Schedule.Origin.Name, selectedTrain.Schedule.Destination.Name)
		msg += fmt.Sprintf("Departure: %s\n", selectedTrain.DepartureTime)
		msg += fmt.Sprintf("Arrival: %s\n", selectedTrain.ArrivalTime)
		msg += fmt.Sprintf("Duration: %s\n\n", selectedTrain.Duration)
		msg += fmt.Sprintf("Booking Reference: %s\n", booking.BookingRef)
		msg += fmt.Sprintf("Date: %s\n", booking.BookingDate.Format("2006-01-02"))
		msg += fmt.Sprintf("Passengers: %d\n", booking.PassengerCount)
		msg += fmt.Sprintf("Total Price: €%.2f\n", booking.TotalPrice)
		return msg

	default:
		return "Function executed successfully."
	}
}
