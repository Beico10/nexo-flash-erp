// Package gemini implementa cliente HTTP para serviço de IA.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	defaultTimeout = 60 * time.Second
	// Serviço local de IA (Python)
	aiServiceEndpoint = "http://127.0.0.1:8003/chat"
)

// Client é o cliente HTTP para serviço de IA.
type Client struct {
	httpClient *http.Client
	sessions   map[string][]Message
	mu         sync.RWMutex
}

// Message representa uma mensagem no histórico.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest é o payload enviado para o serviço de IA.
type ChatRequest struct {
	Question     string `json:"question"`
	SessionID    string `json:"session_id"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	Context      string `json:"context,omitempty"`
}

// ChatResponse é a resposta do serviço de IA.
type ChatResponse struct {
	Suggestion string `json:"suggestion,omitempty"`
	Model      string `json:"model,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// NewClient cria um novo cliente de IA.
func NewClient(_ string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		sessions:   make(map[string][]Message),
	}
}

// WithModel é mantido por compatibilidade mas não é usado.
func (c *Client) WithModel(_ string) *Client {
	return c
}

// Chat envia uma mensagem e retorna a resposta.
func (c *Client) Chat(ctx context.Context, sessionID, userMessage, systemPrompt string) (string, error) {
	req := ChatRequest{
		Question:     userMessage,
		SessionID:    sessionID,
		SystemPrompt: systemPrompt,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", aiServiceEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("erro na requisicao ao servico de IA: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("erro ao deserializar resposta: %w", err)
	}

	if chatResp.Error != "" {
		return "", fmt.Errorf("erro do servico de IA: %s", chatResp.Error)
	}

	// Atualiza histórico local
	c.mu.Lock()
	c.sessions[sessionID] = append(c.sessions[sessionID], Message{Role: "user", Content: userMessage})
	c.sessions[sessionID] = append(c.sessions[sessionID], Message{Role: "assistant", Content: chatResp.Suggestion})
	c.mu.Unlock()

	return chatResp.Suggestion, nil
}

// ClearSession limpa o histórico de uma sessão.
func (c *Client) ClearSession(sessionID string) {
	c.mu.Lock()
	delete(c.sessions, sessionID)
	c.mu.Unlock()
}

// GetHistory retorna o histórico de uma sessão.
func (c *Client) GetHistory(sessionID string) []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessions[sessionID]
}
