package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/expenses"
)

// ExpenseRepo implementa expenses.ExpenseRepository in-memory.
type ExpenseRepo struct {
	mu         sync.RWMutex
	expenses   []*expenses.Expense
	items      map[string][]expenses.ExpenseItem // expenseID -> items
	categories []expenses.ExpenseCategory
	scanLogs   []scanLogEntry
}

type scanLogEntry struct {
	tenantID, userID, content, qrType, expenseID, errorMsg string
	success                                                 bool
	at                                                      time.Time
}

func NewExpenseRepo() *ExpenseRepo {
	r := &ExpenseRepo{
		items: make(map[string][]expenses.ExpenseItem),
	}
	r.seedCategories()
	r.seedDemoExpenses()
	return r
}

func (r *ExpenseRepo) seedCategories() {
	r.categories = []expenses.ExpenseCategory{
		{ID: "cat-1", Code: "pecas", Name: "Pecas e Componentes", Icon: "wrench", Color: "purple", TaxDeductible: true, NCMPatterns: []string{"8708", "8409", "8413"}},
		{ID: "cat-2", Code: "mercadorias", Name: "Mercadorias para Revenda", Icon: "package", Color: "blue", TaxDeductible: true, NCMPatterns: []string{"1901", "1905", "2106"}},
		{ID: "cat-3", Code: "materiais", Name: "Materiais de Consumo", Icon: "box", Color: "green", TaxDeductible: true, NCMPatterns: []string{"3401", "4818", "3923"}},
		{ID: "cat-4", Code: "combustivel", Name: "Combustivel", Icon: "fuel", Color: "yellow", TaxDeductible: true, NCMPatterns: []string{"2710"}},
		{ID: "cat-5", Code: "alimentacao", Name: "Alimentacao", Icon: "utensils", Color: "red", TaxDeductible: false},
		{ID: "cat-6", Code: "servicos", Name: "Servicos Terceiros", Icon: "users", Color: "indigo", TaxDeductible: true},
		{ID: "cat-7", Code: "equipamentos", Name: "Equipamentos", Icon: "monitor", Color: "pink", TaxDeductible: true, NCMPatterns: []string{"8471", "8528"}},
		{ID: "cat-8", Code: "outros", Name: "Outros", Icon: "file", Color: "gray", TaxDeductible: false},
	}
}

func (r *ExpenseRepo) seedDemoExpenses() {
	now := time.Now()
	tenantID := "00000000-0000-0000-0000-000000000001"

	demoExpenses := []*expenses.Expense{
		{
			ID: "exp-1", TenantID: tenantID, Source: "qrcode",
			NFeKey: "35260312345678000199550010000012341234567890", NFeNumber: "001234", NFeType: "nfce",
			SupplierCNPJ: "12345678000199", SupplierName: "Auto Pecas Central",
			TotalProducts: 487.50, TotalAmount: 487.50, IBSCredit: 45.09, CBSCredit: 18.28,
			Category: "pecas", IssueDate: now.AddDate(0, 0, -2), RegisteredAt: now.AddDate(0, 0, -2),
			RegisteredBy: "c1d4741e-7af6-4ba1-a1a3-9c37377c7524", Status: "registered",
		},
		{
			ID: "exp-2", TenantID: tenantID, Source: "qrcode",
			NFeKey: "35260398765432000188550010000056785678901234", NFeNumber: "005678", NFeType: "nfce",
			SupplierCNPJ: "98765432000188", SupplierName: "Posto Shell BR-101",
			TotalProducts: 320.00, TotalAmount: 320.00, IBSCredit: 29.60, CBSCredit: 12.00,
			Category: "combustivel", IssueDate: now.AddDate(0, 0, -5), RegisteredAt: now.AddDate(0, 0, -5),
			RegisteredBy: "c1d4741e-7af6-4ba1-a1a3-9c37377c7524", Status: "registered",
		},
		{
			ID: "exp-3", TenantID: tenantID, Source: "manual",
			SupplierCNPJ: "11222333000144", SupplierName: "Papelaria Express",
			TotalProducts: 89.90, TotalAmount: 89.90, IBSCredit: 8.32, CBSCredit: 3.37,
			Category: "materiais", IssueDate: now.AddDate(0, 0, -7), RegisteredAt: now.AddDate(0, 0, -7),
			RegisteredBy: "c1d4741e-7af6-4ba1-a1a3-9c37377c7524", Status: "registered",
		},
		{
			ID: "exp-4", TenantID: tenantID, Source: "qrcode",
			NFeKey: "35260355566677000155550010000099009900112233", NFeNumber: "009900", NFeType: "nfce",
			SupplierCNPJ: "55566677000155", SupplierName: "Distribuidora Nacional",
			TotalProducts: 1250.00, TotalAmount: 1250.00, IBSCredit: 115.63, CBSCredit: 46.88,
			Category: "mercadorias", IssueDate: now.AddDate(0, 0, -1), RegisteredAt: now.AddDate(0, 0, -1),
			RegisteredBy: "c1d4741e-7af6-4ba1-a1a3-9c37377c7524", Status: "registered",
		},
		{
			ID: "exp-5", TenantID: tenantID, Source: "qrcode",
			NFeKey: "35260377788899000166550010000033003300445566", NFeNumber: "003300", NFeType: "nfce",
			SupplierCNPJ: "77788899000166", SupplierName: "Restaurante Bom Sabor",
			TotalProducts: 45.00, TotalAmount: 45.00, IBSCredit: 0, CBSCredit: 0,
			Category: "alimentacao", IssueDate: now.AddDate(0, 0, -3), RegisteredAt: now.AddDate(0, 0, -3),
			RegisteredBy: "c1d4741e-7af6-4ba1-a1a3-9c37377c7524", Status: "registered",
		},
	}

	r.expenses = demoExpenses
	r.items["exp-1"] = []expenses.ExpenseItem{
		{ID: "item-1", ItemOrder: 1, Description: "Filtro de oleo", Quantity: 2, UnitPrice: 35.00, TotalPrice: 70.00, Unit: "UN", NCM: "84219999"},
		{ID: "item-2", ItemOrder: 2, Description: "Pastilha de freio", Quantity: 4, UnitPrice: 89.00, TotalPrice: 356.00, Unit: "UN", NCM: "87083010"},
		{ID: "item-3", ItemOrder: 3, Description: "Oleo lubrificante 5W30", Quantity: 1, UnitPrice: 61.50, TotalPrice: 61.50, Unit: "UN", NCM: "27101932"},
	}
	r.items["exp-4"] = []expenses.ExpenseItem{
		{ID: "item-4", ItemOrder: 1, Description: "Amortecedor dianteiro", Quantity: 2, UnitPrice: 350.00, TotalPrice: 700.00, Unit: "UN", NCM: "87088000"},
		{ID: "item-5", ItemOrder: 2, Description: "Mola espiral", Quantity: 2, UnitPrice: 275.00, TotalPrice: 550.00, Unit: "UN", NCM: "73202010"},
	}
}

// ── Interface implementation ─────────────────────────────────

func (r *ExpenseRepo) Create(_ context.Context, e *expenses.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e.ID == "" {
		e.ID = fmt.Sprintf("exp-%d", time.Now().UnixNano())
	}
	e.RegisteredAt = time.Now()
	if e.Status == "" {
		e.Status = "registered"
	}
	r.expenses = append(r.expenses, e)
	return nil
}

func (r *ExpenseRepo) GetByID(_ context.Context, tenantID, id string) (*expenses.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.expenses {
		if e.ID == id && e.TenantID == tenantID {
			return e, nil
		}
	}
	return nil, expenses.ErrExpenseNotFound
}

func (r *ExpenseRepo) GetByNFeKey(_ context.Context, tenantID, nfeKey string) (*expenses.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.expenses {
		if e.NFeKey == nfeKey && e.TenantID == tenantID {
			return e, nil
		}
	}
	return nil, expenses.ErrExpenseNotFound
}

func (r *ExpenseRepo) List(_ context.Context, tenantID string, filter expenses.ExpenseFilter) ([]*expenses.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*expenses.Expense
	for _, e := range r.expenses {
		if e.TenantID != tenantID {
			continue
		}
		if filter.Category != "" && e.Category != filter.Category {
			continue
		}
		if filter.Status != "" && e.Status != filter.Status {
			continue
		}
		if filter.SupplierCNPJ != "" && e.SupplierCNPJ != filter.SupplierCNPJ {
			continue
		}
		if filter.DateFrom != nil && e.IssueDate.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && e.IssueDate.After(*filter.DateTo) {
			continue
		}
		result = append(result, e)
	}
	// Apply offset/limit
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *ExpenseRepo) Update(_ context.Context, e *expenses.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, ex := range r.expenses {
		if ex.ID == e.ID && ex.TenantID == e.TenantID {
			r.expenses[i] = e
			return nil
		}
	}
	return expenses.ErrExpenseNotFound
}

func (r *ExpenseRepo) Delete(_ context.Context, tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.expenses {
		if e.ID == id && e.TenantID == tenantID {
			r.expenses[i].Status = "cancelled"
			return nil
		}
	}
	return expenses.ErrExpenseNotFound
}

func (r *ExpenseRepo) CreateItems(_ context.Context, _, expenseID string, items []expenses.ExpenseItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[expenseID] = items
	return nil
}

func (r *ExpenseRepo) GetItems(_ context.Context, _, expenseID string) ([]expenses.ExpenseItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.items[expenseID], nil
}

func (r *ExpenseRepo) GetCategories(_ context.Context, _ string) ([]expenses.ExpenseCategory, error) {
	return r.categories, nil
}

func (r *ExpenseRepo) AutoCategorize(_ context.Context, _, ncm string) (string, error) {
	for _, cat := range r.categories {
		for _, pattern := range cat.NCMPatterns {
			if strings.HasPrefix(ncm, pattern) {
				return cat.Code, nil
			}
		}
	}
	return "outros", nil
}

func (r *ExpenseRepo) GetSummary(_ context.Context, tenantID string, from, to time.Time) ([]expenses.ExpenseSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	catMap := make(map[string]*expenses.ExpenseSummary)
	for _, e := range r.expenses {
		if e.TenantID != tenantID || e.Status == "cancelled" {
			continue
		}
		if e.IssueDate.Before(from) || e.IssueDate.After(to) {
			continue
		}
		key := e.Category
		if catMap[key] == nil {
			catMap[key] = &expenses.ExpenseSummary{Category: key}
		}
		catMap[key].Count++
		catMap[key].Total += e.TotalAmount
		catMap[key].IBSCredit += e.IBSCredit
		catMap[key].CBSCredit += e.CBSCredit
	}
	var result []expenses.ExpenseSummary
	for _, s := range catMap {
		result = append(result, *s)
	}
	return result, nil
}

func (r *ExpenseRepo) GetTaxReport(_ context.Context, tenantID string, year int) ([]expenses.TaxReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	catMap := make(map[string]*expenses.TaxReport)
	for _, e := range r.expenses {
		if e.TenantID != tenantID || e.Status == "cancelled" || e.IssueDate.Year() != year {
			continue
		}
		key := e.Category
		if catMap[key] == nil {
			deductible := false
			for _, c := range r.categories {
				if c.Code == key {
					deductible = c.TaxDeductible
					break
				}
			}
			catMap[key] = &expenses.TaxReport{Year: year, CategoryCode: key, CategoryName: key, TaxDeductible: deductible}
		}
		catMap[key].DocCount++
		catMap[key].Total += e.TotalAmount
		catMap[key].TaxCredit += e.IBSCredit + e.CBSCredit
	}
	var result []expenses.TaxReport
	for _, r := range catMap {
		result = append(result, *r)
	}
	return result, nil
}

func (r *ExpenseRepo) LogScan(_ context.Context, tenantID, userID, content, qrType string, success bool, expenseID, errorMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scanLogs = append(r.scanLogs, scanLogEntry{
		tenantID: tenantID, userID: userID, content: content, qrType: qrType,
		success: success, expenseID: expenseID, errorMsg: errorMsg, at: time.Now(),
	})
	return nil
}
