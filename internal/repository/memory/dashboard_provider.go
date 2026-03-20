package memory

import (
	"github.com/nexoone/nexo-one/internal/handlers"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
)

// DashboardProvider agrega dados de todos os repos in-memory.
type DashboardProvider struct {
	mechanicRepo   *MechanicRepo
	bakeryRepo     *BakeryRepo
	aestheticsRepo *AestheticsRepo
	aiRepo         *AIRepo
}

func NewDashboardProvider(m *MechanicRepo, b *BakeryRepo, a *AestheticsRepo, ai *AIRepo) *DashboardProvider {
	return &DashboardProvider{mechanicRepo: m, bakeryRepo: b, aestheticsRepo: a, aiRepo: ai}
}

func (p *DashboardProvider) GetDashboardStats(tenantID string) handlers.DashboardStats {
	p.mechanicRepo.mu.RLock()
	mTotal, mOpen, mProg, mAwait, mDone := 0, 0, 0, 0, 0
	for _, os := range p.mechanicRepo.orders {
		if os.TenantID == tenantID {
			mTotal++
			switch os.Status {
			case mechanic.OSStatusOpen:
				mOpen++
			case mechanic.OSStatusInProgress:
				mProg++
			case mechanic.OSStatusAwaitApproval:
				mAwait++
			case mechanic.OSStatusDone, mechanic.OSStatusInvoiced:
				mDone++
			}
		}
	}
	p.mechanicRepo.mu.RUnlock()

	p.bakeryRepo.mu.RLock()
	bakeryProducts := 0
	for _, prod := range p.bakeryRepo.products {
		if prod.TenantID == tenantID && prod.Active {
			bakeryProducts++
		}
	}
	p.bakeryRepo.mu.RUnlock()

	p.aestheticsRepo.mu.RLock()
	appointments := 0
	for _, apt := range p.aestheticsRepo.apts {
		if apt.TenantID == tenantID {
			appointments++
		}
	}
	p.aestheticsRepo.mu.RUnlock()

	p.aiRepo.mu.RLock()
	pending := 0
	for _, s := range p.aiRepo.suggestions {
		if s.TenantID == tenantID && s.Status == "pending" {
			pending++
		}
	}
	p.aiRepo.mu.RUnlock()

	chart := []handlers.DashboardDay{
		{Day: "Seg", Revenue: 4200, Tax: 546},
		{Day: "Ter", Revenue: 5800, Tax: 754},
		{Day: "Qua", Revenue: 3900, Tax: 507},
		{Day: "Qui", Revenue: 6700, Tax: 871},
		{Day: "Sex", Revenue: 7200, Tax: 936},
		{Day: "Sab", Revenue: 8100, Tax: 1053},
		{Day: "Dom", Revenue: 5400, Tax: 702},
	}

	return handlers.DashboardStats{
		MechanicOS: handlers.ModuleStats{
			Total: mTotal, Open: mOpen, InProgress: mProg, AwaitApproval: mAwait, Done: mDone,
		},
		BakeryProducts:  bakeryProducts,
		Appointments:    appointments,
		PendingSugg:     pending,
		Revenue: handlers.DashboardRevenue{
			Today: 8100, Week: 41300,
			Chart: chart,
			ByModule: []handlers.ModuleActivity{
				{Module: "Mecanica", Count: mTotal},
				{Module: "Padaria", Count: bakeryProducts},
				{Module: "Estetica", Count: appointments},
			},
		},
	}
}
