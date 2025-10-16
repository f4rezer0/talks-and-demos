package models

import (
	"encoding/json"
	"time"
)

// ConversationHistory represents a stored conversation message
type ConversationHistory struct {
	ID           int             `json:"id"`
	SessionID    string          `json:"session_id"`
	Role         string          `json:"role"` // user, assistant, system
	Message      string          `json:"message"`
	FunctionCall json.RawMessage `json:"function_call,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
}

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

// ChatResponse represents an AI chat response
type ChatResponse struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
	FunctionCall interface{} `json:"function_call,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

// FunctionCall represents an AI function call
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
