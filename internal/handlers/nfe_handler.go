package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/pkg/middleware"
)

type NFEDocument struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"-"`
	Type          string    `json:"type"`
	Number        string    `json:"number"`
	Series        string    `json:"series"`
	AccessKey     string    `json:"access_key"`
	RecipientName string    `json:"recipient_name"`
	RecipientCNPJ string    `json:"recipient_cnpj"`
	Description   string    `json:"description"`
	Total         float64   `json:"total"`
	Status        string    `json:"status"`
	IssuedAt      time.Time `json:"issued_at"`
	XMLContent    string    `json:"-"`
}

type NFEHandler struct {
	mu   sync.RWMutex
	docs []*NFEDocument
	seq  int
}

func NewNFEHandler() *NFEHandler {
	h := &NFEHandler{seq: 100}
	h.seedDocs()
	return h
}

func (h *NFEHandler) seedDocs() {
	tid := "00000000-0000-0000-0000-000000000001"
	now := time.Now()
	h.docs = []*NFEDocument{
		{ID: "nfe-1", TenantID: tid, Type: "nfe", Number: "000098", Series: "001", AccessKey: "35260312345678000199550010000000981234567890", RecipientName: "Empresa ABC Ltda", RecipientCNPJ: "12345678000199", Total: 4500.00, Status: "authorized", IssuedAt: now.AddDate(0, 0, -3)},
		{ID: "nfe-2", TenantID: tid, Type: "nfe", Number: "000099", Series: "001", AccessKey: "35260312345678000199550010000000991234567891", RecipientName: "Comercio XYZ ME", RecipientCNPJ: "98765432000188", Total: 1890.50, Status: "authorized", IssuedAt: now.AddDate(0, 0, -1)},
		{ID: "nfe-3", TenantID: tid, Type: "nfce", Number: "000045", Series: "001", AccessKey: "35260312345678000199650010000000451234567892", RecipientName: "Consumidor Final", RecipientCNPJ: "00000000000", Total: 259.90, Status: "authorized", IssuedAt: now.AddDate(0, 0, -1)},
		{ID: "cte-1", TenantID: tid, Type: "cte", Number: "000012", Series: "001", AccessKey: "35260312345678000199570010000000121234567893", RecipientName: "Transportes Rapido SA", RecipientCNPJ: "55566677000155", Total: 780.00, Status: "authorized", IssuedAt: now.AddDate(0, 0, -5)},
	}
}

func (h *NFEHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/nfe/documents", h.ListDocuments)
	mux.HandleFunc("GET /api/v1/nfe/documents/{id}", h.GetDocument)
	mux.HandleFunc("POST /api/v1/nfe/emit", h.EmitDocument)
	mux.HandleFunc("POST /api/v1/nfe/cancel/{id}", h.CancelDocument)
}

func (h *NFEHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	h.mu.RLock()
	defer h.mu.RUnlock()
	var result []*NFEDocument
	for _, d := range h.docs {
		if d.TenantID == tenantID { result = append(result, d) }
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"documents": result, "count": len(result)})
}

func (h *NFEHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, d := range h.docs {
		if d.ID == id && d.TenantID == tenantID {
			respondJSON(w, http.StatusOK, map[string]interface{}{"document": d})
			return
		}
	}
	respondError(w, http.StatusNotFound, "documento nao encontrado")
}

func (h *NFEHandler) EmitDocument(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	var req struct {
		Type          string  `json:"type"`
		RecipientName string  `json:"recipient_name"`
		RecipientCNPJ string  `json:"recipient_cnpj"`
		Description   string  `json:"description"`
		Total         float64 `json:"total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondError(w, http.StatusBadRequest, "JSON invalido"); return }
	if req.Total <= 0 { respondError(w, http.StatusBadRequest, "total deve ser maior que zero"); return }

	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++

	key := make([]byte, 22)
	rand.Read(key)

	model := "55"
	switch req.Type {
	case "nfce": model = "65"
	case "cte":  model = "57"
	}

	doc := &NFEDocument{
		ID:            fmt.Sprintf("nfe-%d", time.Now().UnixNano()),
		TenantID:      tenantID,
		Type:          req.Type,
		Number:        fmt.Sprintf("%06d", h.seq),
		Series:        "001",
		AccessKey:     fmt.Sprintf("3526031234567800019%s001%010d%s", model, h.seq, fmt.Sprintf("%x", key)[:20]),
		RecipientName: req.RecipientName,
		RecipientCNPJ: req.RecipientCNPJ,
		Description:   req.Description,
		Total:         req.Total,
		Status:        "authorized",
		IssuedAt:      time.Now(),
	}
	h.docs = append([]*NFEDocument{doc}, h.docs...)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"document": doc,
		"message":  "Documento emitido com sucesso (homologacao)",
	})
}

func (h *NFEHandler) CancelDocument(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, d := range h.docs {
		if d.ID == id && d.TenantID == tenantID {
			if d.Status != "authorized" { respondError(w, http.StatusBadRequest, "so documentos autorizados podem ser cancelados"); return }
			d.Status = "cancelled"
			respondJSON(w, http.StatusOK, map[string]interface{}{"cancelled": true, "document": d})
			return
		}
	}
	respondError(w, http.StatusNotFound, "documento nao encontrado")
}
