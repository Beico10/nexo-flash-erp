package industry

import "github.com/nexoone/nexo-one/internal/core"

type Module struct{}

func (m *Module) Init() error { return nil }
func (m *Module) GetPermissions() []string {
	return []string{"industry:bom:read", "industry:bom:write", "industry:pcp:read", "industry:pcp:write", "industry:materials:read"}
}
func (m *Module) GetMenu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "Produção (PCP)", Icon: "factory", Route: "/industry/pcp"},
		{Label: "Fichas Técnicas", Icon: "clipboard", Route: "/industry/bom"},
		{Label: "Insumos", Icon: "package", Route: "/industry/materials"},
	}
}
func (m *Module) BusinessTypes() []string { return []string{"industry"} }

func init() { core.RegisterModule("industry", &Module{}) }
