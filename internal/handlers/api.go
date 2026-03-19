// Package handlers — handlers compartilhados e utilitários de resposta HTTP.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/ai"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
	"github.com/nexoone/nexo-one/internal/modules/logistics"
	"github.com/nexoone/nexo-one/internal/tax"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

// =============================================================================
// MOTOR FISCAL — /api/v1/tax
// =============================================================================

type TaxHandler struct {
	engine *tax.Engine
}

func NewTaxHandler(e *tax.Engine) *TaxHandler { return &TaxHandler{engine: e} }

func (h *TaxHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tax/calculate", h.Calculate)
}

// Calculate calcula IBS/CBS para um produto.
// POST /api/v1/tax/calculate
// Body: { "ncm_code": "01012100", "base_value": 100.00, "operation": "debit_exit" }
func (h *TaxHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	var req struct {
		NCMCode   string  `json:"ncm_code"`
		BaseValue float64 `json:"base_value"`
		Operation string  `json:"operation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	result, err := h.engine.Calculate(r.Context(), tax.TaxInput{
		TenantID:      tenantID,
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

// =============================================================================
// LOGÍSTICA — /api/v1/logistics
// =============================================================================

type LogisticsHandler struct {
	contracts *logistics.ContractService
}

func NewLogisticsHandler(c *logistics.ContractService) *LogisticsHandler {
	return &LogisticsHandler{contracts: c}
}

func (h *LogisticsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/logistics/freight/calculate", h.CalculateFreight)
}

// CalculateFreight calcula o frete e retorna o DRE da Viagem.
// POST /api/v1/logistics/freight/calculate
func (h *LogisticsHandler) CalculateFreight(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	var req struct {
		ShipperID       string  `json:"shipper_id"`
		VehicleType     string  `json:"vehicle_type"`
		DistanceKM      float64 `json:"distance_km"`
		WeightKG        float64 `json:"weight_kg"`
		TollCost        float64 `json:"toll_cost"`
		FuelCostPerKM   float64 `json:"fuel_cost_per_km"`
		DriverCostPerKM float64 `json:"driver_cost_per_km"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	result, err := h.contracts.Calculate(r.Context(), logistics.FreightInput{
		TenantID:        tenantID,
		ShipperID:       req.ShipperID,
		VehicleType:     logistics.VehicleType(req.VehicleType),
		DistanceKM:      req.DistanceKM,
		WeightKG:        req.WeightKG,
		TollCost:        req.TollCost,
		FuelCostPerKM:   req.FuelCostPerKM,
		DriverCostPerKM: req.DriverCostPerKM,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// =============================================================================
// ESTÉTICA — /api/v1/aesthetics
// =============================================================================

type AestheticsHandler struct {
	agenda *aesthetics.AgendaService
}

func NewAestheticsHandler(a *aesthetics.AgendaService) *AestheticsHandler {
	return &AestheticsHandler{agenda: a}
}

func (h *AestheticsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/aesthetics/appointments", h.Book)
	mux.HandleFunc("GET /api/v1/aesthetics/appointments", h.ListByDate)
	mux.HandleFunc("PATCH /api/v1/aesthetics/appointments/{id}/reschedule", h.Reschedule)
	mux.HandleFunc("POST /api/v1/aesthetics/appointments/{id}/split", h.CalculateSplit)
}

// Book cria um agendamento com trava de conflito.
// POST /api/v1/aesthetics/appointments
func (h *AestheticsHandler) Book(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	var req struct {
		ProfessionalID string    `json:"professional_id"`
		CustomerName   string    `json:"customer_name"`
		CustomerPhone  string    `json:"customer_phone"`
		ServiceID      string    `json:"service_id"`
		ServicePrice   float64   `json:"service_price"`
		StartTime      time.Time `json:"start_time"`
		DurationMin    int       `json:"duration_min"`
		Notes          string    `json:"notes"`
		SplitEnabled   bool      `json:"split_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	apt := &aesthetics.Appointment{
		TenantID:       tenantID,
		ProfessionalID: req.ProfessionalID,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		ServiceID:      req.ServiceID,
		ServicePrice:   req.ServicePrice,
		StartTime:      req.StartTime,
		DurationMin:    req.DurationMin,
		Notes:          req.Notes,
		SplitEnabled:   req.SplitEnabled,
	}
	if err := h.agenda.Book(r.Context(), apt); err != nil {
		// ConflictError retorna 409
		if _, ok := err.(*aesthetics.ConflictError); ok {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, apt)
}

// ListByDate lista agendamentos do dia.
// GET /api/v1/aesthetics/appointments?date=2026-03-19
func (h *AestheticsHandler) ListByDate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}
	list, err := h.agenda.ListByDate(r.Context(), tenantID, date)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"date": date.Format("2006-01-02"), "data": list})
}

// Reschedule remarca um agendamento verificando conflito no novo horário.
// PATCH /api/v1/aesthetics/appointments/{id}/reschedule
func (h *AestheticsHandler) Reschedule(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	id := r.PathValue("id")
	var req struct {
		NewStart time.Time `json:"new_start"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if err := h.agenda.Reschedule(r.Context(), tenantID, id, req.NewStart); err != nil {
		if _, ok := err.(*aesthetics.ConflictError); ok {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "agendamento remarcado"})
}

// CalculateSplit calcula o split de pagamento de um agendamento.
// POST /api/v1/aesthetics/appointments/{id}/split
func (h *AestheticsHandler) CalculateSplit(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	id := r.PathValue("id")
	apt, err := h.agenda.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "agendamento não encontrado")
		return
	}
	result, err := aesthetics.CalculateSplit(apt)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// =============================================================================
// IA — /api/v1/ai
// =============================================================================

type AIHandler struct {
	gateway   *ai.Gateway
	concierge *ai.Concierge
}

func NewAIHandler(g *ai.Gateway, c *ai.Concierge) *AIHandler {
	return &AIHandler{gateway: g, concierge: c}
}

func (h *AIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/ai/suggestions", h.ListPending)
	mux.HandleFunc("POST /api/v1/ai/suggestions/{id}/approve", h.Approve)
	mux.HandleFunc("POST /api/v1/ai/suggestions/{id}/reject", h.Reject)
	mux.HandleFunc("POST /api/v1/ai/concierge/nfe", h.ProcessNFe)
}

// ListPending lista sugestões pendentes de aprovação humana.
// GET /api/v1/ai/suggestions
func (h *AIHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	suggestions, err := h.gateway.GetPendingForUser(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": suggestions, "total": len(suggestions)})
}

// Approve aprova uma sugestão da IA — persiste os dados.
// POST /api/v1/ai/suggestions/{id}/approve
func (h *AIHandler) Approve(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	id := r.PathValue("id")
	if err := h.gateway.Approve(r.Context(), id, claims.UserID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "sugestão aprovada e aplicada"})
}

// Reject rejeita uma sugestão da IA — nenhum dado alterado.
// POST /api/v1/ai/suggestions/{id}/reject
func (h *AIHandler) Reject(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	id := r.PathValue("id")
	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.gateway.Reject(r.Context(), id, claims.UserID, req.Reason); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "sugestão rejeitada"})
}

// ProcessNFe recebe um XML de NF-e e gera sugestões de onboarding via IA Concierge.
// POST /api/v1/ai/concierge/nfe
// Content-Type: application/xml  (body = XML da NF-e)
func (h *AIHandler) ProcessNFe(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	if r.ContentLength > 5*1024*1024 { // limite 5MB
		respondError(w, http.StatusRequestEntityTooLarge, "XML muito grande (máx 5MB)")
		return
	}
	var xmlData []byte
	buf := make([]byte, r.ContentLength)
	n, err := r.Body.Read(buf)
	if err != nil && n == 0 {
		respondError(w, http.StatusBadRequest, "body vazio")
		return
	}
	xmlData = buf[:n]

	nfe, count, err := h.concierge.ProcessNFeXML(r.Context(), tenantID, xmlData)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"message":           "NF-e processada com sucesso",
		"nfe_number":        nfe.NumeroNFe,
		"supplier":          nfe.Emitente.RazaoSocial,
		"items_found":       len(nfe.Items),
		"suggestions_created": count,
		"next_step":         "acesse /api/v1/ai/suggestions para aprovar os produtos detectados",
	})
}

// =============================================================================
// UTILITÁRIOS DE RESPOSTA HTTP
// =============================================================================

// respondJSON serializa v como JSON e escreve na resposta.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// respondError escreve um erro JSON padronizado.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{
		"error": msg,
		"code":  http.StatusText(status),
	})
}
