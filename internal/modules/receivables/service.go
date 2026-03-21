// Package receivables - Contas a Receber do Nexo One ERP
package receivables

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Receivable representa uma conta a receber
type Receivable struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Description   string     `json:"description"`
	Customer      string     `json:"customer"`
	Category      string     `json:"category"`
	Amount        float64    `json:"amount"`
	DueDate       time.Time  `json:"due_date"`
	ReceivedAt    *time.Time `json:"received_at,omitempty"`
	Status        string     `json:"status"` // pending, received, overdue, cancelled
	PaymentMethod string     `json:"payment_method,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	InvoiceNumber string     `json:"invoice_number,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Repository interface para contas a receber
type Repository interface {
	Create(ctx context.Context, r *Receivable) error
	GetByID(ctx context.Context, tenantID, id string) (*Receivable, error)
	List(ctx context.Context, tenantID string, filter Filter) ([]Receivable, error)
	Update(ctx context.Context, r *Receivable) error
	Delete(ctx context.Context, tenantID, id string) error
	GetSummary(ctx context.Context, tenantID string) (*Summary, error)
}

// Filter para listagem
type Filter struct {
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	Customer  string
	Category  string
}

// Summary resumo de contas a receber
type Summary struct {
	TotalPending   float64 `json:"total_pending"`
	TotalReceived  float64 `json:"total_received"`
	TotalOverdue   float64 `json:"total_overdue"`
	CountPending   int     `json:"count_pending"`
	CountReceived  int     `json:"count_received"`
	CountOverdue   int     `json:"count_overdue"`
}

// Service para contas a receber
type Service struct {
	repo Repository
}

// NewService cria um novo service de contas a receber
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create cria uma nova conta a receber
func (s *Service) Create(ctx context.Context, r *Receivable) error {
	r.ID = uuid.New().String()
	r.Status = "pending"
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	
	if r.DueDate.Before(time.Now()) {
		r.Status = "overdue"
	}
	
	return s.repo.Create(ctx, r)
}

// GetByID busca uma conta por ID
func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Receivable, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// List lista contas a receber
func (s *Service) List(ctx context.Context, tenantID string, filter Filter) ([]Receivable, error) {
	return s.repo.List(ctx, tenantID, filter)
}

// MarkAsReceived marca uma conta como recebida
func (s *Service) MarkAsReceived(ctx context.Context, tenantID, id, paymentMethod string) error {
	r, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	
	now := time.Now()
	r.ReceivedAt = &now
	r.Status = "received"
	r.PaymentMethod = paymentMethod
	r.UpdatedAt = now
	
	return s.repo.Update(ctx, r)
}

// Delete exclui uma conta
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	return s.repo.Delete(ctx, tenantID, id)
}

// GetSummary retorna resumo das contas
func (s *Service) GetSummary(ctx context.Context, tenantID string) (*Summary, error) {
	return s.repo.GetSummary(ctx, tenantID)
}
