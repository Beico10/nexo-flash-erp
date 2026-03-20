// Package handlers — handler HTTP do Co-Piloto IA.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/nexoone/nexo-one/internal/gemini"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

const systemPromptCopilot = `Voce e o Co-Piloto do Nexo One ERP, um assistente inteligente para negocios brasileiros.
Voce ajuda com: gestao financeira, fiscal (IBS/CBS 2026), estoque, agendamentos, logistica, producao.
Responda sempre em portugues brasileiro, de forma direta e acionavel.
Quando receber dados do sistema, analise e de sugestoes proativas.
Use emojis moderadamente. Seja conciso.`

// CopilotHandler gerencia o endpoint do Co-Piloto IA.
type CopilotHandler struct {
	client *gemini.Client
}

// NewCopilotHandler cria um novo handler do Co-Piloto.
func NewCopilotHandler(client *gemini.Client) *CopilotHandler {
	return &CopilotHandler{client: client}
}

// RegisterRoutes registra as rotas do Co-Piloto.
func (h *CopilotHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/copilot/suggest", h.Suggest)
	mux.HandleFunc("POST /api/v1/copilot/clear", h.ClearSession)
	mux.HandleFunc("GET /api/v1/copilot/history", h.GetHistory)
}

// SuggestRequest é o payload de entrada do Co-Piloto.
type SuggestRequest struct {
	Question  string `json:"question"`
	Context   string `json:"context,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// SuggestResponse é a resposta do Co-Piloto.
type SuggestResponse struct {
	Suggestion string `json:"suggestion"`
	Model      string `json:"model"`
	SessionID  string `json:"session_id"`
}

// Suggest processa uma pergunta do usuário via Co-Piloto.
func (h *CopilotHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}

	var req SuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON invalido")
		return
	}

	if req.Question == "" {
		respondError(w, http.StatusBadRequest, "question e obrigatorio")
		return
	}

	// Define session ID
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "copilot-" + tenantID
	}

	// Monta prompt com contexto
	prompt := req.Question
	if req.Context != "" {
		prompt = "Dados do sistema:\n" + req.Context + "\n\nPergunta: " + req.Question
	}

	// Chama Gemini
	response, err := h.client.Chat(r.Context(), sessionID, prompt, systemPromptCopilot)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao processar: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, SuggestResponse{
		Suggestion: response,
		Model:      "gemini-2.0-flash",
		SessionID:  sessionID,
	})
}

// ClearSession limpa o histórico de uma sessão.
func (h *CopilotHandler) ClearSession(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "copilot-" + tenantID
	}

	h.client.ClearSession(sessionID)
	respondJSON(w, http.StatusOK, map[string]string{"message": "sessao limpa"})
}

// GetHistory retorna o histórico de uma sessão.
func (h *CopilotHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = "copilot-" + tenantID
	}

	history := h.client.GetHistory(sessionID)
	respondJSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID,
		"messages":   history,
	})
}
