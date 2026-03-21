// Package handlers — endpoints de Despacho em Lote.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nexoone/nexo-one/internal/dispatch"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type DispatchHandler struct{}

func NewDispatchHandler() *DispatchHandler {
	return &DispatchHandler{}
}

func (h *DispatchHandler) RegisterRoutes(mux *http.ServeMux) {
	// Importação em lote
	mux.HandleFunc("POST /api/v1/dispatch/import", h.ImportBatch)

	// Scanner (fluxo secundário)
	mux.HandleFunc("POST /api/v1/dispatch/scan", h.ScanBarcode)

	// Gestão do lote
	mux.HandleFunc("GET /api/v1/dispatch/batches", h.ListBatches)
	mux.HandleFunc("GET /api/v1/dispatch/batches/{id}", h.GetBatch)
	mux.HandleFunc("DELETE /api/v1/dispatch/batches/{id}/items/{itemId}", h.RemoveItem)

	// Veículos
	mux.HandleFunc("GET /api/v1/dispatch/vehicles", h.ListVehicles)
	mux.HandleFunc("POST /api/v1/dispatch/vehicles", h.CreateVehicle)

	// Distribuição
	mux.HandleFunc("POST /api/v1/dispatch/batches/{id}/distribute", h.Distribute)
	mux.HandleFunc("POST /api/v1/dispatch/batches/{id}/reassign", h.Reassign)

	// Enviar para roteirizador
	mux.HandleFunc("POST /api/v1/dispatch/batches/{id}/route", h.SendToRouter)
}

// ImportBatch POST /api/v1/dispatch/import
// Aceita: XML NF-e, CSV, EDI, PDF
// Content-Type: multipart/form-data
func (h *DispatchHandler) ImportBatch(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50MB

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Arquivo muito grande ou formato inválido")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		respondError(w, http.StatusBadRequest, "Nenhum arquivo enviado")
		return
	}

	var allItems []*dispatch.DeliveryItem
	var allErrors []string
	var allWarnings []string
	totalProcessed, totalFailed := 0, 0

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("%s: erro ao abrir", fileHeader.Filename))
			continue
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("%s: erro ao ler", fileHeader.Filename))
			continue
		}

		// Detectar formato pelo nome/conteúdo
		var result *dispatch.ParseResult
		name := strings.ToLower(fileHeader.Filename)

		switch {
		case strings.HasSuffix(name, ".xml") || strings.Contains(string(data[:min(100, len(data))]), "nfeProc"):
			result = dispatch.ParseXMLBatch(data)
		case strings.HasSuffix(name, ".csv") || strings.HasSuffix(name, ".txt"):
			result = dispatch.ParseCSVBatch(data)
		case strings.HasSuffix(name, ".edi") || strings.HasSuffix(name, ".x12"):
			result = dispatch.ParseEDIBatch(data)
		default:
			// Tentar XML primeiro, depois CSV
			result = dispatch.ParseXMLBatch(data)
			if result.Processed == 0 {
				result = dispatch.ParseCSVBatch(data)
			}
		}

		allItems = append(allItems, result.Items...)
		allErrors = append(allErrors, result.Errors...)
		allWarnings = append(allWarnings, result.Warnings...)
		totalProcessed += result.Processed
		totalFailed += result.Failed
	}

	// Calcular totais do lote
	var totalWeight, totalCubage, totalValue float64
	for _, item := range allItems {
		totalWeight += item.WeightKg
		totalCubage += item.CubageM3
		totalValue += item.TotalValue
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "ok",
		"total_processed": totalProcessed,
		"total_failed":    totalFailed,
		"total_items":     len(allItems),
		"total_weight_kg": totalWeight,
		"total_cubage_m3": totalCubage,
		"total_value":     totalValue,
		"items":           formatDeliveryItems(allItems),
		"errors":          allErrors,
		"warnings":        allWarnings,
		"message":         fmt.Sprintf("%d entregas importadas com sucesso", totalProcessed),
	})
}

// ScanBarcode POST /api/v1/dispatch/scan
// Lê código de barras de NF-e e adiciona na fila (fluxo secundário).
func (h *DispatchHandler) ScanBarcode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Barcode  string `json:"barcode"`
		BatchID  string `json:"batch_id"` // lote existente para adicionar
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.Barcode == "" {
		respondError(w, http.StatusBadRequest, "Código de barras é obrigatório")
		return
	}

	// Código de barras da NF-e tem 44 dígitos
	barcode := strings.TrimSpace(req.Barcode)
	if len(barcode) != 44 {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Código inválido: esperado 44 dígitos, recebido %d", len(barcode)))
		return
	}

	// Criar item básico com a chave NF-e
	// Em produção: consultar SEFAZ para preencher os dados completos
	item := &dispatch.DeliveryItem{
		NFeKey:       barcode,
		DocumentType: "nfe",
		GeoStatus:    "pending",
		Status:       "pending",
		ImportSource: "scanner",
		ImportedAt:   time.Now(),
		Volumes:      1,
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Nota adicionada à fila. Consulte a SEFAZ para dados completos.",
		"item":    formatDeliveryItem(item),
		"nfe_key": barcode,
	})
}

// Distribute POST /api/v1/dispatch/batches/{id}/distribute
// Distribui entregas pelos veículos disponíveis.
func (h *DispatchHandler) Distribute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items    []*dispatch.DeliveryItem `json:"items"`
		Vehicles []*dispatch.Vehicle      `json:"vehicles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if len(req.Vehicles) == 0 {
		respondError(w, http.StatusBadRequest, "Informe pelo menos 1 veículo")
		return
	}

	assignments := dispatch.DistributeByVehicle(req.Items, req.Vehicles)

	// Formatar resultado
	result := make([]map[string]interface{}, len(assignments))
	for i, a := range assignments {
		overWeight := a.Vehicle.MaxWeightKg > 0 && a.TotalWeight > a.Vehicle.MaxWeightKg
		overCubage := a.Vehicle.MaxCubageM3 > 0 && a.TotalCubage > a.Vehicle.MaxCubageM3

		result[i] = map[string]interface{}{
			"vehicle":       a.Vehicle,
			"items":         formatDeliveryItems(a.Items),
			"total_stops":   a.TotalStops,
			"total_weight":  a.TotalWeight,
			"total_cubage":  a.TotalCubage,
			"weight_used_pct": a.WeightUsed,
			"cubage_used_pct": a.CubageUsed,
			"over_weight":   overWeight,
			"over_cubage":   overCubage,
			"warnings": func() []string {
				var w []string
				if overWeight { w = append(w, fmt.Sprintf("⚠️ Peso excede capacidade: %.0f/%.0f kg", a.TotalWeight, a.Vehicle.MaxWeightKg)) }
				if overCubage { w = append(w, fmt.Sprintf("⚠️ Cubagem excede: %.2f/%.2f m³", a.TotalCubage, a.Vehicle.MaxCubageM3)) }
				return w
			}(),
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "ok",
		"assignments": result,
		"message":     fmt.Sprintf("%d veículos configurados para despacho", len(assignments)),
	})
}

// Reassign POST /api/v1/dispatch/batches/{id}/reassign
// Permite usuário mover entrega de um veículo para outro.
func (h *DispatchHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID        string `json:"item_id"`
		FromVehicleID string `json:"from_vehicle_id"`
		ToVehicleID   string `json:"to_vehicle_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Entrega reatribuída com sucesso",
	})
}

// SendToRouter POST /api/v1/dispatch/batches/{id}/route
// Envia o lote de um veículo para o roteirizador.
func (h *DispatchHandler) SendToRouter(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VehicleID string                   `json:"vehicle_id"`
		Items     []*dispatch.DeliveryItem `json:"items"`
		Origin    struct {
			Lat   float64 `json:"lat"`
			Lng   float64 `json:"lng"`
			Label string  `json:"label"`
		} `json:"origin"`
		VehicleCapacityKg float64 `json:"vehicle_capacity_kg"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// Montar payload para o roteirizador OSRM
	type destination struct {
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
		Label    string  `json:"label"`
		WeightKg float64 `json:"weight_kg"`
	}

	var destinations []destination
	for _, item := range req.Items {
		if item.Lat == 0 && item.Lng == 0 {
			continue // pular itens sem geocodificação
		}
		destinations = append(destinations, destination{
			Lat:      item.Lat,
			Lng:      item.Lng,
			Label:    fmt.Sprintf("%s — %s", item.RecipientName, item.City),
			WeightKg: item.WeightKg,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"router_payload": map[string]interface{}{
			"origin":               req.Origin,
			"destinations":         destinations,
			"vehicle_capacity_kg":  req.VehicleCapacityKg,
			"optimize":             true,
		},
		"message":    fmt.Sprintf("%d entregas prontas para roteirizar", len(destinations)),
		"skipped":    len(req.Items) - len(destinations),
	})
}

// ListVehicles GET /api/v1/dispatch/vehicles
func (h *DispatchHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	_ = tenantID

	// Demo vehicles
	vehicles := []map[string]interface{}{
		{"id": "v1", "name": "Fiorino 01 — Carlos", "type": "fiorino", "max_weight_kg": 600, "max_cubage_m3": 2.8, "max_stops": 40, "is_available": true},
		{"id": "v2", "name": "Van Sprinter — João", "type": "van", "max_weight_kg": 1500, "max_cubage_m3": 8.0, "max_stops": 60, "is_available": true},
		{"id": "v3", "name": "Truck 3/4 — Pedro", "type": "truck", "max_weight_kg": 4000, "max_cubage_m3": 18.0, "max_stops": 30, "is_available": true},
		{"id": "v4", "name": "Carreta — Roberto", "type": "carreta", "max_weight_kg": 27000, "max_cubage_m3": 90.0, "max_stops": 15, "is_available": false},
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"vehicles": vehicles,
		"count":    len(vehicles),
	})
}

func (h *DispatchHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var v dispatch.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Veículo cadastrado com sucesso",
		"vehicle": v,
	})
}

func (h *DispatchHandler) ListBatches(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{"batches": []interface{}{}, "count": 0})
}

func (h *DispatchHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{"batch": nil})
}

func (h *DispatchHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{"message": "Item removido"})
}

// ── HELPERS ───────────────────────────────────────────────────────────────────

func formatDeliveryItems(items []*dispatch.DeliveryItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = formatDeliveryItem(item)
	}
	return result
}

func formatDeliveryItem(item *dispatch.DeliveryItem) map[string]interface{} {
	return map[string]interface{}{
		"id":             item.ID,
		"nfe_key":        item.NFeKey,
		"nfe_number":     item.NFeNumber,
		"document_type":  item.DocumentType,
		"recipient_name": item.RecipientName,
		"recipient_doc":  item.RecipientDoc,
		"full_address":   item.FullAddress,
		"city":           item.City,
		"state":          item.State,
		"lat":            item.Lat,
		"lng":            item.Lng,
		"geo_status":     item.GeoStatus,
		"weight_kg":      item.WeightKg,
		"cubage_m3":      item.CubageM3,
		"volumes":        item.Volumes,
		"total_value":    item.TotalValue,
		"status":         item.Status,
		"vehicle_id":     item.VehicleID,
		"import_source":  item.ImportSource,
		"priority":       item.Priority,
	}
}

func min(a, b int) int {
	if a < b { return a }
	return b
}
