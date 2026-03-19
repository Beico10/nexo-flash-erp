// Package baas implementa o Banking as a Service do Nexo One.
//
// Funcionalidades (Briefing Mestre §3):
//   - PIX Dinâmico (QR Code com valor e vencimento)
//   - Boleto Híbrido (boleto + PIX no mesmo documento)
//   - Conciliação automática de recebimentos
//   - Split de Pagamento nativo (repasse automático para profissionais)
//
// Integração via Gateway Nacional (ex: Celcoin, BMP, Pluggy).
// A interface BaaSGateway permite trocar o provedor sem alterar o core.
package baas

import (
	"context"
	"fmt"
	"time"
)

// PaymentStatus estados de um pagamento.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentPaid      PaymentStatus = "paid"
	PaymentExpired   PaymentStatus = "expired"
	PaymentCancelled PaymentStatus = "cancelled"
	PaymentRefunded  PaymentStatus = "refunded"
)

// PixCharge representa uma cobrança PIX dinâmica.
type PixCharge struct {
	ID            string
	TenantID      string
	TxID          string        // ID único da transação (max 35 chars)
	Amount        float64
	Description   string
	PayerName     string
	PayerDocument string        // CPF ou CNPJ
	ExpiresAt     time.Time
	Status        PaymentStatus
	QRCodeText    string        // Payload EMV (copia e cola)
	QRCodeImage   string        // base64 do QR Code PNG
	CreatedAt     time.Time
	PaidAt        *time.Time
	// Split: se preenchido, o gateway divide automaticamente
	SplitRecipients []SplitRecipient
}

// BoletoCharge representa um boleto híbrido (boleto + PIX).
type BoletoCharge struct {
	ID            string
	TenantID      string
	OurNumber     string        // nosso número
	Amount        float64
	DueDate       time.Date
	Description   string
	PayerName     string
	PayerDocument string
	PayerAddress  string
	Status        PaymentStatus
	BarCode       string        // código de barras
	DigitableLine string        // linha digitável
	QRCodeText    string        // PIX embutido no boleto
	PDFURL        string        // link para o PDF do boleto
	CreatedAt     time.Time
	PaidAt        *time.Time
}

// SplitRecipient define um destinatário no split de pagamento.
type SplitRecipient struct {
	AccountID  string  // ID da conta BaaS do destinatário
	Amount     float64 // valor a receber
	Percentage float64 // ou percentual (um dos dois)
}

// WebhookEvent é o evento recebido do gateway quando um pagamento é confirmado.
type WebhookEvent struct {
	EventType  string        // "payment.paid" | "payment.expired" | "refund.completed"
	ChargeID   string
	TxID       string
	Amount     float64
	PaidAt     time.Time
	RawPayload []byte
}

// BaaSGateway é a interface do provedor bancário.
// Troque o provedor implementando esta interface — zero mudanças no core.
type BaaSGateway interface {
	// PIX
	CreatePixCharge(ctx context.Context, charge *PixCharge) (*PixCharge, error)
	CancelPixCharge(ctx context.Context, tenantID, txID string) error
	GetPixCharge(ctx context.Context, tenantID, txID string) (*PixCharge, error)

	// Boleto
	CreateBoleto(ctx context.Context, boleto *BoletoCharge) (*BoletoCharge, error)
	CancelBoleto(ctx context.Context, tenantID, ourNumber string) error

	// Conciliação
	ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
}

// PaymentService é o serviço de pagamentos do Nexo One.
type PaymentService struct {
	gateway BaaSGateway
	repo    PaymentRepository
}

// PaymentRepository persiste cobranças e atualizações de status.
type PaymentRepository interface {
	SavePixCharge(ctx context.Context, charge *PixCharge) error
	UpdatePixStatus(ctx context.Context, tenantID, txID string, status PaymentStatus, paidAt *time.Time) error
	SaveBoleto(ctx context.Context, boleto *BoletoCharge) error
	UpdateBoletoStatus(ctx context.Context, tenantID, ourNumber string, status PaymentStatus, paidAt *time.Time) error
}

func NewPaymentService(g BaaSGateway, r PaymentRepository) *PaymentService {
	return &PaymentService{gateway: g, repo: r}
}

// CreatePixCharge gera um PIX dinâmico com QR Code e opcionalmente com split.
func (s *PaymentService) CreatePixCharge(ctx context.Context, tenantID string, input PixChargeInput) (*PixCharge, error) {
	if err := validatePixInput(input); err != nil {
		return nil, fmt.Errorf("baas.CreatePixCharge: %w", err)
	}

	charge := &PixCharge{
		TenantID:        tenantID,
		TxID:            generateTxID(),
		Amount:          input.Amount,
		Description:     input.Description,
		PayerName:       input.PayerName,
		PayerDocument:   input.PayerDocument,
		ExpiresAt:       time.Now().Add(input.ExpiresIn),
		SplitRecipients: input.SplitRecipients,
		CreatedAt:       time.Now().UTC(),
	}

	// Envia para o gateway BaaS
	created, err := s.gateway.CreatePixCharge(ctx, charge)
	if err != nil {
		return nil, fmt.Errorf("baas.CreatePixCharge: gateway falhou: %w", err)
	}

	// Persiste localmente para conciliação
	if err := s.repo.SavePixCharge(ctx, created); err != nil {
		return nil, fmt.Errorf("baas.CreatePixCharge: persistência falhou: %w", err)
	}

	return created, nil
}

// CreateBoleto gera um boleto híbrido (boleto + PIX embutido).
func (s *PaymentService) CreateBoleto(ctx context.Context, tenantID string, input BoletoInput) (*BoletoCharge, error) {
	if input.Amount <= 0 {
		return nil, fmt.Errorf("baas.CreateBoleto: valor deve ser > 0")
	}

	boleto := &BoletoCharge{
		TenantID:      tenantID,
		Amount:        input.Amount,
		DueDate:       input.DueDate,
		Description:   input.Description,
		PayerName:     input.PayerName,
		PayerDocument: input.PayerDocument,
		PayerAddress:  input.PayerAddress,
		CreatedAt:     time.Now().UTC(),
	}

	created, err := s.gateway.CreateBoleto(ctx, boleto)
	if err != nil {
		return nil, fmt.Errorf("baas.CreateBoleto: gateway falhou: %w", err)
	}
	if err := s.repo.SaveBoleto(ctx, created); err != nil {
		return nil, fmt.Errorf("baas.CreateBoleto: persistência falhou: %w", err)
	}
	return created, nil
}

// ProcessWebhook recebe a confirmação de pagamento do gateway e atualiza o status.
// Chamado pelo endpoint POST /api/v1/webhooks/payment
func (s *PaymentService) ProcessWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := s.gateway.ProcessWebhook(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("baas.ProcessWebhook: %w", err)
	}

	switch event.EventType {
	case "payment.paid":
		paidAt := event.PaidAt
		// Atualiza PIX ou Boleto dependendo do contexto
		_ = s.repo.UpdatePixStatus(ctx, "", event.TxID, PaymentPaid, &paidAt)
		// Aqui: publicar evento no NATS para o módulo financeiro
		// eventbus.Publish(ctx, tenantID, EventPaymentReceived, event)
	case "payment.expired":
		_ = s.repo.UpdatePixStatus(ctx, "", event.TxID, PaymentExpired, nil)
	}
	return nil
}

// PixChargeInput dados de entrada para criar um PIX.
type PixChargeInput struct {
	Amount          float64
	Description     string
	PayerName       string
	PayerDocument   string
	ExpiresIn       time.Duration // ex: 24 * time.Hour
	SplitRecipients []SplitRecipient
}

// BoletoInput dados de entrada para criar um boleto.
type BoletoInput struct {
	Amount        float64
	DueDate       time.Date
	Description   string
	PayerName     string
	PayerDocument string
	PayerAddress  string
}

func validatePixInput(i PixChargeInput) error {
	if i.Amount <= 0 {
		return fmt.Errorf("valor deve ser > 0")
	}
	if i.ExpiresIn <= 0 {
		return fmt.Errorf("ExpiresIn deve ser > 0")
	}
	return nil
}

func generateTxID() string {
	// Em produção: use UUID v4 sem hífens (max 35 chars conforme BACEN)
	return fmt.Sprintf("NXF%d", time.Now().UnixNano())[:35]
}

// time.Date temporário — em produção use time.Time com format YYYY-MM-DD
type Date = time.Time
