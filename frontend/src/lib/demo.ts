export function isDemoMode(): boolean {
  if (typeof window === 'undefined') return false
  const token = localStorage.getItem('nexo_token') || ''
  const demo = localStorage.getItem('nexo_demo_mode')
  return demo === 'true' || token === 'demo-token' || token.startsWith('demo')
}
export function getBusinessType(): string {
  if (typeof window === 'undefined') return 'mechanic'
  return localStorage.getItem('nexo_business_type') || 'mechanic'
}
export function promptLogin() {
  if (typeof window !== 'undefined' && (window as any).showLoginPrompt) {
    (window as any).showLoginPrompt()
  }
}
export const DEMO_BAKERY_PRODUCTS = [
  { ID: '1', Name: 'Pao Frances', SKU: 'PAO-001', SaleType: 'weight', UnitPrice: 12.9, NCMCode: '1905.90.90', IsBasketItem: true, CurrentStock: 45, MinStock: 20, ScaleCode: 'BAL-01' },
  { ID: '2', Name: 'Bolo de Chocolate', SKU: 'BOL-001', SaleType: 'unit', UnitPrice: 48.0, NCMCode: '1905.90.20', IsBasketItem: false, CurrentStock: 8, MinStock: 5, ScaleCode: 'BAL-02' },
  { ID: '3', Name: 'Croissant', SKU: 'CRO-001', SaleType: 'unit', UnitPrice: 6.5, NCMCode: '1905.90.90', IsBasketItem: false, CurrentStock: 3, MinStock: 10, ScaleCode: 'BAL-01' },
  { ID: '4', Name: 'Pao de Queijo', SKU: 'PDQ-001', SaleType: 'unit', UnitPrice: 4.5, NCMCode: '1905.90.90', IsBasketItem: true, CurrentStock: 60, MinStock: 30, ScaleCode: 'BAL-03' },
  { ID: '5', Name: 'Cuca de Banana', SKU: 'CUC-001', SaleType: 'unit', UnitPrice: 32.0, NCMCode: '1905.90.20', IsBasketItem: false, CurrentStock: 2, MinStock: 4, ScaleCode: 'BAL-02' },
]
export const DEMO_AESTHETICS_APPOINTMENTS = [
  { ID: '1', ProfessionalID: 'prof-1', CustomerName: 'Fernanda Costa', ServiceName: 'Corte + Escova', ServicePrice: 120, StartTime: new Date(new Date().setHours(9,0)).toISOString(), EndTime: new Date(new Date().setHours(10,30)).toISOString(), DurationMin: 90, Status: 'confirmed' },
  { ID: '2', ProfessionalID: 'prof-1', CustomerName: 'Patricia Mendes', ServiceName: 'Coloracao', ServicePrice: 220, StartTime: new Date(new Date().setHours(11,0)).toISOString(), EndTime: new Date(new Date().setHours(13,0)).toISOString(), DurationMin: 120, Status: 'scheduled' },
  { ID: '3', ProfessionalID: 'prof-2', CustomerName: 'Juliana Alves', ServiceName: 'Manicure + Pedicure', ServicePrice: 85, StartTime: new Date(new Date().setHours(9,30)).toISOString(), EndTime: new Date(new Date().setHours(11,0)).toISOString(), DurationMin: 90, Status: 'in_progress' },
  { ID: '4', ProfessionalID: 'prof-2', CustomerName: 'Camila Rocha', ServiceName: 'Hidratacao', ServicePrice: 150, StartTime: new Date(new Date().setHours(14,0)).toISOString(), EndTime: new Date(new Date().setHours(15,30)).toISOString(), DurationMin: 90, Status: 'scheduled' },
  { ID: '5', ProfessionalID: 'prof-3', CustomerName: 'Bruna Souza', ServiceName: 'Sobrancelha + Cilios', ServicePrice: 95, StartTime: new Date(new Date().setHours(10,0)).toISOString(), EndTime: new Date(new Date().setHours(11,0)).toISOString(), DurationMin: 60, Status: 'completed' },
  { ID: '6', ProfessionalID: 'prof-3', CustomerName: 'Larissa Lima', ServiceName: 'Progressiva', ServicePrice: 280, StartTime: new Date(new Date().setHours(13,0)).toISOString(), EndTime: new Date(new Date().setHours(16,0)).toISOString(), DurationMin: 180, Status: 'confirmed' },
]
export const DEMO_MECHANIC_ORDERS = [
  { id: '1', tenant_id: 'demo', number: 'OS-001', vehicle_plate: 'BRA2E19', vehicle_km: 45200, vehicle_model: 'Honda Civic', vehicle_year: 2021, customer_id: 'Joao Silva', customer_phone: '5511999990001', status: 'open', complaint: 'Troca de oleo e filtro', diagnosis: '', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: '2', tenant_id: 'demo', number: 'OS-002', vehicle_plate: 'CDE3F20', vehicle_km: 82000, vehicle_model: 'Toyota Corolla', vehicle_year: 2019, customer_id: 'Maria Santos', customer_phone: '5511999990002', status: 'await_approval', complaint: 'Barulho na suspensao', diagnosis: 'Amortecedor danificado', created_at: new Date(Date.now()-86400000).toISOString(), updated_at: new Date().toISOString() },
  { id: '3', tenant_id: 'demo', number: 'OS-003', vehicle_plate: 'FGH4G21', vehicle_km: 31000, vehicle_model: 'Chevrolet Onix', vehicle_year: 2022, customer_id: 'Carlos Pereira', customer_phone: '5511999990003', status: 'in_progress', complaint: 'Revisao 30.000 km', diagnosis: 'Troca de pastilhas e oleo', created_at: new Date(Date.now()-172800000).toISOString(), updated_at: new Date().toISOString() },
  { id: '4', tenant_id: 'demo', number: 'OS-004', vehicle_plate: 'IJK5H22', vehicle_km: 67500, vehicle_model: 'Volkswagen Gol', vehicle_year: 2018, customer_id: 'Ana Oliveira', customer_phone: '5511999990004', status: 'done', complaint: 'Troca de correia dentada', diagnosis: 'Correia e tensor substituidos', created_at: new Date(Date.now()-259200000).toISOString(), updated_at: new Date().toISOString() },
  { id: '5', tenant_id: 'demo', number: 'OS-005', vehicle_plate: 'LMN6I23', vehicle_km: 12000, vehicle_model: 'Fiat Pulse', vehicle_year: 2023, customer_id: 'Roberto Lima', customer_phone: '5511999990005', status: 'open', complaint: 'Luz do painel acesa', diagnosis: '', created_at: new Date(Date.now()-3600000).toISOString(), updated_at: new Date().toISOString() },
]
