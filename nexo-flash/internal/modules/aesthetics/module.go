package aesthetics

import "github.com/nexoflash/nexo-flash/internal/core"

type Module struct{}

func (m *Module) Init() error { return nil }
func (m *Module) GetPermissions() []string {
	return []string{"aesthetics:agenda:read", "aesthetics:agenda:write", "aesthetics:split:manage"}
}
func (m *Module) GetMenu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "Agenda", Icon: "calendar", Route: "/aesthetics/agenda"},
		{Label: "Serviços", Icon: "scissors", Route: "/aesthetics/services"},
		{Label: "Profissionais", Icon: "users", Route: "/aesthetics/professionals"},
		{Label: "Split de Pagamento", Icon: "divide", Route: "/aesthetics/split"},
	}
}
func (m *Module) BusinessTypes() []string { return []string{"aesthetics"} }

func init() { core.RegisterModule("aesthetics", &Module{}) }
