package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/mechanic"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type MechanicHandler struct {
	service *mechanic.OSService
}

func NewMechanicHandler(s *mechanic.OSService) *MechanicHandler {
	return &MechanicHandler{service: s}
}

func (h *MechanicHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/mechanic/os", h.CreateOS)
	mux.HandleFunc("GET /api/v1/mechanic/os", h.ListOpenOS)
	mux.HandleFunc("GET /api/v1/mechanic/os/plate/{plate}", h.GetByPlate)
	mux.HandleFunc("GET /api/v1/mechanic/os/{id}", h.GetOS)
	mux.HandleFunc("PATCH /api/v1/mechanic/os/{id}/status", h.UpdateStatus)
	mux.HandleFunc("POST /api/v1/mechanic/os/{id}/send-approval", h.SendApproval)
	mux.HandleFunc("POST /api/v1/mechanic/approve/{token}", h.ApproveByToken)
}

func (h *MechanicHandler) CreateOS(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	var req struct {
		VehiclePlate  string `json:"vehicle_plate"`
		VehicleKM     int    `json:"vehicle_km"`
		VehicleModel  string `json:"vehicle_model"`
		VehicleYear   int    `json:"vehicle_year"`
		CustomerID    string `json:"customer_id"`
		CustomerPhone string `json:"customer_phone"`
		Complaint     string `json:"complaint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON invalido")
		return
	}
	if req.VehiclePlate == "" {
		respondError(w, http.StatusBadRequest, "vehicle_plate e obrigatorio")
		return
	}
	os := &mechanic.ServiceOrder{
		TenantID:      tenantID,
		VehiclePlate:  req.VehiclePlate,
		VehicleKM:     req.VehicleKM,
		VehicleModel:  req.VehicleModel,
		VehicleYear:   req.VehicleYear,
		CustomerID:    req.CustomerID,
		CustomerPhone: req.CustomerPhone,
		Complaint:     req.Complaint,
	}
	if err := h.service.Create(r.Context(), os); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, os)
}

func (h *MechanicHandler) GetOS(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	id := r.PathValue("id")
	os, err := h.service.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "OS nao encontrada")
		return
	}
	respondJSON(w, http.StatusOK, os)
}

func (h *MechanicHandler) ListOpenOS(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	list, err := h.service.ListOpen(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": list, "total": len(list)})
}

func (h *MechanicHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	id := r.PathValue("id")
	var req struct {
		Status    mechanic.OSStatus `json:"status"`
		Diagnosis string            `json:"diagnosis,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON invalido")
		return
	}
	os, err := h.service.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "OS nao encontrada")
		return
	}
	os.Status = req.Status
	if req.Diagnosis != "" {
		os.Diagnosis = req.Diagnosis
	}
	os.UpdatedAt = time.Now().UTC()
	if err := h.service.Update(r.Context(), os); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, os)
}

func (h *MechanicHandler) SendApproval(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	id := r.PathValue("id")
	if err := h.service.SendForApproval(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "link de aprovacao enviado via WhatsApp"})
}

func (h *MechanicHandler) ApproveByToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if err := h.service.ApproveByToken(r.Context(), "", token); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "orcamento aprovado com sucesso"})
}

func (h *MechanicHandler) GetByPlate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	plate := r.PathValue("plate")
	list, err := h.service.GetByPlate(r.Context(), tenantID, plate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"plate": plate, "data": list})
}
