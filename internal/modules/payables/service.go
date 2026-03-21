// Package payables - Contas a Pagar do Nexo One ERP
package payables

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Payable representa uma conta a pagar
type Payable struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Description   string    `json:"description"`
	Supplier      string    `json:"supplier"`
	Category      string    `json:"category"`
	Amount        float64   `json:"amount"`
	DueDate       time.Time `json:"due_date"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	Status        string    `json:"status"` // pending, paid, overdue, cancelled
	PaymentMethod string    `json:"payment_method,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Repository interface para contas a pagar
type Repository interface {
	Create(ctx context.Context, p *Payable) error
	GetByID(ctx context.Context, tenantID, id string) (*Payable, error)
	List(ctx context.Context, tenantID string, filter Filter) ([]Payable, error)
	Update(ctx context.Context, p *Payable) error
	Delete(ctx context.Context, tenantID, id string) error
	GetSummary(ctx context.Context, tenantID string) (*Summary, error)
}

// Filter para listagem
type Filter struct {
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	Supplier  string
	Category  string
}

// Summary resumo de contas a pagar
type Summary struct {
	TotalPending  float64 `json:"total_pending"`
	TotalPaid     float64 `json:"total_paid"`
	TotalOverdue  float64 `json:"total_overdue"`
	CountPending  int     `json:"count_pending"`
	CountPaid     int     `json:"count_paid"`
	CountOverdue  int     `json:"count_overdue"`
}

// Service para contas a pagar
type Service struct {
	repo Repository
}

// NewService cria um novo service de contas a pagar
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create cria uma nova conta a pagar
func (s *Service) Create(ctx context.Context, p *Payable) error {
	p.ID = uuid.New().String()
	p.Status = "pending"
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	
	if p.DueDate.Before(time.Now()) {
		p.Status = "overdue"
	}
	
	return s.repo.Create(ctx, p)
}

// GetByID busca uma conta por ID
func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Payable, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// List lista contas a pagar
func (s *Service) List(ctx context.Context, tenantID string, filter Filter) ([]Payable, error) {
	return s.repo.List(ctx, tenantID, filter)
}

// MarkAsPaid marca uma conta como paga
func (s *Service) MarkAsPaid(ctx context.Context, tenantID, id, paymentMethod string) error {
	p, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	
	now := time.Now()
	p.PaidAt = &now
	p.Status = "paid"
	p.PaymentMethod = paymentMethod
	p.UpdatedAt = now
	
	return s.repo.Update(ctx, p)
}

// Delete exclui uma conta
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	return s.repo.Delete(ctx, tenantID, id)
}

// GetSummary retorna resumo das contas
func (s *Service) GetSummary(ctx context.Context, tenantID string) (*Summary, error) {
	return s.repo.GetSummary(ctx, tenantID)
}
