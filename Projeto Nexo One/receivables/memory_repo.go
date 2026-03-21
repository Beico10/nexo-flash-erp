// Package memory — repositório in-memory de Contas a Receber.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/receivables"
)

type ReceivablesRepo struct {
	mu   sync.RWMutex
	data map[string]*receivables.Receivable
}

func NewReceivablesRepo() *ReceivablesRepo {
	r := &ReceivablesRepo{data: make(map[string]*receivables.Receivable)}
	r.seed()
	return r
}

func (r *ReceivablesRepo) key(tenantID, id string) string { return tenantID + ":" + id }

func (r *ReceivablesRepo) seed() {
	now := time.Now()
	paidAt := now.AddDate(0, 0, -3)
	demo := []*receivables.Receivable{
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "OS #1042 — Troca de embreagem — Fiat Uno",
			CustomerName: "João da Silva", CustomerPhone: "5511991234567",
			Category: "servico", Amount: 850.00,
			DueDate: now.AddDate(0, 0, 5),
			Installment: 1, TotalInstallments: 1,
			Status: receivables.StatusPending,
			ReferenceType: "os", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "OS #1039 — Revisão completa — VW Gol",
			CustomerName: "Maria Oliveira", CustomerPhone: "5511987654321",
			Category: "servico", Amount: 420.00,
			DueDate: now.AddDate(0, 0, -4), // vencida
			Installment: 1, TotalInstallments: 1,
			Status: receivables.StatusOverdue,
			ReferenceType: "os", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "Mensalidade Frotas — Transportadora ABC",
			CustomerName: "Transportadora ABC", CustomerPhone: "5511955551234",
			Category: "mensalidade", Amount: 1200.00,
			DueDate: now.AddDate(0, 0, 10),
			Installment: 1, TotalInstallments: 1,
			Recurrence: receivables.RecurrenceMonthly,
			Status: receivables.StatusPending,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "OS #1035 — Alinhamento e balanceamento",
			CustomerName: "Carlos Souza", CustomerPhone: "5511944443333",
			Category: "servico", Amount: 180.00,
			DueDate: now.AddDate(0, 0, -3),
			AmountReceived: 180.00, PaymentMethod: "pix",
			ReceivedAt: &paidAt,
			Installment: 1, TotalInstallments: 1,
			Status: receivables.StatusReceived,
			ReferenceType: "os", CreatedAt: now, UpdatedAt: now,
		},
	}

	for _, rec := range demo {
		r.data[r.key(rec.TenantID, rec.ID)] = rec
	}
}

func (r *ReceivablesRepo) Create(ctx context.Context, rec *receivables.Receivable) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	rec.CreatedAt = time.Now()
	rec.UpdatedAt = time.Now()
	copy := *rec
	r.data[r.key(rec.TenantID, rec.ID)] = &copy
	return nil
}

func (r *ReceivablesRepo) GetByID(ctx context.Context, tenantID, id string) (*receivables.Receivable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.data[r.key(tenantID, id)]
	if !ok {
		return nil, receivables.ErrNotFound
	}
	copy := *rec
	return &copy, nil
}

func (r *ReceivablesRepo) List(ctx context.Context, tenantID string, filter receivables.ReceivableFilter) ([]*receivables.Receivable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := tenantID + ":"
	var result []*receivables.Receivable

	for k, rec := range r.data {
		if len(k) < len(prefix) || k[:len(prefix)] != prefix {
			continue
		}
		if filter.Status != "" && rec.Status != filter.Status {
			continue
		}
		if filter.Category != "" && rec.Category != filter.Category {
			continue
		}
		if filter.Overdue && rec.Status != receivables.StatusOverdue {
			continue
		}
		if filter.Customer != "" && rec.CustomerName != filter.Customer {
			continue
		}
		copy := *rec
		result = append(result, &copy)
	}

	// Ordenar por vencimento
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].DueDate.After(result[j].DueDate) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *ReceivablesRepo) Update(ctx context.Context, rec *receivables.Receivable) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[r.key(rec.TenantID, rec.ID)]; !ok {
		return receivables.ErrNotFound
	}
	rec.UpdatedAt = time.Now()
	copy := *rec
	r.data[r.key(rec.TenantID, rec.ID)] = &copy
	return nil
}

func (r *ReceivablesRepo) Delete(ctx context.Context, tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, r.key(tenantID, id))
	return nil
}

func (r *ReceivablesRepo) GetOverdue(ctx context.Context, tenantID string) ([]*receivables.Receivable, error) {
	filter := receivables.ReceivableFilter{Status: receivables.StatusOverdue, Limit: 100}
	return r.List(ctx, tenantID, filter)
}

func (r *ReceivablesRepo) GetDueSoon(ctx context.Context, tenantID string, days int) ([]*receivables.Receivable, error) {
	target := time.Now().AddDate(0, 0, days).Format("2006-01-02")
	filter := receivables.ReceivableFilter{Status: receivables.StatusPending, Limit: 100}
	all, err := r.List(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}
	var result []*receivables.Receivable
	for _, rec := range all {
		if rec.DueDate.Format("2006-01-02") == target {
			result = append(result, rec)
		}
	}
	return result, nil
}

func (r *ReceivablesRepo) GetSummary(ctx context.Context, tenantID string) (*receivables.ReceivablesSummary, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	all, _ := r.List(ctx, tenantID, receivables.ReceivableFilter{Limit: 1000})

	summary := &receivables.ReceivablesSummary{}
	total := 0

	for _, rec := range all {
		total++
		switch rec.Status {
		case receivables.StatusPending:
			summary.TotalPending += rec.Amount
			summary.CountPending++
			if summary.NextDue == nil || rec.DueDate.Before(summary.NextDue.DueDate) {
				copy := *rec
				summary.NextDue = &copy
			}
		case receivables.StatusOverdue:
			summary.TotalOverdue += rec.Amount
			summary.CountOverdue++
		case receivables.StatusReceived:
			if rec.ReceivedAt != nil && rec.ReceivedAt.After(startOfMonth) {
				summary.TotalReceivedMonth += rec.AmountReceived
			}
		}
	}

	if total > 0 {
		summary.OverdueRate = float64(summary.CountOverdue) / float64(total) * 100
	}

	return summary, nil
}
