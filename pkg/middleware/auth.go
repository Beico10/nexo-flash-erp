// Package middleware implementa os middlewares HTTP do Nexo One.
//
// TenantMiddleware é o middleware mais crítico do sistema:
// ele extrai o tenant_id do JWT e configura a sessão PostgreSQL
// com SET LOCAL app.tenant_id, ativando o Row Level Security.
//
// Sem isso, qualquer query retornaria zero rows (RLS bloqueia tudo).
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Claims representa o payload do JWT do Nexo One.
type Claims struct {
	UserID       string   `json:"sub"`
	TenantID     string   `json:"tenant_id"`
	TenantSlug   string   `json:"tenant_slug"`
	BusinessType string   `json:"business_type"`
	Role         string   `json:"role"`
	Permissions  []string `json:"permissions"`
	ExpiresAt    time.Time `json:"exp"`
}

// contextKey para armazenar claims no contexto HTTP.
type contextKey string

const (
	claimsKey    contextKey = "claims"
	tenantIDKey  contextKey = "tenant_id"
)

// JWTValidator valida e decodifica tokens JWT.
type JWTValidator interface {
	Validate(token string) (*Claims, error)
}

// DBSessionSetter configura a sessão do banco para o tenant atual.
// Implementação real: SET LOCAL app.tenant_id = '<uuid>'
type DBSessionSetter interface {
	SetTenantSession(ctx context.Context, tenantID string) error
}

// AuthMiddleware valida o JWT e injeta as claims no contexto.
// Deve ser o primeiro middleware na cadeia.
func AuthMiddleware(validator JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				http.Error(w, `{"error":"token não fornecido"}`, http.StatusUnauthorized)
				return
			}

			claims, err := validator.Validate(token)
			if err != nil {
				http.Error(w, `{"error":"token inválido"}`, http.StatusUnauthorized)
				return
			}

			if claims.ExpiresAt.Before(time.Now()) {
				http.Error(w, `{"error":"token expirado"}`, http.StatusUnauthorized)
				return
			}

			// Injeta claims no contexto para uso nos handlers
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantDBMiddleware configura o Row Level Security no banco para cada requisição.
// DEVE ser executado após AuthMiddleware — depende das claims no contexto.
// Cada request obtém uma conexão do pool e executa SET LOCAL app.tenant_id.
func TenantDBMiddleware(setter DBSessionSetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsKey).(*Claims)
			if !ok || claims.TenantID == "" {
				http.Error(w, `{"error":"tenant não identificado"}`, http.StatusForbidden)
				return
			}

			// Esta chamada executa: SET LOCAL app.tenant_id = '<uuid>'
			// O PostgreSQL RLS usa current_setting('app.tenant_id') em todas as policies.
			if err := setter.SetTenantSession(r.Context(), claims.TenantID); err != nil {
				http.Error(w, `{"error":"erro interno de sessão"}`, http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission verifica se o usuário tem uma permissão específica.
// Uso: router.Use(RequirePermission("mechanic:os:write"))
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsKey).(*Claims)
			if !ok {
				http.Error(w, `{"error":"não autenticado"}`, http.StatusUnauthorized)
				return
			}

			for _, p := range claims.Permissions {
				if p == permission || p == "*" {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w,
				fmt.Sprintf(`{"error":"sem permissão: %s"}`, permission),
				http.StatusForbidden)
		})
	}
}

// GetClaims extrai as claims do contexto da requisição.
func GetClaims(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

// GetTenantID extrai o tenant_id do contexto.
func GetTenantID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(tenantIDKey).(string)
	return id, ok
}

// extractBearerToken extrai o token do header Authorization: Bearer <token>
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

// WebAuthMiddleware valida o JWT para páginas web.
// Tenta extrair o token do header ou do query param (para redirects).
// Se não encontrar token válido, redireciona para /login.
func WebAuthMiddleware(validator JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Tenta extrair token de várias fontes
			token := extractBearerToken(r)
			
			// Também aceita via query param para casos especiais
			if token == "" {
				token = r.URL.Query().Get("token")
			}
			
			if token == "" {
				// Para requisições AJAX/fetch, retorna 401
				if r.Header.Get("Accept") == "application/json" {
					http.Error(w, `{"error":"nao autenticado"}`, http.StatusUnauthorized)
					return
				}
				// Para navegação normal, redireciona para login
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			claims, err := validator.Validate(token)
			if err != nil || claims.ExpiresAt.Before(time.Now()) {
				if r.Header.Get("Accept") == "application/json" {
					http.Error(w, `{"error":"token invalido"}`, http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Injeta claims no contexto
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
