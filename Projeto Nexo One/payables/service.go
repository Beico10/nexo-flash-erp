// Package payables implementa o módulo de Contas a Pagar do Nexo One.
//
// Funcionalidades:
//   - Cadastro de contas com vencimento e recorrência
//   - Baixa manual com forma de pagamento
//   - Alerta de vencimento via WhatsApp (D-3, D-1, D0)
//   - Parcelamento automático
//   - Fluxo de caixa projetado
//   - Relatório de inadimplência
package payables

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ── CONSTANTES ────────────────────────────────────────────────────────────────

const (
	StatusPending   = "pending"   // Aguardando pagamento
	StatusPaid      = "paid"      // Pago
	StatusOverdue   = "overdue"   // Vencido
	StatusCancelled = "cancelled" // Cancelado

	RecurrenceNone    = "none"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
	RecurrenceYearly  = "yearly"
)

// ── TIPOS ─────────────────────────────────────────────────────────────────────

// Payable representa uma conta a pagar.
type Payable struct {
	ID              string
	TenantID        string
	Description     string
	SupplierName    string
	SupplierCNPJ    string
	Category        string    // aluguel, fornecedor, imposto, folha, servico, outros
	Amount          float64
	AmountPaid      float64
	DueDate         time.Time
	PaidAt          *time.Time
	PaymentMethod   string    // pix, boleto, cartao, dinheiro, transferencia
	BankAccount     string
	Installment     int       // parcela atual (1 de 3)
	TotalInstallments int     // total de parcelas
	ParentID        string    // ID da conta original (parcelamento)
	Recurrence      string    // none, weekly, monthly, yearly
	Status          string
	Notes           string
	NFEKey          string    // chave NF-e vinculada
	CostCenter      string
	AlertSent       bool      // alerta de vencimento enviado
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PayableFilter filtros para listagem.
type PayableFilter struct {
	Status    string
	Category  string
	DueFrom   *time.Time
	DueTo     *time.Time
	Overdue   bool
	Limit     int
	Offset    int
}

// CashFlowItem item do fluxo de caixa projetado.
type CashFlowItem struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Cumulative  float64   `json:"cumulative"` // saldo acumulado
}

// PayablesSummary resumo de contas a pagar.
type PayablesSummary struct {
	TotalPending  float64 `json:"total_pending"`
	TotalOverdue  float64 `json:"total_overdue"`
	TotalPaidMonth float64 `json:"total_paid_month"`
	CountPending  int     `json:"count_pending"`
	CountOverdue  int     `json:"count_overdue"`
	NextDue       *Payable `json:"next_due"`
}

// ── ERROS ─────────────────────────────────────────────────────────────────────

var (
	ErrPayableNotFound    = errors.New("conta não encontrada")
	ErrAlreadyPaid        = errors.New("conta já foi paga")
	ErrInvalidAmount      = errors.New("valor inválido")
	ErrInvalidDueDate     = errors.New("data de vencimento inválida")
)

// ── REPOSITÓRIO ───────────────────────────────────────────────────────────────

// Repository interface de persistência.
type Repository interface {
	Create(ctx context.Context, p *Payable) error
	GetByID(ctx context.Context, tenantID, id string) (*Payable, error)
	List(ctx context.Context, tenantID string, filter PayableFilter) ([]*Payable, error)
	Update(ctx context.Context, p *Payable) error
	Delete(ctx context.Context, tenantID, id string) error
	GetOverdue(ctx context.Context, tenantID string) ([]*Payable, error)
	GetDueSoon(ctx context.Context, tenantID string, days int) ([]*Payable, error)
	GetSummary(ctx context.Context, tenantID string) (*PayablesSummary, error)
	GetCashFlow(ctx context.Context, tenantID string, from, to time.Time) ([]*CashFlowItem, error)
	MarkAlertSent(ctx context.Context, tenantID, id string) error
}

// WhatsAppNotifier interface para envio de alertas.
type WhatsAppNotifier interface {
	SendPayableAlert(ctx context.Context, phone, supplierName string, amount float64, dueDate time.Time, daysUntilDue int) error
}

// ── SERVIÇO ───────────────────────────────────────────────────────────────────

// Service gerencia contas a pagar.
type Service struct {
	repo      Repository
	notifier  WhatsAppNotifier
}

func NewService(repo Repository, notifier WhatsAppNotifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

// Create cria uma nova conta a pagar.
// Se TotalInstallments > 1, cria todas as parcelas automaticamente.
func (s *Service) Create(ctx context.Context, p *Payable) ([]*Payable, error) {
	if p.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if p.DueDate.IsZero() {
		return nil, ErrInvalidDueDate
	}

	p.Status = StatusPending
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	// Parcela única
	if p.TotalInstallments <= 1 {
		p.Installment = 1
		p.TotalInstallments = 1
		if err := s.repo.Create(ctx, p); err != nil {
			return nil, fmt.Errorf("payables.Create: %w", err)
		}
		return []*Payable{p}, nil
	}

	// Parcelamento — criar uma conta por parcela
	installmentAmount := p.Amount / float64(p.TotalInstallments)
	var created []*Payable
	baseDescription := p.Description

	for i := 1; i <= p.TotalInstallments; i++ {
		installment := &Payable{
			TenantID:          p.TenantID,
			Description:       fmt.Sprintf("%s (%d/%d)", baseDescription, i, p.TotalInstallments),
			SupplierName:      p.SupplierName,
			SupplierCNPJ:      p.SupplierCNPJ,
			Category:          p.Category,
			Amount:            installmentAmount,
			DueDate:           p.DueDate.AddDate(0, i-1, 0), // +1 mês por parcela
			Installment:       i,
			TotalInstallments: p.TotalInstallments,
			Recurrence:        RecurrenceNone,
			Status:            StatusPending,
			Notes:             p.Notes,
			CostCenter:        p.CostCenter,
			CreatedBy:         p.CreatedBy,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := s.repo.Create(ctx, installment); err != nil {
			return nil, fmt.Errorf("payables.Create parcela %d: %w", i, err)
		}
		created = append(created, installment)
	}

	return created, nil
}

// Pay realiza a baixa de uma conta.
func (s *Service) Pay(ctx context.Context, tenantID, id string, amountPaid float64, paymentMethod string, paidAt *time.Time) (*Payable, error) {
	p, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, ErrPayableNotFound
	}

	if p.Status == StatusPaid {
		return nil, ErrAlreadyPaid
	}
	if p.Status == StatusCancelled {
		return nil, errors.New("conta cancelada não pode ser paga")
	}

	if amountPaid <= 0 {
		amountPaid = p.Amount
	}

	now := time.Now()
	if paidAt == nil {
		paidAt = &now
	}

	p.AmountPaid = amountPaid
	p.PaidAt = paidAt
	p.PaymentMethod = paymentMethod
	p.Status = StatusPaid
	p.UpdatedAt = time.Now()

	// Se pagamento recorrente, criar próxima ocorrência
	if p.Recurrence != RecurrenceNone {
		s.createNextRecurrence(ctx, p)
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("payables.Pay: %w", err)
	}

	return p, nil
}

// createNextRecurrence cria a próxima ocorrência de uma conta recorrente.
func (s *Service) createNextRecurrence(ctx context.Context, p *Payable) {
	var nextDue time.Time
	switch p.Recurrence {
	case RecurrenceWeekly:
		nextDue = p.DueDate.AddDate(0, 0, 7)
	case RecurrenceMonthly:
		nextDue = p.DueDate.AddDate(0, 1, 0)
	case RecurrenceYearly:
		nextDue = p.DueDate.AddDate(1, 0, 0)
	default:
		return
	}

	next := &Payable{
		TenantID:     p.TenantID,
		Description:  p.Description,
		SupplierName: p.SupplierName,
		SupplierCNPJ: p.SupplierCNPJ,
		Category:     p.Category,
		Amount:       p.Amount,
		DueDate:      nextDue,
		Recurrence:   p.Recurrence,
		Status:       StatusPending,
		Notes:        p.Notes,
		CostCenter:   p.CostCenter,
		CreatedBy:    p.CreatedBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.repo.Create(ctx, next)
}

// Cancel cancela uma conta.
func (s *Service) Cancel(ctx context.Context, tenantID, id, reason string) error {
	p, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return ErrPayableNotFound
	}

	if p.Status == StatusPaid {
		return errors.New("conta já paga não pode ser cancelada")
	}

	p.Status = StatusCancelled
	p.Notes = fmt.Sprintf("%s [Cancelado: %s]", p.Notes, reason)
	p.UpdatedAt = time.Now()

	return s.repo.Update(ctx, p)
}

// UpdateStatus atualiza status de contas vencidas.
// Chamar via cron job diário.
func (s *Service) UpdateOverdueStatus(ctx context.Context, tenantID string) error {
	overdue, err := s.repo.GetOverdue(ctx, tenantID)
	if err != nil {
		return err
	}

	for _, p := range overdue {
		if p.Status == StatusPending && time.Now().After(p.DueDate) {
			p.Status = StatusOverdue
			p.UpdatedAt = time.Now()
			s.repo.Update(ctx, p)
		}
	}

	return nil
}

// ── LISTAGEM ──────────────────────────────────────────────────────────────────

func (s *Service) List(ctx context.Context, tenantID string, filter PayableFilter) ([]*Payable, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Payable, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *Service) GetSummary(ctx context.Context, tenantID string) (*PayablesSummary, error) {
	return s.repo.GetSummary(ctx, tenantID)
}

func (s *Service) GetCashFlow(ctx context.Context, tenantID string, from, to time.Time) ([]*CashFlowItem, error) {
	items, err := s.repo.GetCashFlow(ctx, tenantID, from, to)
	if err != nil {
		return nil, err
	}

	// Calcular saldo acumulado
	var cumulative float64
	for _, item := range items {
		cumulative -= item.Amount // despesa reduz saldo
		item.Cumulative = cumulative
	}

	return items, nil
}

// ── ALERTAS WHATSAPP ──────────────────────────────────────────────────────────

// SendDueAlerts envia alertas de vencimento via WhatsApp.
// Chamar via cron job diário às 8h.
func (s *Service) SendDueAlerts(ctx context.Context, tenantID, phone string) error {
	if s.notifier == nil {
		return nil
	}

	// Alertas D-3, D-1 e D0
	for _, days := range []int{3, 1, 0} {
		dueSoon, err := s.repo.GetDueSoon(ctx, tenantID, days)
		if err != nil {
			continue
		}

		for _, p := range dueSoon {
			if p.AlertSent {
				continue
			}

			if err := s.notifier.SendPayableAlert(
				ctx, phone,
				p.SupplierName, p.Amount, p.DueDate, days,
			); err == nil {
				s.repo.MarkAlertSent(ctx, tenantID, p.ID)
			}
		}
	}

	return nil
}
