package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/tax"
)

// NCMListItem e a representacao publica de um NCM para o simulador.
type NCMListItem struct {
	NCMCode       string  `json:"ncm_code"`
	Description   string  `json:"description"`
	IBSRate       float64 `json:"ibs_rate"`
	CBSRate       float64 `json:"cbs_rate"`
	SelectiveRate float64 `json:"selective_rate"`
	IsBasket      bool    `json:"is_basket"`
	BasketType    string  `json:"basket_type,omitempty"`
}

// SimulatorHandler fornece endpoints publicos para o simulador fiscal.
type SimulatorHandler struct {
	engine   *tax.Engine
	ncmList  func() []*tax.NCMRate
}

func NewSimulatorHandler(e *tax.Engine, listFn func() []*tax.NCMRate) *SimulatorHandler {
	return &SimulatorHandler{engine: e, ncmList: listFn}
}

func (h *SimulatorHandler) RegisterPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tax/simulate", h.Simulate)
	mux.HandleFunc("GET /api/v1/tax/ncm-list", h.ListNCM)
}

// Simulate calcula impostos sem autenticacao (simulador publico).
func (h *SimulatorHandler) Simulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NCMCode   string  `json:"ncm_code"`
		BaseValue float64 `json:"base_value"`
		Operation string  `json:"operation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON invalido")
		return
	}
	if req.Operation == "" {
		req.Operation = "debit_exit"
	}
	result, err := h.engine.Calculate(r.Context(), tax.TaxInput{
		TenantID:      "simulator",
		NCMCode:       req.NCMCode,
		BaseValue:     req.BaseValue,
		Operation:     tax.OperationType(req.Operation),
		ReferenceDate: time.Now(),
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// ListNCM retorna todos os NCMs disponiveis para simulacao.
func (h *SimulatorHandler) ListNCM(w http.ResponseWriter, r *http.Request) {
	rates := h.ncmList()
	items := make([]NCMListItem, 0, len(rates))
	for _, rate := range rates {
		items = append(items, NCMListItem{
			NCMCode:       rate.NCMCode,
			Description:   rate.NCMDescription,
			IBSRate:       rate.IBSRate,
			CBSRate:       rate.CBSRate,
			SelectiveRate: rate.SelectiveRate,
			IsBasket:      rate.BasketReduced,
			BasketType:    rate.BasketType,
		})
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data":  items,
		"total": len(items),
	})
}
