package shoes

import "github.com/nexoflash/nexo-flash/internal/core"

type Module struct{}

func (m *Module) Init() error { return nil }
func (m *Module) GetPermissions() []string {
	return []string{"shoes:grid:read", "shoes:grid:write", "shoes:sales:read", "shoes:commission:read"}
}
func (m *Module) GetMenu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "Grades", Icon: "grid", Route: "/shoes/grids"},
		{Label: "Vendas", Icon: "shopping-bag", Route: "/shoes/sales"},
		{Label: "Comissões", Icon: "percent", Route: "/shoes/commissions"},
		{Label: "Estoque", Icon: "box", Route: "/shoes/stock"},
	}
}
func (m *Module) BusinessTypes() []string { return []string{"shoes"} }

func init() { core.RegisterModule("shoes", &Module{}) }
