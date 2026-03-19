// Package logistics implementa a lógica de contratos multi-cliente e
// o "DRE da Viagem" (Demonstrativo de Resultado da Viagem) do Nexo Flash.
//
// Hierarquia de contratos (Briefing Mestre §4):
//  1. Tabela específica do Embarcador (shipper_id preenchido) → PRIORIDADE
//  2. Tabela geral da transportadora (shipper_id = NULL) → fallback
//
// O DRE da Viagem mostra o lucro estimado ANTES da partida, exibindo:
//   - Receita bruta (fretes)
//   - Custos variáveis (combustível, pedágios, motorista)
//   - Resultado estimado (lucro/prejuízo)
package logistics

import (
	"context"
	"fmt"
	"math"
)

// VehicleType define os tipos de veículo com pricing diferenciado.
type VehicleType string

const (
	VehicleGeneral VehicleType = "general"
	VehicleVUC     VehicleType = "vuc"     // Veículo Urbano de Carga (restrições urbanas)
	VehicleTruck   VehicleType = "truck"
	VehicleCarreta VehicleType = "carreta"
	VehicleVan     VehicleType = "van"
)

// Contract representa um contrato de frete.
type Contract struct {
	ID            string
	TenantID      string
	ContractName  string
	ShipperID     *string     // nil = tabela geral; preenchido = específico do embarcador
	VehicleType   VehicleType
	PricePerKM    float64
	PricePerKG    float64
	MinimumCharge float64
	TollPolicy    string // "included" | "excluded" | "split"
}

// ContractRepository busca contratos do banco.
type ContractRepository interface {
	// GetApplicable retorna o contrato mais específico para o par (embarcador, veículo).
	// Implementa a hierarquia: embarcador específico > tabela geral.
	GetApplicable(ctx context.Context, tenantID, shipperID string, vt VehicleType) (*Contract, error)
}

// FreightInput são os dados para calcular o frete e o DRE da Viagem.
type FreightInput struct {
	TenantID   string
	ShipperID  string
	VehicleType VehicleType
	DistanceKM float64
	WeightKG   float64
	TollCost   float64  // custo real de pedágios da rota
	FuelCostPerKM float64
	DriverCostPerKM float64
}

// FreightResult é o resultado do cálculo de frete.
type FreightResult struct {
	AppliedContractID   string
	AppliedContractName string
	IsShipperSpecific   bool    // true = contrato específico do embarcador
	GrossRevenue        float64 // receita bruta do frete
	TollRevenue         float64 // repasse de pedágios (se TollPolicy=excluded)
	FuelCost            float64
	DriverCost          float64
	TollCost            float64
	TotalCost           float64
	EstimatedProfit     float64
	ProfitMarginPct     float64
	// DRE da Viagem — campos para exibição antes da partida
	DRE DREViagem
}

// DREViagem é o Demonstrativo de Resultado da Viagem.
// Exibido para o motorista/gestor antes da partida.
type DREViagem struct {
	Receita       LineItem
	CustoTotal    LineItem
	LucroEstimado LineItem
	Margem        string
	Viavel        bool // false = prejuízo estimado → alerta ao gestor
}

type LineItem struct {
	Label string
	Value float64
}

// ContractService resolve contratos e calcula fretes.
type ContractService struct {
	repo ContractRepository
}

func NewContractService(r ContractRepository) *ContractService {
	return &ContractService{repo: r}
}

// Calculate resolve o contrato aplicável e calcula o DRE da Viagem.
func (s *ContractService) Calculate(ctx context.Context, input FreightInput) (*FreightResult, error) {
	if err := validateFreightInput(input); err != nil {
		return nil, fmt.Errorf("logistics.Calculate: %w", err)
	}

	contract, err := s.repo.GetApplicable(ctx, input.TenantID, input.ShipperID, input.VehicleType)
	if err != nil {
		return nil, fmt.Errorf("logistics.Calculate: contrato não encontrado: %w", err)
	}

	// Calcular receita
	freightByKM := contract.PricePerKM * input.DistanceKM
	freightByKG := contract.PricePerKG * input.WeightKG
	grossRevenue := math.Max(freightByKM+freightByKG, contract.MinimumCharge)

	// Pedágio conforme política do contrato
	var tollRevenue, tollCost float64
	tollCost = input.TollCost
	switch contract.TollPolicy {
	case "excluded":
		tollRevenue = tollCost // repassa o custo ao cliente
	case "split":
		tollRevenue = tollCost * 0.5 // divide o custo
		tollCost = tollCost * 0.5
	case "included":
		// pedágio já está no preço — sem repasse adicional
	}

	// Custos variáveis
	fuelCost := input.FuelCostPerKM * input.DistanceKM
	driverCost := input.DriverCostPerKM * input.DistanceKM
	totalCost := fuelCost + driverCost + tollCost

	// DRE
	totalRevenue := grossRevenue + tollRevenue
	profit := totalRevenue - totalCost
	var marginPct float64
	if totalRevenue > 0 {
		marginPct = (profit / totalRevenue) * 100
	}

	dre := DREViagem{
		Receita:       LineItem{"Receita bruta de frete", round2(totalRevenue)},
		CustoTotal:    LineItem{"Custo total da viagem", round2(totalCost)},
		LucroEstimado: LineItem{"Lucro estimado", round2(profit)},
		Margem:        fmt.Sprintf("%.1f%%", marginPct),
		Viavel:        profit >= 0,
	}

	return &FreightResult{
		AppliedContractID:   contract.ID,
		AppliedContractName: contract.ContractName,
		IsShipperSpecific:   contract.ShipperID != nil,
		GrossRevenue:        round2(grossRevenue),
		TollRevenue:         round2(tollRevenue),
		FuelCost:            round2(fuelCost),
		DriverCost:          round2(driverCost),
		TollCost:            round2(tollCost),
		TotalCost:           round2(totalCost),
		EstimatedProfit:     round2(profit),
		ProfitMarginPct:     round2(marginPct),
		DRE:                 dre,
	}, nil
}

func validateFreightInput(i FreightInput) error {
	if i.TenantID == "" {
		return fmt.Errorf("TenantID obrigatório")
	}
	if i.DistanceKM <= 0 {
		return fmt.Errorf("DistanceKM deve ser > 0")
	}
	return nil
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
