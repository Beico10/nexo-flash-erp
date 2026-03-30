'use client'
import { useState, useEffect } from 'react'
import { Brain, CheckCircle, ThumbsUp, ThumbsDown, AlertTriangle } from 'lucide-react'
import { getBusinessType, promptLogin, isDemoMode } from '@/lib/demo'

const SUGGESTIONS_BY_NICHO: Record<string, any[]> = {
  mechanic: [
    { ID: 'ai-001', Type: 'missing_labor_cost', Reason: 'Detectei que a OS do Honda Civic nao tem custo de mao de obra registrado. Valor sugerido: R$ 120,00 baseado em servicos similares.', Confidence: 0.92, CreatedByAI: 'Copiloto IA' },
    { ID: 'ai-002', Type: 'ncm_correction', Reason: 'O NCM 8708.30.19 esta incorreto para pastilhas de freio. NCM correto: 8708.30.90 com aliquota IBS/CBS menor.', Confidence: 0.87, CreatedByAI: 'Motor Fiscal IA' },
    { ID: 'ai-003', Type: 'onboard_field', Reason: 'Encontrei nota fiscal de pecas do fornecedor AutoPecas Silva que pode ser vinculada a OS-2024-003 automaticamente.', Confidence: 0.78, CreatedByAI: 'Copiloto IA' },
  ],
  bakery: [
    { ID: 'ai-001', Type: 'validade_proxima', Reason: 'Detectei 3 produtos com validade proxima nos proximos 2 dias: Croissant, Cuca de Banana e Bolo de Chocolate. Sugiro promocao de 20% de desconto.', Confidence: 0.95, CreatedByAI: 'Copiloto IA' },
    { ID: 'ai-002', Type: 'ncm_correction', Reason: 'O NCM do Pao Frances esta incorreto. NCM correto: 1905.90.90 com aliquota IBS/CBS zero por ser item de cesta basica.', Confidence: 0.91, CreatedByAI: 'Motor Fiscal IA' },
    { ID: 'ai-003', Type: 'estoque_minimo', Reason: 'Estoque de Farinha esta abaixo do minimo. Com base na producao da semana passada, sugiro pedido de 50kg ao fornecedor Moinho Sul.', Confidence: 0.88, CreatedByAI: 'Copiloto IA' },
  ],
  aesthetics: [
    { ID: 'ai-001', Type: 'agenda_vazia', Reason: 'Detectei 3 horarios vagos amanha entre 14h e 17h. Sugiro enviar mensagem WhatsApp para 5 clientes que nao agendam ha mais de 30 dias.', Confidence: 0.93, CreatedByAI: 'Copiloto IA' },
    { ID: 'ai-002', Type: 'ncm_correction', Reason: 'O servico de Progressiva pode ter retencao de ISS diferenciada no municipio. Sugiro revisar aliquota aplicada nas ultimas NFS-e emitidas.', Confidence: 0.85, CreatedByAI: 'Motor Fiscal IA' },
    { ID: 'ai-003', Type: 'produto_faltando', Reason: 'Estoque de Botox Capilar esta zerado e ha 2 agendamentos de Progressiva para a proxima semana. Sugiro reposicao urgente.', Confidence: 0.97, CreatedByAI: 'Copiloto IA' },
  ],
  logistics: [
    { ID: 'ai-001', Type: 'rota_otimizada', Reason: 'Detectei que as entregas de amanha podem ser otimizadas reduzindo 47km no total. Sugiro reagrupar as paradas por regiao sul primeiro.', Confidence: 0.91, CreatedByAI: 'Copiloto IA' },
    { ID: 'ai-002', Type: 'ncm_correction', Reason: 'O CT-e 000123 tem CFOP incorreto para operacao interestadual. CFOP correto: 6.354 com credito de ICMS aproveitavel.', Confidence: 0.89, CreatedByAI: 'Motor Fiscal IA' },
    { ID: 'ai-003', Type: 'veiculo_manutencao', Reason: 'Veiculo placa ABC-1234 atingiu 90.000km. Revisao preventiva recomendada antes das entregas da proxima semana.', Confidence: 0.94, CreatedByAI: 'Copiloto IA' },
  ],
  industry: [
    { ID: 'ai-001', Type: 'materia_prima', Reason: 'Estoque de aco carbono esta para acabar em 3 dias com base na producao atual. Sugiro pedido urgente ao fornecedor Metalurgica Norte.', Confidence: 0.96, CreatedByAI: 'Copiloto IA' },
    { ID: 'ai-002', Type: 'ncm_correction', Reason: 'O produto ID-447 tem NCM desatualizado para 2026. NCM correto gera credito IBS de R$ 340,00 por lote produzido.', Confidence: 0.92, CreatedByAI: 'Motor Fiscal IA' },
    { ID: 'ai-003', Type: 'pedido_atrasado', Reason: 'Pedido PED-2024-089 esta 2 dias atrasado. Cliente Construtora ABC aguarda entrega. Sugiro priorizar na linha de producao.', Confidence: 0.88, CreatedByAI: 'Copiloto IA' },
  ],
  shoes: [
    { ID: 'ai-001', Type: 'grade_incompleta', Reason: 'Tenis Running Pro tem grade incompleta — faltam os numeros 38 e 39 que sao os mais vendidos. Sugiro reposicao prioritaria.', Confidence: 0.94, CreatedByAI: 'Copiloto IA' },
    { ID: 'ai-002', Type: 'ncm_correction', Reason: 'Sandalia Feminina Verao tem NCM incorreto. NCM 6402.99.90 tem aliquota IBS/CBS menor e gera economia de R$ 8,50 por par.', Confidence: 0.90, CreatedByAI: 'Motor Fiscal IA' },
    { ID: 'ai-003', Type: 'colecao_parada', Reason: 'Colecao inverno 2025 tem 47 pares sem movimento ha 60 dias. Sugiro liquidacao com 35% de desconto para liberar espaco.', Confidence: 0.86, CreatedByAI: 'Copiloto IA' },
  ],
}

const typeLabels: Record<string, string> = {
  missing_labor_cost: 'Custo de mao de obra ausente',
  ncm_correction: 'Correcao de NCM',
  onboard_field: 'Importacao de NF-e',
  validade_proxima: 'Validade Proxima',
  estoque_minimo: 'Estoque Minimo',
  agenda_vazia: 'Agenda com Horarios Vagos',
  produto_faltando: 'Produto em Falta',
  rota_otimizada: 'Rota Otimizada',
  veiculo_manutencao: 'Veiculo para Manutencao',
  materia_prima: 'Materia Prima Critica',
  pedido_atrasado: 'Pedido Atrasado',
  grade_incompleta: 'Grade Incompleta',
  colecao_parada: 'Colecao Parada',
}

export default function AIApprovalsPage() {
  const [suggestions, setSuggestions] = useState<any[]>([])

  useEffect(() => {
    const nicho = getBusinessType()
    setSuggestions(SUGGESTIONS_BY_NICHO[nicho] || SUGGESTIONS_BY_NICHO.mechanic)
  }, [])

  const handleAction = (id: string) => {
    if (isDemoMode()) { promptLogin(); return }
    setSuggestions(prev => prev.filter(s => s.ID !== id))
  }

  return (
    <div style={{ padding: 24, fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
        <div>
          <h2 style={{ fontSize: 20, fontWeight: 800, color: '#1C1917', margin: '0 0 4px' }}>Aprovacoes IA</h2>
          <p style={{ fontSize: 13, color: '#64748b', margin: 0 }}>Todas as sugestoes da IA precisam de aprovacao humana</p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, background: '#F5F3FF', border: '1px solid #EDE9FE', borderRadius: 100, padding: '6px 14px' }}>
          <Brain size={14} color="#7C3AED" />
          <span style={{ fontSize: 12, fontWeight: 700, color: '#7C3AED' }}>{suggestions.length} pendente{suggestions.length !== 1 ? 's' : ''}</span>
        </div>
      </div>

      {suggestions.length === 0 ? (
        <div style={{ background: 'white', borderRadius: 16, border: '0.5px solid #e8e8e8', padding: 48, textAlign: 'center' }}>
          <CheckCircle size={40} color="#4ADE80" style={{ margin: '0 auto 12px', display: 'block' }} />
          <p style={{ fontSize: 13, color: '#64748b', margin: 0 }}>Nenhuma sugestao pendente!</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {suggestions.map(s => (
            <div key={s.ID} style={{ background: 'white', borderRadius: 16, border: '0.5px solid #e8e8e8', padding: 20 }}>
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16 }}>
                <div style={{ width: 40, height: 40, background: '#F5F3FF', borderRadius: 12, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  <Brain size={18} color="#7C3AED" />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                    <span style={{ fontSize: 13, fontWeight: 700, color: '#1C1917' }}>{typeLabels[s.Type] || s.Type}</span>
                    <span style={{ fontSize: 10, fontWeight: 700, background: '#FFFBEB', color: '#D97706', border: '1px solid #FDE68A', borderRadius: 100, padding: '2px 8px' }}>
                      {Math.round(s.Confidence * 100)}% confianca
                    </span>
                    <span style={{ fontSize: 11, color: '#94a3b8' }}>{s.CreatedByAI}</span>
                  </div>
                  <p style={{ fontSize: 13, color: '#475569', margin: '0 0 8px', lineHeight: 1.5 }}>{s.Reason}</p>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <AlertTriangle size={11} color="#94a3b8" />
                    <span style={{ fontSize: 11, color: '#94a3b8' }}>Requer sua aprovacao antes de aplicar</span>
                  </div>
                </div>
                <div style={{ display: 'flex', gap: 8, flexShrink: 0 }}>
                  <button onClick={() => handleAction(s.ID)} style={{ display: 'flex', alignItems: 'center', gap: 6, background: '#ECFDF5', color: '#059669', border: '1px solid #BBF7D0', borderRadius: 10, padding: '8px 14px', cursor: 'pointer', fontSize: 12, fontWeight: 600 }}>
                    <ThumbsUp size={13} /> Aprovar
                  </button>
                  <button onClick={() => handleAction(s.ID)} style={{ display: 'flex', alignItems: 'center', gap: 6, background: '#FFF1F2', color: '#E11D48', border: '1px solid #FECDD3', borderRadius: 10, padding: '8px 14px', cursor: 'pointer', fontSize: 12, fontWeight: 600 }}>
                    <ThumbsDown size={13} /> Rejeitar
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}