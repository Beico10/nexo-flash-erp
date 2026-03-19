// Package ai implementa a camada Human-in-the-Loop do Nexo One.
//
// DIRETRIZ CRÍTICA (do Briefing Mestre):
// Nenhuma IA tem autonomia para alterar dados (Preço, Imposto, Financeiro)
// sem autorização humana explícita.
//
// Fluxo obrigatório:
//  1. IA gera uma sugestão → cria ai_suggestions com status='pending'
//  2. Usuário vê o card de aprovação na interface
//  3. Usuário clica "Aprovar" → ApproveHandler executa a ação e persiste
//  4. Usuário clica "Rejeitar" → sugestão arquivada, nenhum dado alterado
//
// A IA NUNCA escreve diretamente em tabelas de negócio.
package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SuggestionType define os tipos de sugestão que a IA pode gerar.
type SuggestionType string

const (
	SuggestionMissingLaborCost SuggestionType = "missing_labor_cost"    // "Faltou mão de obra nesta OS"
	SuggestionNCMCorrection    SuggestionType = "ncm_correction"        // NCM possivelmente errado
	SuggestionPriceAnomaly     SuggestionType = "price_anomaly"         // preço fora da faixa histórica
	SuggestionStockLow         SuggestionType = "stock_low_reorder"     // sugerir reposição
	SuggestionRouteOptimize    SuggestionType = "route_optimization"    // rota mais eficiente
	SuggestionOnboardField     SuggestionType = "onboard_field"         // campo detectado no XML de NF
)

// Suggestion representa uma sugestão pendente de aprovação humana.
// Status sempre começa em 'pending' — NUNCA persiste sem aprovação.
type Suggestion struct {
	ID             string
	TenantID       string
	Type           SuggestionType
	TargetTable    string         // tabela que seria alterada
	TargetID       string         // ID do registro alvo (se aplicável)
	SuggestedData  map[string]any // dados propostos — a aplicar se aprovado
	Reason         string         // explicação legível pelo usuário
	Confidence     float64        // 0.0 a 1.0
	CreatedByAI    string         // nome do agente ("co-pilot","concierge")
	CreatedAt      time.Time
	ExpiresAt      time.Time
	Status         string         // "pending"|"approved"|"rejected"|"expired"
}

// ApprovalRepository persiste e busca sugestões.
type ApprovalRepository interface {
	CreatePending(ctx context.Context, s *Suggestion) error
	GetPending(ctx context.Context, tenantID string) ([]*Suggestion, error)
	Approve(ctx context.Context, suggestionID, approvedByUserID string) error
	Reject(ctx context.Context, suggestionID, userID, reason string) error
}

// ActionExecutor executa a ação real após aprovação humana.
// Cada módulo registra seu executor para seu tipo de sugestão.
type ActionExecutor interface {
	Execute(ctx context.Context, s *Suggestion) error
	CanHandle(t SuggestionType) bool
}

// Gateway é o ponto de entrada para toda ação de IA no sistema.
// A IA chama apenas Gateway.Suggest() — nunca acessa o banco diretamente.
type Gateway struct {
	repo      ApprovalRepository
	executors []ActionExecutor
}

// NewGateway cria o gateway de aprovação humana.
func NewGateway(repo ApprovalRepository, executors ...ActionExecutor) *Gateway {
	return &Gateway{repo: repo, executors: executors}
}

// Suggest é o único método que a IA pode chamar.
// Cria um registro PENDING — nenhum dado de negócio é alterado ainda.
func (g *Gateway) Suggest(ctx context.Context, s *Suggestion) error {
	if s.TenantID == "" || s.Type == "" || s.Reason == "" {
		return fmt.Errorf("ai.Gateway.Suggest: TenantID, Type e Reason são obrigatórios")
	}

	s.ID = uuid.New().String()
	s.Status = "pending"
	s.CreatedAt = time.Now().UTC()
	s.ExpiresAt = s.CreatedAt.Add(7 * 24 * time.Hour)

	if err := g.repo.CreatePending(ctx, s); err != nil {
		return fmt.Errorf("ai.Gateway.Suggest: falha ao criar sugestão: %w", err)
	}

	// Aqui publicamos no event bus para notificar a interface em tempo real
	// eventbus.Publish(ctx, s.TenantID, EventAISuggestion, s)
	return nil
}

// Approve executa a ação aprovada pelo usuário humano.
// Esta é a ÚNICA forma de uma sugestão de IA persistir no banco.
func (g *Gateway) Approve(ctx context.Context, suggestionID, userID string) error {
	if err := g.repo.Approve(ctx, suggestionID, userID); err != nil {
		return fmt.Errorf("ai.Gateway.Approve: %w", err)
	}
	// buscar a sugestão aprovada e executar
	// implementação completa: g.repo.GetByID(ctx, suggestionID) → executor.Execute()
	return nil
}

// Reject descarta a sugestão sem nenhum efeito nos dados.
func (g *Gateway) Reject(ctx context.Context, suggestionID, userID, reason string) error {
	return g.repo.Reject(ctx, suggestionID, userID, reason)
}

// GetPendingForUser retorna todas as sugestões pendentes de aprovação.
// Usado para renderizar os cards de aprovação na interface.
func (g *Gateway) GetPendingForUser(ctx context.Context, tenantID string) ([]*Suggestion, error) {
	return g.repo.GetPending(ctx, tenantID)
}
