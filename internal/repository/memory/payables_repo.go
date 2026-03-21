// Package memory — repositório in-memory de Contas a Pagar.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/payables"
)

// PayablesRepo repositório in-memory thread-safe.
type PayablesRepo struct {
	mu   sync.RWMutex
	data map[string]*payables.Payable // key: tenantID+id
}

func NewPayablesRepo() *PayablesRepo {
	r := &PayablesRepo{
		data: make(map[string]*payables.Payable),
	}
	r.seed()
	return r
}

func (r *PayablesRepo) key(tenantID, id string) string {
	return tenantID + ":" + id
}

// seed cria dados demo para o tenant demo.
func (r *PayablesRepo) seed() {
	now := time.Now()
	demo := []*payables.Payable{
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "Aluguel do galpão — Março/2026",
			SupplierName: "Imobiliária Central", Category: "aluguel",
			Amount: 3500.00, DueDate: now.AddDate(0, 0, 3),
			Installment: 1, TotalInstallments: 1,
			Recurrence: payables.RecurrenceMonthly,
			Status: payables.StatusPending, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "Fornecedor Peças Automotivas — NF 4521",
			SupplierName: "Auto Peças Silva", SupplierCNPJ: "12.345.678/0001-90",
			Category: "fornecedor",
			Amount: 1850.00, DueDate: now.AddDate(0, 0, -2), // vencida
			Installment: 1, TotalInstallments: 1,
			Recurrence: payables.RecurrenceNone,
			Status: payables.StatusOverdue, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "SIMPLES Nacional — Fevereiro/2026",
			SupplierName: "Receita Federal", Category: "imposto",
			Amount: 780.00, DueDate: now.AddDate(0, 0, 1),
			Installment: 1, TotalInstallments: 1,
			Recurrence: payables.RecurrenceMonthly,
			Status: payables.StatusPending, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "Sistema de Gestão Nexo One — Março/2026",
			SupplierName: "Nexo One ERP", Category: "servico",
			Amount: 197.00, DueDate: now.AddDate(0, 0, 10),
			Installment: 1, TotalInstallments: 1,
			Recurrence: payables.RecurrenceMonthly,
			Status: payables.StatusPending, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo",
			Description: "Conta de Energia — Fevereiro/2026",
			SupplierName: "CPFL Energia", Category: "servico",
			Amount: 450.00, DueDate: now.AddDate(0, 0, -5), // já paga
			AmountPaid: 450.00, PaymentMethod: "pix",
			Installment: 1, TotalInstallments: 1,
			Recurrence: payables.RecurrenceMonthly,
			Status: payables.StatusPaid, CreatedAt: now, UpdatedAt: now,
		},
	}

	// Definir PaidAt para as pagas
	paidTime := now.AddDate(0, 0, -5)
	demo[4].PaidAt = &paidTime

	for _, p := range demo {
		r.data[r.key(p.TenantID, p.ID)] = p
	}
}

func (r *PayablesRepo) Create(ctx context.Context, p *payables.Payable) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	// Cópia para evitar modificações externas
	copy := *p
	r.data[r.key(p.TenantID, p.ID)] = &copy
	return nil
}

func (r *PayablesRepo) GetByID(ctx context.Context, tenantID, id string) (*payables.Payable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.data[r.key(tenantID, id)]
	if !ok {
		return nil, payables.ErrPayableNotFound
	}
	copy := *p
	return &copy, nil
}

func (r *PayablesRepo) List(ctx context.Context, tenantID string, filter payables.PayableFilter) ([]*payables.Payable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := tenantID + ":"
	var result []*payables.Payable

	for k, p := range r.data {
		if len(k) < len(prefix) || k[:len(prefix)] != prefix {
			continue
		}

		// Filtros
		if filter.Status != "" && p.Status != filter.Status {
			continue
		}
		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.Overdue && p.Status != payables.StatusOverdue {
			continue
		}
		if filter.DueFrom != nil && p.DueDate.Before(*filter.DueFrom) {
			continue
		}
		if filter.DueTo != nil && p.DueDate.After(*filter.DueTo) {
			continue
		}

		copy := *p
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

	// Paginação
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

func (r *PayablesRepo) Update(ctx context.Context, p *payables.Payable) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[r.key(p.TenantID, p.ID)]; !ok {
		return payables.ErrPayableNotFound
	}

	p.UpdatedAt = time.Now()
	copy := *p
	r.data[r.key(p.TenantID, p.ID)] = &copy
	return nil
}

func (r *PayablesRepo) Delete(ctx context.Context, tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, r.key(tenantID, id))
	return nil
}

func (r *PayablesRepo) GetOverdue(ctx context.Context, tenantID string) ([]*payables.Payable, error) {
	now := time.Now()
	filter := payables.PayableFilter{Status: payables.StatusPending}
	all, err := r.List(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	var overdue []*payables.Payable
	for _, p := range all {
		if p.DueDate.Before(now) {
			overdue = append(overdue, p)
		}
	}
	return overdue, nil
}

func (r *PayablesRepo) GetDueSoon(ctx context.Context, tenantID string, days int) ([]*payables.Payable, error) {
	now := time.Now()
	target := now.AddDate(0, 0, days)
	targetDate := target.Format("2006-01-02")

	filter := payables.PayableFilter{Status: payables.StatusPending}
	all, err := r.List(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	var dueSoon []*payables.Payable
	for _, p := range all {
		if p.DueDate.Format("2006-01-02") == targetDate && !p.AlertSent {
			dueSoon = append(dueSoon, p)
		}
	}
	return dueSoon, nil
}

func (r *PayablesRepo) GetSummary(ctx context.Context, tenantID string) (*payables.PayablesSummary, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	all, err := r.List(ctx, tenantID, payables.PayableFilter{Limit: 1000})
	if err != nil {
		return nil, err
	}

	summary := &payables.PayablesSummary{}
	for _, p := range all {
		switch p.Status {
		case payables.StatusPending:
			summary.TotalPending += p.Amount
			summary.CountPending++
			// Próximo vencimento
			if summary.NextDue == nil || p.DueDate.Before(summary.NextDue.DueDate) {
				copy := *p
				summary.NextDue = &copy
			}
		case payables.StatusOverdue:
			summary.TotalOverdue += p.Amount
			summary.CountOverdue++
		case payables.StatusPaid:
			if p.PaidAt != nil && p.PaidAt.After(startOfMonth) {
				summary.TotalPaidMonth += p.AmountPaid
			}
		}
	}

	return summary, nil
}

func (r *PayablesRepo) GetCashFlow(ctx context.Context, tenantID string, from, to time.Time) ([]*payables.CashFlowItem, error) {
	filter := payables.PayableFilter{
		DueFrom: &from,
		DueTo:   &to,
		Limit:   200,
	}
	all, err := r.List(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	var items []*payables.CashFlowItem
	for _, p := range all {
		items = append(items, &payables.CashFlowItem{
			Date:        p.DueDate,
			Description: fmt.Sprintf("%s — %s", p.Description, p.SupplierName),
			Amount:      p.Amount,
			Category:    p.Category,
			Status:      p.Status,
		})
	}
	return items, nil
}

func (r *PayablesRepo) MarkAlertSent(ctx context.Context, tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.data[r.key(tenantID, id)]; ok {
		p.AlertSent = true
	}
	return nil
}
