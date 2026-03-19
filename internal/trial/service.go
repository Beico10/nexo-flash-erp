// Package trial implementa o controle de trial e verificação por WhatsApp.
//
// Fluxo:
//  1. Usuário cadastra com telefone
//  2. Sistema gera código 6 dígitos, salva no Redis (5 min TTL)
//  3. Abre WhatsApp com mensagem pré-preenchida
//  4. Usuário envia → Webhook recebe → Valida código
//  5. Trial liberado
//
// Anti-abuso:
//  - 1 trial por telefone (hash SHA256 para LGPD)
//  - Device fingerprint para detectar tentativas de burlar
//  - Score de abuso incrementado a cada suspeita
package trial

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"
)

const (
	CodeLength     = 6
	CodeTTL        = 5 * time.Minute
	WhatsAppNumber = "5511999999999" // Número do Nexo One para receber códigos
)

var (
	ErrPhoneAlreadyUsed   = errors.New("este telefone já possui uma conta")
	ErrCodeExpired        = errors.New("código expirado, solicite um novo")
	ErrCodeInvalid        = errors.New("código inválido")
	ErrTooManyAttempts    = errors.New("muitas tentativas, aguarde 15 minutos")
	ErrBlocked            = errors.New("conta bloqueada por atividade suspeita")
)

// TrialControl representa o controle de trial de um usuário.
type TrialControl struct {
	ID                 string
	PhoneNumber        string
	PhoneHash          string
	Email              string
	CNPJ               string
	VerificationCode   string
	CodeExpiresAt      *time.Time
	VerifiedAt         *time.Time
	DeviceHash         string
	IPAddress          string
	TenantID           string
	IsBlocked          bool
	AbuseScore         int
	CreatedAt          time.Time
}

// TrialRepository interface de persistência.
type TrialRepository interface {
	GetByPhoneHash(ctx context.Context, hash string) (*TrialControl, error)
	GetByDeviceHash(ctx context.Context, hash string, since time.Time) ([]*TrialControl, error)
	Create(ctx context.Context, tc *TrialControl) error
	Update(ctx context.Context, tc *TrialControl) error
	SaveCode(ctx context.Context, phoneHash, code string, ttl time.Duration) error
	GetCode(ctx context.Context, phoneHash string) (string, error)
	IncrementAttempts(ctx context.Context, phoneHash string) (int, error)
}

// Service gerencia trials e verificação.
type Service struct {
	repo         TrialRepository
	whatsappNum  string
}

func NewService(repo TrialRepository) *Service {
	return &Service{
		repo:        repo,
		whatsappNum: WhatsAppNumber,
	}
}

// ════════════════════════════════════════════════════════════
// VERIFICAÇÃO POR WHATSAPP
// ════════════════════════════════════════════════════════════

// StartVerification inicia o processo de verificação por WhatsApp.
// Retorna a URL do WhatsApp com a mensagem pré-preenchida.
func (s *Service) StartVerification(ctx context.Context, phone, email, deviceHash, ip string) (string, error) {
	phoneHash := hashPhone(phone)

	// Verificar se telefone já foi usado
	existing, _ := s.repo.GetByPhoneHash(ctx, phoneHash)
	if existing != nil {
		if existing.VerifiedAt != nil {
			return "", ErrPhoneAlreadyUsed
		}
		// Já iniciou mas não verificou - permite reenviar
	}

	// Verificar device fingerprint (detectar abuso)
	abuseScore := 0
	recentFromDevice, _ := s.repo.GetByDeviceHash(ctx, deviceHash, time.Now().Add(-30*24*time.Hour))
	if len(recentFromDevice) > 2 {
		abuseScore = 50 // Suspeito: mesmo device, múltiplos trials
	}

	// Gerar código de 6 dígitos
	code := generateCode(CodeLength)

	// Salvar no Redis com TTL
	if err := s.repo.SaveCode(ctx, phoneHash, code, CodeTTL); err != nil {
		return "", fmt.Errorf("trial.StartVerification: %w", err)
	}

	// Criar ou atualizar controle
	tc := &TrialControl{
		PhoneNumber:  phone,
		PhoneHash:    phoneHash,
		Email:        email,
		DeviceHash:   deviceHash,
		IPAddress:    ip,
		AbuseScore:   abuseScore,
	}

	if existing == nil {
		if err := s.repo.Create(ctx, tc); err != nil {
			return "", err
		}
	} else {
		tc.ID = existing.ID
		if err := s.repo.Update(ctx, tc); err != nil {
			return "", err
		}
	}

	// Gerar URL do WhatsApp
	whatsappURL := s.buildWhatsAppURL(code)

	return whatsappURL, nil
}

// VerifyCode valida o código recebido via WhatsApp.
func (s *Service) VerifyCode(ctx context.Context, phone, code string) (*TrialControl, error) {
	phoneHash := hashPhone(phone)

	// Verificar tentativas (rate limit)
	attempts, _ := s.repo.IncrementAttempts(ctx, phoneHash)
	if attempts > 5 {
		return nil, ErrTooManyAttempts
	}

	// Buscar código salvo
	savedCode, err := s.repo.GetCode(ctx, phoneHash)
	if err != nil || savedCode == "" {
		return nil, ErrCodeExpired
	}

	// Validar
	if savedCode != code {
		return nil, ErrCodeInvalid
	}

	// Buscar controle
	tc, err := s.repo.GetByPhoneHash(ctx, phoneHash)
	if err != nil {
		return nil, err
	}

	if tc.IsBlocked {
		return nil, ErrBlocked
	}

	// Marcar como verificado
	now := time.Now()
	tc.VerifiedAt = &now

	if err := s.repo.Update(ctx, tc); err != nil {
		return nil, err
	}

	return tc, nil
}

// buildWhatsAppURL cria a URL do WhatsApp com mensagem pré-preenchida.
func (s *Service) buildWhatsAppURL(code string) string {
	message := fmt.Sprintf("Meu código Nexo One: %s", code)
	encoded := url.QueryEscape(message)
	return fmt.Sprintf("https://wa.me/%s?text=%s", s.whatsappNum, encoded)
}

// ════════════════════════════════════════════════════════════
// WEBHOOK WHATSAPP (Recebe mensagem do usuário)
// ════════════════════════════════════════════════════════════

// ProcessWhatsAppMessage processa mensagem recebida via webhook.
// Extrai o código e valida automaticamente.
func (s *Service) ProcessWhatsAppMessage(ctx context.Context, from, body string) (*TrialControl, error) {
	// Extrair código da mensagem
	code := extractCode(body)
	if code == "" {
		return nil, ErrCodeInvalid
	}

	// Verificar
	return s.VerifyCode(ctx, from, code)
}

// ════════════════════════════════════════════════════════════
// ANTI-ABUSO
// ════════════════════════════════════════════════════════════

// CheckAbuse verifica se há sinais de abuso.
func (s *Service) CheckAbuse(ctx context.Context, deviceHash, ip string) (int, []string) {
	var warnings []string
	score := 0

	// Verificar múltiplos trials do mesmo device
	recent, _ := s.repo.GetByDeviceHash(ctx, deviceHash, time.Now().Add(-30*24*time.Hour))
	if len(recent) > 1 {
		score += 30
		warnings = append(warnings, fmt.Sprintf("%d trials do mesmo dispositivo em 30 dias", len(recent)))
	}
	if len(recent) > 3 {
		score += 40
		warnings = append(warnings, "Possível abuso: muitos trials do mesmo dispositivo")
	}

	return score, warnings
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func hashPhone(phone string) string {
	h := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(h[:])
}

func generateCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

func extractCode(message string) string {
	// Procura por 6 dígitos consecutivos
	var code string
	for _, r := range message {
		if r >= '0' && r <= '9' {
			code += string(r)
			if len(code) == CodeLength {
				return code
			}
		} else {
			code = ""
		}
	}
	return ""
}
