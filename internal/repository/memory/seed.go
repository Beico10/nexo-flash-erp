package memory

import (
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/ai"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
	"github.com/nexoone/nexo-one/internal/modules/bakery"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
)

const demoTenantID = "00000000-0000-0000-0000-000000000001"

func SeedAllDemoData(
	mechRepo *MechanicRepo,
	bakRepo *BakeryRepo,
	aesRepo *AestheticsRepo,
	aiRepo *AIRepo,
) {
	seedMechanicDemo(mechRepo)
	seedBakeryDemo(bakRepo)
	seedAestheticsDemo(aesRepo)
	seedAIDemo(aiRepo)
}

func seedMechanicDemo(r *MechanicRepo) {
	now := time.Now().UTC()
	orders := []*mechanic.ServiceOrder{
		{ID: uuid.New().String(), TenantID: demoTenantID, Number: "OS-2026-001842", VehiclePlate: "BRA2E19", VehicleKM: 45200, VehicleModel: "Civic 2021", VehicleYear: 2021, CustomerID: "Carlos Silva", CustomerPhone: "5511999887766", Status: mechanic.OSStatusAwaitApproval, Complaint: "Barulho no freio dianteiro", Diagnosis: "Pastilha e disco gastos", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, Number: "OS-2026-001841", VehiclePlate: "ABC1D23", VehicleKM: 78000, VehicleModel: "HB20 2019", VehicleYear: 2019, CustomerID: "Maria Santos", CustomerPhone: "5511988776655", Status: mechanic.OSStatusInProgress, Complaint: "Motor falhando em baixa rotacao", Diagnosis: "Bobina de ignicao com defeito", CreatedAt: now.Add(-5 * time.Hour), UpdatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, Number: "OS-2026-001840", VehiclePlate: "XYZ9K87", VehicleKM: 32000, VehicleModel: "Onix 2023", VehicleYear: 2023, CustomerID: "Joao Lima", CustomerPhone: "5511977665544", Status: mechanic.OSStatusDone, Complaint: "Revisao 30.000km", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, Number: "OS-2026-001839", VehiclePlate: "DEF4G56", VehicleKM: 95000, VehicleModel: "Corolla 2018", VehicleYear: 2018, CustomerID: "Ana Costa", CustomerPhone: "5511966554433", Status: mechanic.OSStatusOpen, Complaint: "Ar condicionado nao gela", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, Number: "OS-2026-001838", VehiclePlate: "GHI7J89", VehicleKM: 120000, VehicleModel: "Strada 2020", VehicleYear: 2020, CustomerID: "Pedro Rocha", CustomerPhone: "5511955443322", Status: mechanic.OSStatusOpen, Complaint: "Vazamento de oleo no motor", CreatedAt: now.Add(-30 * time.Minute), UpdatedAt: now},
	}
	r.mu.Lock()
	for _, os := range orders {
		r.orders[os.ID] = os
	}
	r.mu.Unlock()
}

func seedBakeryDemo(r *BakeryRepo) {
	products := []*bakery.BakeryProduct{
		{ID: uuid.New().String(), TenantID: demoTenantID, SKU: "PAO-001", Name: "Pao Frances", SaleType: bakery.SaleByWeight, UnitPrice: 8.50, NCMCode: "19052000", IsBasketItem: true, BasketCategory: "pao_frances", ScaleCode: "P001", CurrentStock: 45.2, MinStock: 10, Active: true},
		{ID: uuid.New().String(), TenantID: demoTenantID, SKU: "BOL-001", Name: "Bolo de Cenoura", SaleType: bakery.SaleByUnit, UnitPrice: 35.00, NCMCode: "19053100", ScaleCode: "B001", CurrentStock: 8, MinStock: 3, Active: true},
		{ID: uuid.New().String(), TenantID: demoTenantID, SKU: "CRO-001", Name: "Croissant", SaleType: bakery.SaleByUnit, UnitPrice: 7.50, NCMCode: "19059090", ScaleCode: "C001", CurrentStock: 24, MinStock: 10, Active: true},
		{ID: uuid.New().String(), TenantID: demoTenantID, SKU: "PAO-002", Name: "Pao de Queijo", SaleType: bakery.SaleByUnit, UnitPrice: 4.00, NCMCode: "19052000", IsBasketItem: true, BasketCategory: "pao_queijo", ScaleCode: "P002", CurrentStock: 60, MinStock: 20, Active: true},
		{ID: uuid.New().String(), TenantID: demoTenantID, SKU: "TOR-001", Name: "Torta de Frango", SaleType: bakery.SaleByUnit, UnitPrice: 12.00, NCMCode: "19059090", ScaleCode: "T001", CurrentStock: 15, MinStock: 5, Active: true},
		{ID: uuid.New().String(), TenantID: demoTenantID, SKU: "BRI-001", Name: "Brioche", SaleType: bakery.SaleByUnit, UnitPrice: 9.00, NCMCode: "19059090", ScaleCode: "B002", CurrentStock: 18, MinStock: 8, Active: true},
	}
	r.mu.Lock()
	for _, p := range products {
		r.products[p.ID] = p
	}
	r.mu.Unlock()
}

func seedAestheticsDemo(r *AestheticsRepo) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	apts := []*aesthetics.Appointment{
		{ID: uuid.New().String(), TenantID: demoTenantID, ProfessionalID: "prof-1", CustomerName: "Maria Santos", ServiceName: "Coloracao completa", StartTime: today.Add(9 * time.Hour), EndTime: today.Add(11 * time.Hour), ServicePrice: 280, DurationMin: 120, Status: aesthetics.AppointmentConfirmed, CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, ProfessionalID: "prof-1", CustomerName: "Joana Pereira", ServiceName: "Corte + escova", StartTime: today.Add(11*time.Hour + 30*time.Minute), EndTime: today.Add(13 * time.Hour), ServicePrice: 120, DurationMin: 90, Status: aesthetics.AppointmentScheduled, CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, ProfessionalID: "prof-2", CustomerName: "Fernanda Costa", ServiceName: "Manicure + pedicure", StartTime: today.Add(9 * time.Hour), EndTime: today.Add(10*time.Hour + 30*time.Minute), ServicePrice: 85, DurationMin: 90, Status: aesthetics.AppointmentInProgress, CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, ProfessionalID: "prof-2", CustomerName: "Paula Rodrigues", ServiceName: "Design de sobrancelha", StartTime: today.Add(11 * time.Hour), EndTime: today.Add(12 * time.Hour), ServicePrice: 60, DurationMin: 60, Status: aesthetics.AppointmentScheduled, CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, ProfessionalID: "prof-3", CustomerName: "Leticia Alves", ServiceName: "Hidratacao", StartTime: today.Add(10 * time.Hour), EndTime: today.Add(11*time.Hour + 30*time.Minute), ServicePrice: 95, DurationMin: 90, Status: aesthetics.AppointmentConfirmed, CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, ProfessionalID: "prof-3", CustomerName: "Camila Torres", ServiceName: "Limpeza de pele", StartTime: today.Add(13 * time.Hour), EndTime: today.Add(14 * time.Hour), ServicePrice: 110, DurationMin: 60, Status: aesthetics.AppointmentScheduled, CreatedAt: now},
	}
	r.mu.Lock()
	for _, a := range apts {
		r.apts[a.ID] = a
	}
	r.mu.Unlock()
}

func seedAIDemo(r *AIRepo) {
	now := time.Now().UTC()
	suggestions := []*ai.Suggestion{
		{ID: uuid.New().String(), TenantID: demoTenantID, Type: "missing_labor_cost", TargetTable: "mechanic_os", TargetID: "OS-2026-001842", Reason: "OS contem 3 pecas (pastilha, disco, fluido) mas nenhum item de mao de obra foi registrado.", Confidence: 0.94, CreatedByAI: "co-pilot-v1", Status: "pending", CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, Type: "ncm_correction", TargetTable: "products", TargetID: "Pastilha de Freio Dianteira", Reason: "NCM cadastrado (84714900) parece incorreto para este produto. Sugestao: 87083000.", Confidence: 0.87, CreatedByAI: "concierge-v1", Status: "pending", CreatedAt: now},
		{ID: uuid.New().String(), TenantID: demoTenantID, Type: "onboard_field", TargetTable: "products", TargetID: "NF-e 001.234 Auto Pecas Silva", Reason: "12 produtos detectados no XML da NF-e de compra. Nenhum deles consta no catalogo atual.", Confidence: 0.92, CreatedByAI: "concierge-v1", Status: "pending", CreatedAt: now},
	}
	r.mu.Lock()
	for _, s := range suggestions {
		r.suggestions[s.ID] = s
	}
	r.mu.Unlock()
}
