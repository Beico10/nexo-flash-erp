// Package mechanic — registro do módulo no micro-kernel do Nexo One.
package mechanic

import "github.com/nexoone/nexo-one/internal/core"

// Module implementa a interface core.Module para o nicho de mecânica.
type Module struct{}

func (m *Module) Init() error {
	// Aqui: registrar rotas HTTP do módulo no router principal
	// ex: router.POST("/os", handlers.CreateOS)
	return nil
}

func (m *Module) GetPermissions() []string {
	return []string{
		"mechanic:os:read",
		"mechanic:os:write",
		"mechanic:os:approve",
		"mechanic:parts:read",
		"mechanic:parts:write",
		"mechanic:whatsapp:send",
	}
}

func (m *Module) GetMenu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "Ordens de Serviço", Icon: "wrench", Route: "/mechanic/os"},
		{Label: "Peças", Icon: "box", Route: "/mechanic/parts"},
		{Label: "Veículos", Icon: "car", Route: "/mechanic/vehicles"},
	}
}

func (m *Module) BusinessTypes() []string {
	return []string{"mechanic"}
}

// Registra automaticamente ao importar o pacote
func init() {
	core.RegisterModule("mechanic", &Module{})
}
