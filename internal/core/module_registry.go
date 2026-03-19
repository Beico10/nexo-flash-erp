// Package core defines o micro-kernel do Nexo One.
// O ModuleRegistry carrega apenas os módulos autorizados para cada tenant,
// baseado no campo business_type da tabela tenants.
// Nenhuma lógica de nicho deve residir aqui — apenas contratos (interfaces).
package core

import (
	"fmt"
	"sync"
)

// MenuItem representa um item no menu lateral gerado dinamicamente por cada módulo.
type MenuItem struct {
	Label    string
	Icon     string
	Route    string
	Children []MenuItem
}

// Module é o contrato que todo nicho deve assinar.
// Qualquer novo módulo (futuro) precisa implementar esta interface
// e se registrar via RegisterModule — zero alterações no core.
type Module interface {
	// Init inicializa rotas, handlers e workers do módulo.
	Init() error
	// GetPermissions retorna as permissões RBAC declaradas pelo módulo.
	GetPermissions() []string
	// GetMenu retorna os itens de menu que este módulo adiciona ao sidebar.
	GetMenu() []MenuItem
	// BusinessTypes retorna os tipos de negócio que ativam este módulo.
	BusinessTypes() []string
}

// ModuleRegistry gerencia todos os módulos registrados no sistema.
// Thread-safe para uso concorrente durante hot-reload futuro.
type ModuleRegistry struct {
	mu      sync.RWMutex
	modules map[string]Module // key = nome do módulo
}

var globalRegistry = &ModuleRegistry{
	modules: make(map[string]Module),
}

// RegisterModule registra um módulo globalmente.
// Deve ser chamado em init() de cada pacote de módulo.
func RegisterModule(name string, m Module) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if _, exists := globalRegistry.modules[name]; exists {
		panic(fmt.Sprintf("módulo '%s' já registrado — nome duplicado", name))
	}
	globalRegistry.modules[name] = m
}

// ModulesForTenant retorna apenas os módulos habilitados para o business_type do tenant.
func ModulesForTenant(businessType string) []Module {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	var result []Module
	for _, m := range globalRegistry.modules {
		for _, bt := range m.BusinessTypes() {
			if bt == businessType || bt == "*" {
				result = append(result, m)
				break
			}
		}
	}
	return result
}

// InitModulesForTenant inicializa todos os módulos do tenant.
func InitModulesForTenant(businessType string) error {
	for _, m := range ModulesForTenant(businessType) {
		if err := m.Init(); err != nil {
			return fmt.Errorf("falha ao inicializar módulo: %w", err)
		}
	}
	return nil
}
