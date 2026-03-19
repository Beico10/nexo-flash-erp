package logistics

import "github.com/nexoone/nexo-one/internal/core"

type Module struct{}

func (m *Module) Init() error { return nil }
func (m *Module) GetPermissions() []string {
	return []string{
		"logistics:cte:read", "logistics:cte:write", "logistics:cte:issue",
		"logistics:contracts:read", "logistics:contracts:write",
		"logistics:routes:read", "logistics:routes:write",
		"logistics:fleet:read",
	}
}
func (m *Module) GetMenu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "CT-e / MDF-e", Icon: "truck", Route: "/logistics/cte"},
		{Label: "Contratos", Icon: "file-text", Route: "/logistics/contracts"},
		{Label: "Roteirização", Icon: "map", Route: "/logistics/routes"},
		{Label: "Frota", Icon: "settings", Route: "/logistics/fleet"},
	}
}
func (m *Module) BusinessTypes() []string { return []string{"logistics"} }

func init() { core.RegisterModule("logistics", &Module{}) }
