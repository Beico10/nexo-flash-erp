// Package auth — handlers HTTP de autenticação.
package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Handler agrupa os endpoints de autenticação.
type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// RegisterRoutes registra as rotas de autenticação.
// Estas rotas NÃO passam pelo AuthMiddleware — são públicas.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login",   h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
	mux.HandleFunc("POST /auth/logout",  h.Logout)
	mux.HandleFunc("GET  /auth/me",      h.Me)
}

// Login autentica usuário e retorna par de tokens.
// POST /auth/login
// Body: { "tenant_slug": "mecanica-joao", "email": "joao@example.com", "password": "..." }
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Email == "" || req.Password == "" || req.TenantSlug == "" {
		respondError(w, http.StatusBadRequest, "tenant_slug, email e password são obrigatórios")
		return
	}

	pair, err := h.service.Login(r.Context(), req)
	if err != nil {
		// Sempre 401 para não revelar qual campo está errado
		respondError(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}

	// Refresh token em cookie HttpOnly (mais seguro que retornar no body)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		Expires:  pair.RefreshExpiresAt,
		HttpOnly: true,
		Secure:   true,   // apenas HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth",
	})

	// Access token no body (o frontend guarda em memória, nunca em localStorage)
	respondJSON(w, http.StatusOK, map[string]any{
		"access_token":       pair.AccessToken,
		"token_type":         pair.TokenType,
		"expires_in":         int(AccessTokenTTL.Seconds()),
		"access_expires_at":  pair.AccessExpiresAt,
	})
}

// Refresh troca um refresh token por um novo par de tokens.
// POST /auth/refresh
// Lê o refresh_token do cookie HttpOnly.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Tentar cookie primeiro, depois header (para apps mobile)
	refreshToken := ""
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = cookie.Value
	}
	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		refreshToken = body.RefreshToken
	}

	if refreshToken == "" {
		respondError(w, http.StatusUnauthorized, "refresh_token não fornecido")
		return
	}

	pair, err := h.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		// Limpar cookie inválido
		http.SetCookie(w, &http.Cookie{
			Name:    "refresh_token",
			Value:   "",
			Expires: time.Unix(0, 0),
			Path:    "/auth",
		})
		respondError(w, http.StatusUnauthorized, "refresh token inválido ou expirado — faça login novamente")
		return
	}

	// Novo cookie com o refresh token rotacionado
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		Expires:  pair.RefreshExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth",
	})

	respondJSON(w, http.StatusOK, map[string]any{
		"access_token":      pair.AccessToken,
		"token_type":        pair.TokenType,
		"expires_in":        int(AccessTokenTTL.Seconds()),
		"access_expires_at": pair.AccessExpiresAt,
	})
}

// Logout invalida o refresh token do usuário.
// POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}

	// Limpar cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   "",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
		Path:    "/auth",
	})

	respondJSON(w, http.StatusOK, map[string]string{"message": "logout realizado com sucesso"})
}

// Me retorna os dados do usuário autenticado.
// GET /auth/me
// Requer: Authorization: Bearer <access_token>
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		respondError(w, http.StatusUnauthorized, "token não fornecido")
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := h.service.ValidateAccessToken(tokenStr)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "token inválido")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"user_id":       claims.UserID,
		"tenant_id":     claims.TenantID,
		"tenant_slug":   claims.TenantSlug,
		"business_type": claims.BusinessType,
		"role":          claims.Role,
		"permissions":   claims.Permissions,
		"expires_at":    claims.ExpiresAt,
	})
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
