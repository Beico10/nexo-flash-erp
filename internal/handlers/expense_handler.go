// Package handlers — endpoints de despesas e leitor de QR Code.
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nexoone/nexo-one/internal/expenses"
)

type ExpenseHandler struct {
	svc *expenses.Service
}

func NewExpenseHandler(svc *expenses.Service) *ExpenseHandler {
	return &ExpenseHandler{svc: svc}
}

// ════════════════════════════════════════════════════════════
// LEITOR DE QR CODE
// ════════════════════════════════════════════════════════════

// ScanQRCode POST /api/expenses/scan
// Recebe conteúdo do QR Code e registra despesa automaticamente.
func (h *ExpenseHandler) ScanQRCode(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)

	var req struct {
		QRContent string `json:"qr_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.QRContent == "" {
		respondError(w, http.StatusBadRequest, "qr_content é obrigatório")
		return
	}

	// Processar QR Code
	expense, err := h.svc.ProcessQRCode(r.Context(), tenantID, userID, req.QRContent)
	if err != nil {
		switch err {
		case expenses.ErrInvalidQRCode:
			respondError(w, http.StatusBadRequest, "QR Code não reconhecido. Certifique-se de que é uma NFC-e ou NF-e válida.")
		case expenses.ErrDuplicateExpense:
			respondError(w, http.StatusConflict, "Esta nota já foi registrada anteriormente.")
		case expenses.ErrSEFAZUnavailable:
			respondError(w, http.StatusServiceUnavailable, "SEFAZ indisponível. Tente novamente em alguns segundos.")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Despesa registrada com sucesso!",
		"expense": formatExpenseResponse(expense),
	})
}

// ParseQRCode POST /api/expenses/parse-qr
// Apenas analisa o QR Code sem registrar (preview).
func (h *ExpenseHandler) ParseQRCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QRContent string `json:"qr_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	result, err := h.svc.ParseQRCode(req.QRContent)
	if err != nil {
		respondError(w, http.StatusBadRequest, "QR Code não reconhecido")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"type":       result.Type,
		"access_key": result.AccessKey,
		"uf":         result.UF,
		"url":        result.URL,
		"is_valid":   result.IsValid,
	})
}

// UploadXML POST /api/expenses/upload-xml
// Recebe arquivo XML de NF-e e registra despesa.
func (h *ExpenseHandler) UploadXML(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)

	// Limite de 5MB
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	file, _, err := r.FormFile("xml")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Arquivo XML não encontrado")
		return
	}
	defer file.Close()

	xmlData, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Erro ao ler arquivo")
		return
	}

	// Parse do XML
	expense, err := expenses.ParseNFeXML(xmlData)
	if err != nil {
		respondError(w, http.StatusBadRequest, "XML inválido: "+err.Error())
		return
	}

	// Preencher dados
	expense.TenantID = tenantID
	expense.Source = "xml_upload"
	expense.RegisteredBy = userID

	// Salvar
	if err := h.svc.CreateManual(r.Context(), expense); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Nota importada com sucesso!",
		"expense": formatExpenseResponse(expense),
	})
}

// ════════════════════════════════════════════════════════════
// CRUD DE DESPESAS
// ════════════════════════════════════════════════════════════

// CreateExpense POST /api/expenses
// Cria despesa manual.
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)

	var expense expenses.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	expense.TenantID = tenantID
	expense.RegisteredBy = userID

	if err := h.svc.CreateManual(r.Context(), &expense); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Despesa criada com sucesso!",
		"expense": formatExpenseResponse(&expense),
	})
}

// ListExpenses GET /api/expenses
// Lista despesas com filtros.
func (h *ExpenseHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	q := r.URL.Query()

	filter := expenses.ExpenseFilter{
		Category:     q.Get("category"),
		SupplierCNPJ: q.Get("supplier_cnpj"),
		Status:       q.Get("status"),
		Limit:        50,
	}

	if limit, err := strconv.Atoi(q.Get("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = offset
	}
	if from := q.Get("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.DateFrom = &t
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			filter.DateTo = &t
		}
	}

	list, err := h.svc.List(r.Context(), tenantID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"expenses": list,
		"count":    len(list),
	})
}

// GetExpense GET /api/expenses/{id}
// Retorna detalhes de uma despesa.
func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	expenseID := r.PathValue("id")

	expense, err := h.svc.GetByID(r.Context(), tenantID, expenseID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Despesa não encontrada")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"expense": formatExpenseResponse(expense),
	})
}

// DeleteExpense DELETE /api/expenses/{id}
// Cancela uma despesa.
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	expenseID := r.PathValue("id")

	if err := h.svc.Delete(r.Context(), tenantID, expenseID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Despesa cancelada",
	})
}

// ════════════════════════════════════════════════════════════
// CATEGORIAS
// ════════════════════════════════════════════════════════════

// GetCategories GET /api/expenses/categories
// Lista categorias de despesa.
func (h *ExpenseHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	categories, err := h.svc.GetCategories(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// ════════════════════════════════════════════════════════════
// RELATÓRIOS
// ════════════════════════════════════════════════════════════

// GetSummary GET /api/expenses/summary
// Resumo de despesas por período.
func (h *ExpenseHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	q := r.URL.Query()

	// Default: últimos 30 dias
	to := time.Now()
	from := to.AddDate(0, -1, 0)

	if f := q.Get("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			from = t
		}
	}
	if t := q.Get("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			to = parsed
		}
	}

	summary, err := h.svc.GetSummary(r.Context(), tenantID, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calcular totais
	var totalAmount, totalIBS, totalCBS float64
	for _, s := range summary {
		totalAmount += s.Total
		totalIBS += s.IBSCredit
		totalCBS += s.CBSCredit
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"period": map[string]string{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		"summary": summary,
		"totals": map[string]float64{
			"amount":     totalAmount,
			"ibs_credit": totalIBS,
			"cbs_credit": totalCBS,
			"tax_credit": totalIBS + totalCBS,
		},
	})
}

// GetTaxReport GET /api/expenses/tax-report
// Relatório para IR/imposto.
func (h *ExpenseHandler) GetTaxReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	year := time.Now().Year()
	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}

	report, err := h.svc.GetTaxReport(r.Context(), tenantID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calcular totais
	var totalDeductible, totalNonDeductible, totalCredit float64
	for _, r := range report {
		if r.TaxDeductible {
			totalDeductible += r.Total
		} else {
			totalNonDeductible += r.Total
		}
		totalCredit += r.TaxCredit
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"year":   year,
		"report": report,
		"totals": map[string]float64{
			"deductible":     totalDeductible,
			"non_deductible": totalNonDeductible,
			"tax_credit":     totalCredit,
		},
	})
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func formatExpenseResponse(e *expenses.Expense) map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"source":         e.Source,
		"nfe_key":        e.NFeKey,
		"nfe_number":     e.NFeNumber,
		"nfe_type":       e.NFeType,
		"supplier_cnpj":  e.SupplierCNPJ,
		"supplier_name":  e.SupplierName,
		"total_amount":   e.TotalAmount,
		"category":       e.Category,
		"ibs_credit":     e.IBSCredit,
		"cbs_credit":     e.CBSCredit,
		"issue_date":     e.IssueDate.Format("2006-01-02"),
		"registered_at":  e.RegisteredAt.Format(time.RFC3339),
		"status":         e.Status,
		"items_count":    len(e.Items),
	}
}
