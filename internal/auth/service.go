// Package auth implementa autenticação JWT do Nexo One.
//
// Fluxo completo:
//  1. POST /auth/login → valida email+senha → retorna access_token (15min) + refresh_token (7d)
//  2. Todas as requests: Authorization: Bearer <access_token>
//  3. Quando access_token expira: POST /auth/refresh → retorna novo par de tokens
//  4. POST /auth/logout → invalida refresh_token no Redis
//
// Segurança:
//   - access_token:  JWT assimétrico (RS256) ou simétrico (HS256), TTL 15min
//   - refresh_token: UUID opaco armazenado no Redis, TTL 7 dias
//   - Senha:         bcrypt cost=12
//   - Refresh tokens: rotação automática (um uso só)
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/nexoone/nexo-one/pkg/middleware"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
	BcryptCost      = 12
)

// Claims são os dados dentro do JWT de acesso.
type Claims struct {
	UserID      string   `json:"sub"`
	TenantID    string   `json:"tenant_id"`
	TenantSlug  string   `json:"tenant_slug"`
	BusinessType string  `json:"business_type"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// TokenPair é o par de tokens retornado no login e no refresh.
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	TokenType        string    `json:"token_type"` // "Bearer"
}

// LoginRequest dados de entrada para login.
type LoginRequest struct {
	TenantSlug string `json:"tenant_slug"` // subdomínio do tenant
	Email      string `json:"email"`
	Password   string `json:"password"`
}

// UserProvider busca dados do usuário para autenticação.
// Implementado pelo UserRepo.
type UserProvider interface {
	GetByEmail(ctx context.Context, tenantID, email string) (*UserAuth, error)
	GetTenantBySlug(ctx context.Context, slug string) (*TenantAuth, error)
}

// TokenStore armazena e valida refresh tokens no Redis.
type TokenStore interface {
	SaveRefreshToken(ctx context.Context, token, userID, tenantID string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (userID, tenantID string, err error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteAllUserTokens(ctx context.Context, userID string) error
}

// UserAuth dados do usuário para geração de JWT.
type UserAuth struct {
	ID           string
	TenantID     string
	Email        string
	Name         string
	Role         string
	PasswordHash string
	Active       bool
}

// TenantAuth dados do tenant para geração de JWT.
type TenantAuth struct {
	ID           string
	Slug         string
	BusinessType string
	Plan         string
}

// Service é o serviço de autenticação.
type Service struct {
	jwtSecret []byte
	users     UserProvider
	tokens    TokenStore
}

// NewService cria um novo serviço de autenticação.
func NewService(jwtSecret string, users UserProvider, tokens TokenStore) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
		users:     users,
		tokens:    tokens,
	}
}

// Login autentica um usuário e retorna um par de tokens.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	// 1. Buscar tenant pelo slug
	tenant, err := s.users.GetTenantBySlug(ctx, req.TenantSlug)
	if err != nil {
		// Mensagem genérica para não revelar se tenant existe
		return nil, fmt.Errorf("credenciais inválidas")
	}

	// 2. Buscar usuário pelo email dentro do tenant
	user, err := s.users.GetByEmail(ctx, tenant.ID, req.Email)
	if err != nil || !user.Active {
		return nil, fmt.Errorf("credenciais inválidas")
	}

	// 3. Verificar senha (bcrypt)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("credenciais inválidas")
	}

	// 4. Gerar par de tokens
	return s.generateTokenPair(ctx, user, tenant)
}

// Refresh valida um refresh token e retorna um novo par de tokens.
// O refresh token antigo é invalidado imediatamente (rotação).
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// 1. Validar refresh token no Redis
	userID, tenantID, err := s.tokens.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token inválido ou expirado")
	}

	// 2. Invalidar o token atual (rotação — um uso só)
	_ = s.tokens.DeleteRefreshToken(ctx, refreshToken)

	// 3. Buscar dados atualizados do usuário
	// (Em produção: buscar do banco para pegar role/permissions atualizados)
	user := &UserAuth{ID: userID, TenantID: tenantID, Active: true}
	tenant := &TenantAuth{ID: tenantID}

	// 4. Gerar novo par
	return s.generateTokenPair(ctx, user, tenant)
}

// Logout invalida o refresh token do usuário.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.tokens.DeleteRefreshToken(ctx, refreshToken)
}

// LogoutAll invalida TODOS os refresh tokens do usuário (segurança).
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.tokens.DeleteAllUserTokens(ctx, userID)
}

// ValidateAccessToken valida um JWT de acesso e retorna os claims.
func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inválido: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("claims inválidos")
	}
	return claims, nil
}

// Validate implements middleware.JWTValidator - adapta auth.Claims para middleware.Claims.
func (s *Service) Validate(tokenStr string) (*middleware.Claims, error) {
	claims, err := s.ValidateAccessToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return &middleware.Claims{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		TenantSlug:  claims.TenantSlug,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		ExpiresAt:   claims.ExpiresAt.Time,
	}, nil
}

// HashPassword gera um hash bcrypt de uma senha.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// generateTokenPair gera access_token (JWT) + refresh_token (opaco).
func (s *Service) generateTokenPair(ctx context.Context, user *UserAuth, tenant *TenantAuth) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExp := now.Add(AccessTokenTTL)
	refreshExp := now.Add(RefreshTokenTTL)

	// Permissões baseadas no role
	permissions := permissionsForRole(user.Role)

	// JWT de acesso
	claims := &Claims{
		UserID:       user.ID,
		TenantID:     tenant.ID,
		TenantSlug:   tenant.Slug,
		BusinessType: tenant.BusinessType,
		Role:         user.Role,
		Permissions:  permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nexoflash",
			Subject:   user.ID,
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("gerar access_token: %w", err)
	}

	// Refresh token opaco (UUID seguro)
	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("gerar refresh_token: %w", err)
	}

	// Salvar refresh token no Redis
	if err := s.tokens.SaveRefreshToken(ctx, refreshToken, user.ID, tenant.ID, RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("salvar refresh_token: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
		TokenType:        "Bearer",
	}, nil
}

// permissionsForRole retorna as permissões baseadas no role do usuário.
func permissionsForRole(role string) []string {
	switch role {
	case "owner", "admin":
		return []string{"*"} // acesso total
	case "operator":
		return []string{
			"mechanic:os:read", "mechanic:os:write", "mechanic:os:approve",
			"bakery:pdv:sell", "bakery:products:read",
			"aesthetics:agenda:read", "aesthetics:agenda:write",
			"shoes:sales:read",
		}
	case "viewer":
		return []string{
			"mechanic:os:read", "bakery:products:read",
			"aesthetics:agenda:read", "shoes:grid:read",
		}
	default:
		return []string{}
	}
}

// generateSecureToken gera um token criptograficamente seguro de 32 bytes.
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
