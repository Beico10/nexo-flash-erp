package bakery

import "github.com/nexoone/nexo-one/internal/core"

type Module struct{}

func (m *Module) Init() error          { return nil }
func (m *Module) GetPermissions() []string {
	return []string{"bakery:pdv:sell", "bakery:products:write", "bakery:loss:write", "bakery:reports:read"}
}
func (m *Module) GetMenu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "PDV Rápido", Icon: "cash-register", Route: "/bakery/pdv"},
		{Label: "Produtos", Icon: "bread", Route: "/bakery/products"},
		{Label: "Perdas", Icon: "trash", Route: "/bakery/losses"},
		{Label: "Estoque", Icon: "box", Route: "/bakery/stock"},
	}
}
func (m *Module) BusinessTypes() []string { return []string{"bakery"} }

func init() { core.RegisterModule("bakery", &Module{}) }
