// Package dispatch — tipos e funções para despacho em lote.
package dispatch

import "time"

// DeliveryItem representa uma entrega individual.
type DeliveryItem struct {
	ID            string    `json:"id"`
	NFeKey        string    `json:"nfe_key"`
	NFeNumber     string    `json:"nfe_number"`
	DocumentType  string    `json:"document_type"` // nfe, cte, mdfe
	RecipientName string    `json:"recipient_name"`
	RecipientDoc  string    `json:"recipient_doc"` // CPF/CNPJ
	FullAddress   string    `json:"full_address"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	ZipCode       string    `json:"zip_code"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	GeoStatus     string    `json:"geo_status"` // pending, geocoded, failed
	WeightKg      float64   `json:"weight_kg"`
	CubageM3      float64   `json:"cubage_m3"`
	Volumes       int       `json:"volumes"`
	TotalValue    float64   `json:"total_value"`
	Status        string    `json:"status"` // pending, assigned, dispatched, delivered
	VehicleID     string    `json:"vehicle_id"`
	ImportSource  string    `json:"import_source"` // xml, csv, edi, scanner
	ImportedAt    time.Time `json:"imported_at"`
	Priority      int       `json:"priority"` // 1=alta, 2=media, 3=baixa
}

// Vehicle representa um veículo de entrega.
type Vehicle struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // fiorino, van, truck, carreta
	MaxWeightKg float64 `json:"max_weight_kg"`
	MaxCubageM3 float64 `json:"max_cubage_m3"`
	MaxStops    int     `json:"max_stops"`
	IsAvailable bool    `json:"is_available"`
}

// VehicleAssignment representa entregas atribuídas a um veículo.
type VehicleAssignment struct {
	Vehicle     *Vehicle
	Items       []*DeliveryItem
	TotalStops  int
	TotalWeight float64
	TotalCubage float64
	WeightUsed  float64 // percentual
	CubageUsed  float64 // percentual
}

// ParseResult resultado do parsing de arquivo.
type ParseResult struct {
	Items     []*DeliveryItem
	Processed int
	Failed    int
	Errors    []string
	Warnings  []string
}
