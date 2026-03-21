// Package receivables implementa o módulo de Contas a Receber do Nexo One.
//
// Funcionalidades:
//   - Cadastro de contas a receber com vencimento
//   - Baixa com forma de recebimento
//   - Parcelamento automático
//   - Recorrência (mensalidades, contratos)
//   - Alerta de atraso via WhatsApp para o cliente
//   - Fluxo de caixa projetado
//   - Relatório de inadimplência
package receivables

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ── CONSTANTES ────────────────────────────────────────────────────────────────

const (
	StatusPending   = "pending"
	StatusReceived  = "received"
	StatusOverdue   = "overdue"
	StatusCancelled = "cancelled"
	StatusPartial   = "partial" // recebimento parcial

	RecurrenceNone    = "none"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
	RecurrenceYearly  = "yearly"
)

// ── TIPOS ─────────────────────────────────────────────────────────────────────

// Receivable representa uma conta a receber.
type Receivable struct {
	ID                string
	TenantID          string
	Description       string
	CustomerName      string
	CustomerPhone     string
	CustomerDocument  string // CPF ou CNPJ
	Category          string // servico, produto, mensalidade, contrato, outros
	Amount            float64
	AmountReceived    float64
	DueDate           time.Time
	ReceivedAt        *time.Time
	PaymentMethod     string
	Installment       int
	TotalInstallments int
	ParentID          string
	Recurrence        string
	Status            string
	Notes             string
	ReferenceID       string // ID da OS, venda, etc.
	ReferenceType     string // os, sale, contract
	AlertSent         bool
	AlertCount        int
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ReceivableFilter filtros para listagem.
type ReceivableFilter struct {
	Status   string
	Category string
	DueFrom  *time.Time
	DueTo    *time.Time
	Overdue  bool
	Customer string
	Limit    int
	Offset   int
}

// ReceivablesSummary resumo de contas a receber.
type ReceivablesSummary struct {
	TotalPending      float64     `json:"total_pending"`
	TotalOverdue      float64     `json:"total_overdue"`
	TotalReceivedMonth float64    `json:"total_received_month"`
	CountPending      int         `json:"count_pending"`
	CountOverdue      int         `json:"count_overdue"`
	NextDue           *Receivable `json:"next_due"`
	OverdueRate       float64     `json:"overdue_rate"` // % de inadimplência
}

// ── ERROS ─────────────────────────────────────────────────────────────────────

var (
	ErrNotFound       = errors.New("conta não encontrada")
	ErrAlreadyReceived = errors.New("conta já foi recebida")
	ErrInvalidAmount  = errors.New("valor inválido")
	ErrInvalidDueDate = errors.New("data de vencimento inválida")
)

// ── REPOSITÓRIO ───────────────────────────────────────────────────────────────

type Repository interface {
	Create(ctx context.Context, r *Receivable) error
	GetByID(ctx context.Context, tenantID, id string) (*Receivable, error)
	List(ctx context.Context, tenantID string, filter ReceivableFilter) ([]*Receivable, error)
	Update(ctx context.Context, r *Receivable) error
	Delete(ctx context.Context, tenantID, id string) error
	GetOverdue(ctx context.Context, tenantID string) ([]*Receivable, error)
	GetDueSoon(ctx context.Context, tenantID string, days int) ([]*Receivable, error)
	GetSummary(ctx context.Context, tenantID string) (*ReceivablesSummary, error)
}

// WhatsAppNotifier alerta de cobrança.
type WhatsAppNotifier interface {
	SendOverdueAlert(ctx context.Context, phone, customerName, description string, amount float64, daysOverdue int) error
	SendDueReminder(ctx context.Context, phone, customerName, description string, amount float64, daysUntilDue int) error
}

// ── SERVIÇO ───────────────────────────────────────────────────────────────────

type Service struct {
	repo     Repository
	notifier WhatsAppNotifier
}

func NewService(repo Repository, notifier WhatsAppNotifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

// Create cria conta a receber. Suporta parcelamento.
func (s *Service) Create(ctx context.Context, r *Receivable) ([]*Receivable, error) {
	if r.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if r.DueDate.IsZero() {
		return nil, ErrInvalidDueDate
	}

	r.Status = StatusPending
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	// Parcela única
	if r.TotalInstallments <= 1 {
		r.Installment = 1
		r.TotalInstallments = 1
		if err := s.repo.Create(ctx, r); err != nil {
			return nil, fmt.Errorf("receivables.Create: %w", err)
		}
		return []*Receivable{r}, nil
	}

	// Parcelamento
	installmentAmount := r.Amount / float64(r.TotalInstallments)
	baseDescription := r.Description
	var created []*Receivable

	for i := 1; i <= r.TotalInstallments; i++ {
		installment := &Receivable{
			TenantID:          r.TenantID,
			Description:       fmt.Sprintf("%s (%d/%d)", baseDescription, i, r.TotalInstallments),
			CustomerName:      r.CustomerName,
			CustomerPhone:     r.CustomerPhone,
			CustomerDocument:  r.CustomerDocument,
			Category:          r.Category,
			Amount:            installmentAmount,
			DueDate:           r.DueDate.AddDate(0, i-1, 0),
			Installment:       i,
			TotalInstallments: r.TotalInstallments,
			Recurrence:        RecurrenceNone,
			Status:            StatusPending,
			Notes:             r.Notes,
			ReferenceID:       r.ReferenceID,
			ReferenceType:     r.ReferenceType,
			CreatedBy:         r.CreatedBy,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := s.repo.Create(ctx, installment); err != nil {
			return nil, fmt.Errorf("receivables.Create parcela %d: %w", i, err)
		}
		created = append(created, installment)
	}

	return created, nil
}

// Receive registra o recebimento de uma conta.
func (s *Service) Receive(ctx context.Context, tenantID, id string, amountReceived float64, paymentMethod string, receivedAt *time.Time) (*Receivable, error) {
	r, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}

	if r.Status == StatusReceived {
		return nil, ErrAlreadyReceived
	}
	if r.Status == StatusCancelled {
		return nil, errors.New("conta cancelada")
	}

	if amountReceived <= 0 {
		amountReceived = r.Amount
	}

	now := time.Now()
	if receivedAt == nil {
		receivedAt = &now
	}

	r.AmountReceived = amountReceived
	r.ReceivedAt = receivedAt
	r.PaymentMethod = paymentMethod
	r.UpdatedAt = time.Now()

	// Recebimento parcial ou total
	if amountReceived >= r.Amount {
		r.Status = StatusReceived
	} else {
		r.Status = StatusPartial
	}

	// Se recorrente, criar próxima
	if r.Recurrence != RecurrenceNone {
		s.createNextRecurrence(ctx, r)
	}

	if err := s.repo.Update(ctx, r); err != nil {
		return nil, fmt.Errorf("receivables.Receive: %w", err)
	}

	return r, nil
}

func (s *Service) createNextRecurrence(ctx context.Context, r *Receivable) {
	var nextDue time.Time
	switch r.Recurrence {
	case RecurrenceWeekly:
		nextDue = r.DueDate.AddDate(0, 0, 7)
	case RecurrenceMonthly:
		nextDue = r.DueDate.AddDate(0, 1, 0)
	case RecurrenceYearly:
		nextDue = r.DueDate.AddDate(1, 0, 0)
	default:
		return
	}

	next := &Receivable{
		TenantID:         r.TenantID,
		Description:      r.Description,
		CustomerName:     r.CustomerName,
		CustomerPhone:    r.CustomerPhone,
		CustomerDocument: r.CustomerDocument,
		Category:         r.Category,
		Amount:           r.Amount,
		DueDate:          nextDue,
		Recurrence:       r.Recurrence,
		Status:           StatusPending,
		Notes:            r.Notes,
		ReferenceID:      r.ReferenceID,
		ReferenceType:    r.ReferenceType,
		CreatedBy:        r.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	s.repo.Create(ctx, next)
}

// Cancel cancela uma conta.
func (s *Service) Cancel(ctx context.Context, tenantID, id, reason string) error {
	r, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return ErrNotFound
	}
	if r.Status == StatusReceived {
		return errors.New("conta já recebida não pode ser cancelada")
	}
	r.Status = StatusCancelled
	r.Notes = fmt.Sprintf("%s [Cancelado: %s]", r.Notes, reason)
	r.UpdatedAt = time.Now()
	return s.repo.Update(ctx, r)
}

// SendOverdueAlerts envia cobranças via WhatsApp para clientes em atraso.
func (s *Service) SendOverdueAlerts(ctx context.Context, tenantID string) error {
	if s.notifier == nil {
		return nil
	}

	overdue, err := s.repo.GetOverdue(ctx, tenantID)
	if err != nil {
		return err
	}

	for _, r := range overdue {
		if r.CustomerPhone == "" || r.AlertCount >= 3 {
			continue
		}
		daysOverdue := int(time.Since(r.DueDate).Hours() / 24)
		if err := s.notifier.SendOverdueAlert(
			ctx, r.CustomerPhone, r.CustomerName,
			r.Description, r.Amount, daysOverdue,
		); err == nil {
			r.AlertCount++
			r.AlertSent = true
			s.repo.Update(ctx, r)
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context, tenantID string, filter ReceivableFilter) ([]*Receivable, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Receivable, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *Service) GetSummary(ctx context.Context, tenantID string) (*ReceivablesSummary, error) {
	return s.repo.GetSummary(ctx, tenantID)
}
