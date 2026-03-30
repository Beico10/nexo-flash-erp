// ============================================================
// DADOS DEMO — Gestão Para Todos
// ============================================================

export const DEMO_MODE_KEY = 'nexo_demo_mode'
export const BUSINESS_TYPE_KEY = 'nexo_business_type'

export function isDemoMode(): boolean {
  if (typeof window === 'undefined') return false
  return localStorage.getItem(DEMO_MODE_KEY) === 'true'
}

export function getBusinessType(): string {
  if (typeof window === 'undefined') return 'mechanic'
  return localStorage.getItem(BUSINESS_TYPE_KEY) || 'mechanic'
}
export const DEMO_MECHANIC_ORDERS = [
  { id: 'os-001', number: 'OS-2024-001', tenant_id: 'demo', vehicle_plate: 'BRA2E19', vehicle_km: 87340, vehicle_model: 'Honda Civic', vehicle_year: 2021, customer_id: 'Carlos Eduardo Ferreira', customer_phone: '5511998234567', status: 'await_approval', complaint: 'Barulho ao frear, pedal esponjoso', diagnosis: 'Pastilhas dianteiras desgastadas, disco com estrias. Recomendo troca completa.', created_at: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(), valor: 480.00, ibs_cbs: 38.40 },
  { id: 'os-002', number: 'OS-2024-002', tenant_id: 'demo', vehicle_plate: 'ABC1D23', vehicle_km: 45200, vehicle_model: 'Toyota Corolla', vehicle_year: 2022, customer_id: 'Ana Paula Rodrigues', customer_phone: '5511997654321', status: 'in_progress', complaint: 'Troca de óleo e revisão geral', diagnosis: 'Óleo 5W30 sintético, filtros trocados, verificado freios e suspensão.', created_at: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 60).toISOString(), valor: 320.00, ibs_cbs: 25.60 },
  { id: 'os-003', number: 'OS-2024-003', tenant_id: 'demo', vehicle_plate: 'DEF4G56', vehicle_km: 112000, vehicle_model: 'Volkswagen Gol', vehicle_year: 2018, customer_id: 'Roberto Almeida Santos', customer_phone: '5511996543210', status: 'open', complaint: 'Motor falhando em temperatura alta', diagnosis: '', created_at: new Date(Date.now() - 1000 * 60 * 20).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 20).toISOString(), valor: 0, ibs_cbs: 0 },
  { id: 'os-004', number: 'OS-2024-004', tenant_id: 'demo', vehicle_plate: 'GHI7J89', vehicle_km: 67800, vehicle_model: 'Chevrolet Onix', vehicle_year: 2020, customer_id: 'Fernanda Lima Costa', customer_phone: '5511995432109', status: 'done', complaint: 'Ar condicionado não resfria', diagnosis: 'Recarga de gás R134a, limpeza do evaporador.', created_at: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 60 * 3).toISOString(), valor: 280.00, ibs_cbs: 22.40 },
  { id: 'os-005', number: 'OS-2024-005', tenant_id: 'demo', vehicle_plate: 'JKL0M12', vehicle_km: 203000, vehicle_model: 'Ford Ka', vehicle_year: 2016, customer_id: 'Marcos Vinícius Pereira', customer_phone: '5511994321098', status: 'diagnosed', complaint: 'Suspensão barulhenta na frente', diagnosis: 'Amortecedores dianteiros danificados, buchas do braço gasto. Peças solicitadas.', created_at: new Date(Date.now() - 1000 * 60 * 60 * 8).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), valor: 850.00, ibs_cbs: 68.00 },
  { id: 'os-006', number: 'OS-2024-006', tenant_id: 'demo', vehicle_plate: 'NOP3Q45', vehicle_km: 31500, vehicle_model: 'Hyundai HB20', vehicle_year: 2023, customer_id: 'Juliana Carvalho Melo', customer_phone: '5511993210987', status: 'invoiced', complaint: 'Revisão dos 30.000 km', diagnosis: 'Revisão completa realizada conforme manual. Tudo OK.', created_at: new Date(Date.now() - 1000 * 60 * 60 * 48).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), valor: 620.00, ibs_cbs: 49.60 },
  { id: 'os-007', number: 'OS-2024-007', tenant_id: 'demo', vehicle_plate: 'RST6U78', vehicle_km: 156000, vehicle_model: 'Fiat Uno', vehicle_year: 2014, customer_id: 'Paulo Henrique Souza', customer_phone: '5511992109876', status: 'open', complaint: 'Vazamento de óleo pelo cárter', diagnosis: '', created_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(), updated_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(), valor: 0, ibs_cbs: 0 },
]
export function calcularImpostos(valor: number, regime: 'simples' | 'lucro_presumido' | 'lucro_real' = 'simples') {
  const aliquotas = {
    simples: { ibs: 0.06, cbs: 0.04, total: 0.10 },
    lucro_presumido: { ibs: 0.08, cbs: 0.08, total: 0.16 },
    lucro_real: { ibs: 0.10, cbs: 0.09, total: 0.19 },
  }
  const a = aliquotas[regime]
  return {
    valor_base: valor,
    ibs: valor * a.ibs,
    cbs: valor * a.cbs,
    total_imposto: valor * a.total,
    credito_presumido: valor * a.total * 0.4,
    valor_liquido: valor - (valor * a.total * 0.6),
    regime,
  }
}

export const COPILOTO_MESSAGES: Record<string, string[]> = {
  mechanic: [
    'Você tem 2 OS aguardando aprovação do cliente. Quer que eu coloque isso em destaque no seu painel?',
    'Notei que o Marcos Vinícius tem uma OS de suspensão há 8 horas sem atualização. Precisa de atenção?',
    'Essa semana sua quinta-feira foi o melhor dia — R$ 5.200 faturados. Quer ver o que mudou nesse dia?',
  ],
  bakery: [
    'Você tem 3 ingredientes em falta: Farinha, Fermento e Manteiga. Quer que eu monte uma lista de compras?',
    'O Sonho está zerado no estoque e tem um pedido do Buffet Sabor & Arte esperando. Precisa produzir hoje?',
    'Quarta-feira costuma ser seu melhor dia de produção. Quer planejar o que produzir amanhã?',
  ],
  aesthetics: [
    'Você tem 8 agendamentos hoje. A Patrícia tem a progressiva às 15h — são 3 horas de serviço. Está confirmado?',
    'Tem 5 clientes sem retorno há mais de 30 dias. Quer que eu monte uma mensagem de reativação no WhatsApp?',
    'A Mariana Oliveira marcou coloração completa — R$ 280. Lembrar de confirmar 24h antes?',
  ],
  logistics: [
    'A rota RT-SP-02 está atrasada desde as 07:00. O Anderson ainda está no pátio. Precisa acionar?',
    'Você tem 3 CT-es pendentes de emissão antes de sair as cargas. Quer que eu organize por prioridade?',
    'Hoje são 12 entregas previstas. Quinta-feira é seu melhor dia da semana historicamente.',
  ],
  industry: [
    'A Resina EP-40 está crítica — 45 litros, mínimo de 200. Dois pedidos dependem desse material.',
    'PVC Rígido está zerado e o pedido PED-2024-1044 está parado por isso. Qual fornecedor acionar?',
    'Os pedidos #1042 e #1043 saem hoje às 14:00. Produção confirmada?',
  ],
  shoes: [
    'A Sandália Comfort está em estoque crítico — 30 pares, sendo que o 35 já zerou. Quer repor?',
    'O Tênis Runner Pro nº42 está com só 3 pares. Foi o número mais vendido essa semana.',
    'Hoje você já fez 5 vendas — R$ 1.139,40. Falta R$ 2.310,60 para bater a meta diária.',
  ],
}