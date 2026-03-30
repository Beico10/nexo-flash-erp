'use client'
import { useState, useEffect } from 'react'
import { isDemoMode, promptLogin, getBusinessType } from '@/lib/demo'
import { Plus, Search, AlertTriangle, Package, TrendingDown, DollarSign, X, ArrowDown, ArrowUp, RotateCcw } from 'lucide-react'

interface Product {
  id: string; code: string; barcode: string; name: string
  category: string; unit: string; quantity: number
  min_quantity: number; cost_price: number; sale_price: number
  total_value: number; location: string; is_active: boolean
  is_low_stock: boolean; is_out_of_stock: boolean
  extra: Record<string, any>
}

interface Summary {
  total_products: number; total_value: number
  low_stock_count: number; out_of_stock_count: number
}

interface NichoConfig {
  name: string; icon: string; unit_default: string
  categories: string[]; extra_fields: string[]
  show_expiry: boolean; disabled?: boolean
}

const DEMO_BY_NICHO: Record<string, { products: Product[]; summary: Summary; config: NichoConfig }> = {
  mechanic: {
    config: { name: '🔧 Estoque da Mecânica', icon: '🔧', unit_default: 'un', categories: ['Peça', 'Óleo/Fluido', 'Filtro', 'Pneu', 'Elétrica'], extra_fields: ['brand', 'application'], show_expiry: false },
    summary: { total_products: 8, total_value: 4872.5, low_stock_count: 3, out_of_stock_count: 1 },
    products: [
      { id: '1', code: 'PE-001', barcode: '7891001', name: 'Pastilha de Freio Dianteira', category: 'Peça', unit: 'jg', quantity: 4, min_quantity: 5, cost_price: 48.0, sale_price: 95.0, total_value: 192.0, location: 'Prateleira A1', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Fremax', application: 'Freios' } },
      { id: '2', code: 'OF-001', barcode: '7891002', name: 'Óleo Motor 5W30 Sintético 1L', category: 'Óleo/Fluido', unit: 'lt', quantity: 0, min_quantity: 10, cost_price: 28.5, sale_price: 48.0, total_value: 0, location: 'Prateleira B1', is_active: true, is_low_stock: false, is_out_of_stock: true, extra: { brand: 'Castrol', application: 'Motor' } },
      { id: '3', code: 'FI-001', barcode: '7891003', name: 'Filtro de Óleo Universal', category: 'Filtro', unit: 'un', quantity: 18, min_quantity: 10, cost_price: 12.0, sale_price: 22.0, total_value: 216.0, location: 'Prateleira A2', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Tecfil', application: 'Motor' } },
      { id: '4', code: 'PE-002', barcode: '7891004', name: 'Amortecedor Dianteiro', category: 'Peça', unit: 'un', quantity: 2, min_quantity: 4, cost_price: 180.0, sale_price: 320.0, total_value: 360.0, location: 'Estoque C1', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Monroe', application: 'Suspensão' } },
      { id: '5', code: 'PE-003', barcode: '7891005', name: 'Correia Dentada K1', category: 'Peça', unit: 'un', quantity: 6, min_quantity: 3, cost_price: 95.0, sale_price: 160.0, total_value: 570.0, location: 'Prateleira A3', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Gates', application: 'Motor' } },
      { id: '6', code: 'OF-002', barcode: '7891006', name: 'Fluido de Freio DOT4 500ml', category: 'Óleo/Fluido', unit: 'un', quantity: 12, min_quantity: 6, cost_price: 18.0, sale_price: 32.0, total_value: 216.0, location: 'Prateleira B2', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Bosch', application: 'Freios' } },
      { id: '7', code: 'EL-001', barcode: '7891007', name: 'Vela de Ignição NGK', category: 'Elétrica', unit: 'un', quantity: 3, min_quantity: 8, cost_price: 22.0, sale_price: 40.0, total_value: 66.0, location: 'Prateleira D1', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'NGK', application: 'Motor' } },
      { id: '8', code: 'FI-002', barcode: '7891008', name: 'Filtro de Ar Esportivo', category: 'Filtro', unit: 'un', quantity: 9, min_quantity: 5, cost_price: 42.0, sale_price: 75.0, total_value: 378.0, location: 'Prateleira A4', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'K&N', application: 'Motor' } },
    ],
  },
  bakery: {
    config: { name: '🥖 Estoque da Padaria', icon: '🥖', unit_default: 'kg', categories: ['Farinha/Grão', 'Laticínio', 'Açúcar/Doce', 'Embalagem', 'Gás/Insumo'], extra_fields: ['brand'], show_expiry: true },
    summary: { total_products: 8, total_value: 2846.0, low_stock_count: 3, out_of_stock_count: 1 },
    products: [
      { id: '1', code: 'FG-001', barcode: '7892001', name: 'Farinha de Trigo T55 25kg', category: 'Farinha/Grão', unit: 'sc', quantity: 3, min_quantity: 10, cost_price: 68.0, sale_price: 0, total_value: 204.0, location: 'Depósito Seco', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Anaconda' } },
      { id: '2', code: 'LA-001', barcode: '7892002', name: 'Manteiga sem Sal 500g', category: 'Laticínio', unit: 'un', quantity: 0, min_quantity: 12, cost_price: 14.5, sale_price: 0, total_value: 0, location: 'Câmara Fria', is_active: true, is_low_stock: false, is_out_of_stock: true, extra: { brand: 'Aviação' } },
      { id: '3', code: 'AC-001', barcode: '7892003', name: 'Açúcar Refinado 5kg', category: 'Açúcar/Doce', unit: 'pc', quantity: 8, min_quantity: 6, cost_price: 22.0, sale_price: 0, total_value: 176.0, location: 'Depósito Seco', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'União' } },
      { id: '4', code: 'LA-002', barcode: '7892004', name: 'Queijo Mussarela kg', category: 'Laticínio', unit: 'kg', quantity: 4, min_quantity: 8, cost_price: 32.0, sale_price: 0, total_value: 128.0, location: 'Câmara Fria', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Tirolez' } },
      { id: '5', code: 'FG-002', barcode: '7892005', name: 'Fermento Biológico Seco 500g', category: 'Farinha/Grão', unit: 'un', quantity: 2, min_quantity: 5, cost_price: 18.0, sale_price: 0, total_value: 36.0, location: 'Depósito Seco', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Fleischmann' } },
      { id: '6', code: 'EM-001', barcode: '7892006', name: 'Saco Plástico Pão Francês 100un', category: 'Embalagem', unit: 'pc', quantity: 15, min_quantity: 5, cost_price: 8.5, sale_price: 0, total_value: 127.5, location: 'Almoxarifado', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'BioEmbalagem' } },
      { id: '7', code: 'AC-002', barcode: '7892007', name: 'Chocolate em Pó 50% 200g', category: 'Açúcar/Doce', unit: 'un', quantity: 22, min_quantity: 10, cost_price: 9.8, sale_price: 0, total_value: 215.6, location: 'Depósito Seco', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Nestlé' } },
      { id: '8', code: 'GI-001', barcode: '7892008', name: 'Gás P13 Botijão', category: 'Gás/Insumo', unit: 'un', quantity: 3, min_quantity: 2, cost_price: 110.0, sale_price: 0, total_value: 330.0, location: 'Área Externa', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Ultragaz' } },
    ],
  },
  aesthetics: {
    config: { name: '💅 Estoque do Studio', icon: '💅', unit_default: 'un', categories: ['Capilar', 'Esmalte/Unha', 'Tintura', 'Descartável', 'Equipamento'], extra_fields: ['brand'], show_expiry: false },
    summary: { total_products: 8, total_value: 3218.0, low_stock_count: 2, out_of_stock_count: 1 },
    products: [
      { id: '1', code: 'CA-001', barcode: '7893001', name: 'Shampoo Profissional 1L', category: 'Capilar', unit: 'un', quantity: 3, min_quantity: 5, cost_price: 68.0, sale_price: 120.0, total_value: 204.0, location: 'Armário A', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: "L'Oreal Pro" } },
      { id: '2', code: 'TI-001', barcode: '7893002', name: 'Tinta Coloração 60g - Louro Dourado', category: 'Tintura', unit: 'un', quantity: 0, min_quantity: 8, cost_price: 22.0, sale_price: 38.0, total_value: 0, location: 'Armário B', is_active: true, is_low_stock: false, is_out_of_stock: true, extra: { brand: 'Igora Schwarzkopf' } },
      { id: '3', code: 'CA-002', barcode: '7893003', name: 'Máscara Capilar Hidratação 500g', category: 'Capilar', unit: 'un', quantity: 6, min_quantity: 4, cost_price: 45.0, sale_price: 80.0, total_value: 270.0, location: 'Armário A', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Wella' } },
      { id: '4', code: 'ES-001', barcode: '7893004', name: 'Esmalte Colorido 9ml', category: 'Esmalte/Unha', unit: 'un', quantity: 45, min_quantity: 20, cost_price: 4.5, sale_price: 9.0, total_value: 202.5, location: 'Expositor', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Colorama' } },
      { id: '5', code: 'CA-003', barcode: '7893005', name: 'Progressiva Orgânica 1L', category: 'Capilar', unit: 'un', quantity: 2, min_quantity: 3, cost_price: 120.0, sale_price: 200.0, total_value: 240.0, location: 'Armário B', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Inoar' } },
      { id: '6', code: 'DC-001', barcode: '7893006', name: 'Papel-toalha Salão 250 folhas', category: 'Descartável', unit: 'pc', quantity: 20, min_quantity: 8, cost_price: 12.0, sale_price: 0, total_value: 240.0, location: 'Almoxarifado', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Kimberly' } },
      { id: '7', code: 'TI-002', barcode: '7893007', name: 'Oxidante 30vol 900ml', category: 'Tintura', unit: 'un', quantity: 8, min_quantity: 6, cost_price: 18.0, sale_price: 32.0, total_value: 144.0, location: 'Armário B', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Wella' } },
      { id: '8', code: 'DC-002', barcode: '7893008', name: 'Luva Descartável Caixa 100un', category: 'Descartável', unit: 'cx', quantity: 7, min_quantity: 4, cost_price: 28.0, sale_price: 0, total_value: 196.0, location: 'Almoxarifado', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Supermax' } },
    ],
  },
  logistics: {
    config: { name: '🚛 Estoque de Insumos', icon: '🚛', unit_default: 'un', categories: ['Combustível', 'Pneu/Câmara', 'Peça Veicular', 'Embalagem', 'EPI'], extra_fields: ['brand', 'application'], show_expiry: false },
    summary: { total_products: 7, total_value: 18640.0, low_stock_count: 2, out_of_stock_count: 1 },
    products: [
      { id: '1', code: 'CB-001', barcode: '7894001', name: 'Diesel S10 (litros)', category: 'Combustível', unit: 'lt', quantity: 800, min_quantity: 500, cost_price: 6.2, sale_price: 0, total_value: 4960.0, location: 'Tanque Interno', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Petrobras', application: 'Frota' } },
      { id: '2', code: 'PN-001', barcode: '7894002', name: 'Pneu 215/75R17.5', category: 'Pneu/Câmara', unit: 'un', quantity: 2, min_quantity: 6, cost_price: 680.0, sale_price: 0, total_value: 1360.0, location: 'Borracharia', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Bridgestone', application: 'Eixo Dianteiro' } },
      { id: '3', code: 'PV-001', barcode: '7894003', name: 'Filtro Combustível Truck', category: 'Peça Veicular', unit: 'un', quantity: 0, min_quantity: 4, cost_price: 85.0, sale_price: 0, total_value: 0, location: 'Almoxarifado', is_active: true, is_low_stock: false, is_out_of_stock: true, extra: { brand: 'Mann Filter', application: 'Motor' } },
      { id: '4', code: 'EM-001', barcode: '7894004', name: 'Pallet PBR 1200x1000', category: 'Embalagem', unit: 'un', quantity: 42, min_quantity: 20, cost_price: 38.0, sale_price: 0, total_value: 1596.0, location: 'Pátio A', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'ABRAS', application: 'Carga Geral' } },
      { id: '5', code: 'CB-002', barcode: '7894005', name: 'Arla 32 - 20L', category: 'Combustível', unit: 'bb', quantity: 8, min_quantity: 4, cost_price: 95.0, sale_price: 0, total_value: 760.0, location: 'Tanque Interno', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'YPF', application: 'Frota Euro 5' } },
      { id: '6', code: 'PV-002', barcode: '7894006', name: 'Óleo Câmbio 80W90 1L', category: 'Peça Veicular', unit: 'lt', quantity: 3, min_quantity: 8, cost_price: 42.0, sale_price: 0, total_value: 126.0, location: 'Almoxarifado', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Shell', application: 'Câmbio' } },
      { id: '7', code: 'EP-001', barcode: '7894007', name: 'Colete Refletivo Motorista', category: 'EPI', unit: 'un', quantity: 12, min_quantity: 6, cost_price: 35.0, sale_price: 0, total_value: 420.0, location: 'Almoxarifado', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Plastcor', application: 'Segurança' } },
    ],
  },
  industry: {
    config: { name: '🏭 Estoque Industrial', icon: '🏭', unit_default: 'kg', categories: ['Materia Prima', 'Produto Acabado', 'Insumo', 'Embalagem'], extra_fields: ['brand', 'application'], show_expiry: false },
    summary: { total_products: 8, total_value: 9524.7, low_stock_count: 4, out_of_stock_count: 1 },
    products: [
      { id: '1', code: 'MP-001', barcode: '7891234560001', name: 'Resina EP-40', category: 'Materia Prima', unit: 'kg', quantity: 45, min_quantity: 200, cost_price: 32.5, sale_price: 0, total_value: 1462.5, location: 'Dep. A / Prateleira 1', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Quimica Brasil', application: 'Moldagem' } },
      { id: '2', code: 'MP-002', barcode: '7891234560002', name: 'PVC Rigido', category: 'Materia Prima', unit: 'kg', quantity: 0, min_quantity: 150, cost_price: 18.2, sale_price: 0, total_value: 0, location: 'Dep. A / Prateleira 2', is_active: true, is_low_stock: false, is_out_of_stock: true, extra: { brand: 'Plasticos Sul', application: 'Extrussao' } },
      { id: '3', code: 'MP-003', barcode: '7891234560003', name: 'Tinta Epoxi Cinza', category: 'Materia Prima', unit: 'lt', quantity: 80, min_quantity: 50, cost_price: 42.0, sale_price: 0, total_value: 3360.0, location: 'Dep. B / Armario 3', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Akzo Nobel', application: 'Acabamento' } },
      { id: '4', code: 'PA-001', barcode: '7891234560004', name: 'Conexao PVC DN50', category: 'Produto Acabado', unit: 'un', quantity: 320, min_quantity: 100, cost_price: 4.8, sale_price: 9.5, total_value: 1536.0, location: 'Dep. C / Box 1', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { application: 'Hidraulica' } },
      { id: '5', code: 'PA-002', barcode: '7891234560005', name: 'Tubo PVC DN100 6m', category: 'Produto Acabado', unit: 'un', quantity: 18, min_quantity: 30, cost_price: 28.4, sale_price: 52.0, total_value: 511.2, location: 'Dep. C / Box 2', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { application: 'Hidraulica' } },
      { id: '6', code: 'IN-001', barcode: '7891234560006', name: 'Parafuso Inox M8x40', category: 'Insumo', unit: 'cx', quantity: 12, min_quantity: 5, cost_price: 35.0, sale_price: 0, total_value: 420.0, location: 'Dep. B / Gaveta 1', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Ciser' } },
      { id: '7', code: 'MP-004', barcode: '7891234560007', name: 'Borracha EPDM', category: 'Materia Prima', unit: 'kg', quantity: 6, min_quantity: 20, cost_price: 68.0, sale_price: 0, total_value: 408.0, location: 'Dep. A / Prateleira 4', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Taborin', application: 'Vedacao' } },
      { id: '8', code: 'PA-003', barcode: '7891234560008', name: 'Tampa Industrial TI-200', category: 'Produto Acabado', unit: 'un', quantity: 145, min_quantity: 50, cost_price: 12.6, sale_price: 24.0, total_value: 1827.0, location: 'Dep. C / Box 3', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { application: 'Industrial' } },
    ],
  },
  shoes: {
    config: { name: '👟 Estoque de Calçados', icon: '👟', unit_default: 'par', categories: ['Calçado Masculino', 'Calçado Feminino', 'Infantil', 'Acessório', 'Matéria-Prima'], extra_fields: ['brand', 'application'], show_expiry: false },
    summary: { total_products: 8, total_value: 21480.0, low_stock_count: 3, out_of_stock_count: 1 },
    products: [
      { id: '1', code: 'CF-001', barcode: '7895001', name: 'Sandália Comfort Feminina nº35', category: 'Calçado Feminino', unit: 'par', quantity: 0, min_quantity: 6, cost_price: 55.0, sale_price: 119.9, total_value: 0, location: 'Prateleira F1', is_active: true, is_low_stock: false, is_out_of_stock: true, extra: { brand: 'Comfort Step', application: 'Uso Diário' } },
      { id: '2', code: 'CM-001', barcode: '7895002', name: 'Tênis Runner Pro Masculino nº42', category: 'Calçado Masculino', unit: 'par', quantity: 3, min_quantity: 8, cost_price: 98.0, sale_price: 219.9, total_value: 294.0, location: 'Prateleira M2', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'SpeedX', application: 'Corrida' } },
      { id: '3', code: 'CF-002', barcode: '7895003', name: 'Sandália Comfort Feminina nº36', category: 'Calçado Feminino', unit: 'par', quantity: 14, min_quantity: 6, cost_price: 55.0, sale_price: 119.9, total_value: 770.0, location: 'Prateleira F1', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Comfort Step', application: 'Uso Diário' } },
      { id: '4', code: 'CM-002', barcode: '7895004', name: 'Bota Social Couro nº41', category: 'Calçado Masculino', unit: 'par', quantity: 5, min_quantity: 4, cost_price: 145.0, sale_price: 329.9, total_value: 725.0, location: 'Prateleira M3', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Nobre Couro', application: 'Social' } },
      { id: '5', code: 'CI-001', barcode: '7895005', name: 'Tênis Infantil nº28', category: 'Infantil', unit: 'par', quantity: 2, min_quantity: 6, cost_price: 42.0, sale_price: 89.9, total_value: 84.0, location: 'Prateleira I1', is_active: true, is_low_stock: true, is_out_of_stock: false, extra: { brand: 'Mini Steps', application: 'Escolar' } },
      { id: '6', code: 'AC-001', barcode: '7895006', name: 'Palmilha Conforto Gel Un.', category: 'Acessório', unit: 'par', quantity: 30, min_quantity: 10, cost_price: 8.5, sale_price: 19.9, total_value: 255.0, location: 'Expositor', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Ortopé', application: 'Conforto' } },
      { id: '7', code: 'MP-001', barcode: '7895007', name: 'Couro Bovino Hidrofugado m²', category: 'Matéria-Prima', unit: 'm2', quantity: 85, min_quantity: 30, cost_price: 28.0, sale_price: 0, total_value: 2380.0, location: 'Depósito Couro', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Curtume Gaúcho', application: 'Cabedal' } },
      { id: '8', code: 'MP-002', barcode: '7895008', name: 'Solado Borracha Antiderrapante', category: 'Matéria-Prima', unit: 'par', quantity: 120, min_quantity: 50, cost_price: 12.0, sale_price: 0, total_value: 1440.0, location: 'Depósito MP', is_active: true, is_low_stock: false, is_out_of_stock: false, extra: { brand: 'Vulcabrás', application: 'Solado' } },
    ],
  },
}

const MOVEMENT_LABEL: Record<string, string> = {
  in: '↑ Entrada', out: '↓ Saída', adjust: '⚖️ Ajuste', loss: '🗑️ Perda'
}

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

export default function InventoryPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [summary, setSummary] = useState<Summary | null>(null)
  const [config, setConfig] = useState<NichoConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [filterCategory, setFilterCategory] = useState('')
  const [filterLow, setFilterLow] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [movModal, setMovModal] = useState<{product: Product, type: string} | null>(null)
  const [movQty, setMovQty] = useState('')
  const [movCost, setMovCost] = useState('')
  const [movNotes, setMovNotes] = useState('')
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  const [form, setForm] = useState({
    code: '', name: '', category: '', unit: '',
    quantity: '0', min_quantity: '0', cost_price: '0', sale_price: '0',
    location: '', ncm: '', description: ''
  })

  useEffect(() => { fetchAll() }, [search, filterCategory, filterLow])

  const fetchAll = async () => {
    if (isDemoMode()) {
      const nicho = getBusinessType()
      const demoData = DEMO_BY_NICHO[nicho] || DEMO_BY_NICHO.industry
      setConfig(demoData.config)
      let filtered = demoData.products
      if (search) filtered = filtered.filter(p => p.name.toLowerCase().includes(search.toLowerCase()) || p.code.toLowerCase().includes(search.toLowerCase()))
      if (filterCategory) filtered = filtered.filter(p => p.category === filterCategory)
      if (filterLow) filtered = filtered.filter(p => p.is_low_stock || p.is_out_of_stock)
      setProducts(filtered)
      setSummary(demoData.summary)
      setLoading(false)
      return
    }
    setLoading(true)
    const h = { Authorization: `Bearer ${token}` }
    try {
      const qs = new URLSearchParams()
      if (search) qs.set('search', search)
      if (filterCategory) qs.set('category', filterCategory)
      if (filterLow) qs.set('low_stock', 'true')

      const [cfgRes, listRes, sumRes] = await Promise.all([
        fetch('/api/v1/inventory/config', { headers: h }),
        fetch(`/api/v1/inventory/products?${qs}`, { headers: h }),
        fetch('/api/v1/inventory/products/summary', { headers: h }),
      ])
      if (cfgRes.ok) setConfig(await cfgRes.json())
      if (listRes.ok) setProducts((await listRes.json()).products || [])
      if (sumRes.ok) setSummary(await sumRes.json())
    } finally { setLoading(false) }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (isDemoMode()) { promptLogin(); return }
    const res = await fetch('/api/v1/inventory/products', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        ...form,
        quantity: parseFloat(form.quantity),
        min_quantity: parseFloat(form.min_quantity),
        cost_price: parseFloat(form.cost_price),
        sale_price: parseFloat(form.sale_price),
      }),
    })
    if (res.ok) { setShowForm(false); fetchAll() }
  }

  const handleMovement = async () => {
    if (!movModal) return
    if (isDemoMode()) { promptLogin(); return }
    const endpoint = movModal.type === 'adjust'
      ? `/api/v1/inventory/products/${movModal.product.id}/adjust`
      : movModal.type === 'in'
        ? `/api/v1/inventory/products/${movModal.product.id}/add`
        : `/api/v1/inventory/products/${movModal.product.id}/remove`

    const body = movModal.type === 'adjust'
      ? { new_quantity: parseFloat(movQty), notes: movNotes }
      : { quantity: parseFloat(movQty), unit_cost: parseFloat(movCost || '0'), notes: movNotes }

    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    })
    if (res.ok) { setMovModal(null); setMovQty(''); setMovCost(''); setMovNotes(''); fetchAll() }
  }

  if (config?.disabled) {
    return (
      <div style={{ padding: 40, textAlign: 'center', color: '#9E9E9E' }}>
        <Package size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
        <p style={{ fontSize: 16, fontWeight: 600 }}>Estoque não disponível para Logística</p>
      </div>
    )
  }

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header camaleão */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>
            {config?.icon} {config?.name || 'Estoque'}
          </h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>
            Custo Médio Ponderado automático • Alertas de mínimo via WhatsApp
          </p>
        </div>
        <button onClick={() => setShowForm(true)} style={{
          display: 'flex', alignItems: 'center', gap: 8,
          background: 'linear-gradient(135deg, #4A148C, #7B1FA2)',
          color: 'white', padding: '10px 20px', borderRadius: 10,
          border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 14,
        }}>
          <Plus size={16} /> Novo Produto
        </button>
      </div>

      {/* Cards resumo */}
      {summary && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14, marginBottom: 24 }}>
          {[
            { label: 'Produtos', value: summary.total_products, sub: 'cadastrados', color: '#1565C0', bg: '#E3F2FD', icon: <Package size={18} /> },
            { label: 'Valor Total', value: fmt(summary.total_value), sub: 'em estoque', color: '#2E7D32', bg: '#E8F5E9', icon: <DollarSign size={18} /> },
            { label: 'Estoque Baixo', value: summary.low_stock_count, sub: 'abaixo do mínimo', color: '#E65100', bg: '#FFF3E0', icon: <AlertTriangle size={18} /> },
            { label: 'Sem Estoque', value: summary.out_of_stock_count, sub: 'zerados', color: '#B71C1C', bg: '#FFEBEE', icon: <TrendingDown size={18} /> },
          ].map((c, i) => (
            <div key={i} style={{ background: c.bg, borderRadius: 14, padding: 16, border: `1.5px solid ${c.color}22` }}>
              <div className="flex items-center justify-between mb-2">
                <span style={{ fontSize: 11, fontWeight: 700, color: c.color, textTransform: 'uppercase' }}>{c.label}</span>
                <span style={{ color: c.color }}>{c.icon}</span>
              </div>
              <div style={{ fontSize: 22, fontWeight: 700, color: c.color }}>{c.value}</div>
              <div style={{ fontSize: 11, color: '#757575', marginTop: 3 }}>{c.sub}</div>
            </div>
          ))}
        </div>
      )}

      {/* Filtros */}
      <div style={{ display: 'flex', gap: 10, marginBottom: 16, flexWrap: 'wrap' }}>
        <div style={{ position: 'relative', flex: 1, minWidth: 200 }}>
          <Search size={14} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: '#9E9E9E' }} />
          <input value={search} onChange={e => setSearch(e.target.value)}
            placeholder="Buscar produto..." style={{
              width: '100%', padding: '9px 12px 9px 34px',
              border: '1.5px solid #E0E4F0', borderRadius: 10, fontSize: 13, outline: 'none',
            }} />
        </div>
        <select value={filterCategory} onChange={e => setFilterCategory(e.target.value)} style={{
          padding: '9px 14px', border: '1.5px solid #E0E4F0', borderRadius: 10, fontSize: 13, outline: 'none',
        }}>
          <option value="">Todas categorias</option>
          {config?.categories.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
        <button onClick={() => setFilterLow(v => !v)} style={{
          padding: '9px 16px', borderRadius: 10, border: '1.5px solid',
          borderColor: filterLow ? '#E65100' : '#E0E4F0',
          background: filterLow ? '#FFF3E0' : 'white',
          color: filterLow ? '#E65100' : '#757575',
          fontSize: 12, fontWeight: 600, cursor: 'pointer',
        }}>
          ⚠️ Estoque Baixo
        </button>
      </div>

      {/* Lista */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#757575' }}>Carregando...</div>
      ) : products.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
          <Package size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
          <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhum produto encontrado</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {products.map(p => (
            <div key={p.id} style={{
              background: 'white', borderRadius: 12, padding: '14px 16px',
              border: `1.5px solid ${p.is_out_of_stock ? '#EF9A9A' : p.is_low_stock ? '#FFCC02' : '#E0E4F0'}`,
              display: 'flex', alignItems: 'center', gap: 14,
            }}>
              {/* Indicador */}
              <div style={{
                width: 8, height: 8, borderRadius: '50%', flexShrink: 0,
                background: p.is_out_of_stock ? '#B71C1C' : p.is_low_stock ? '#E65100' : '#2E7D32',
              }} />

              {/* Info */}
              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-2">
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#212121' }}>{p.name}</span>
                  {p.code && <span style={{ fontSize: 10, background: '#F5F7FF', color: '#5C6BC0', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>{p.code}</span>}
                  {p.is_out_of_stock && <span style={{ fontSize: 10, background: '#FFEBEE', color: '#B71C1C', padding: '1px 8px', borderRadius: 100, fontWeight: 700 }}>SEM ESTOQUE</span>}
                  {p.is_low_stock && !p.is_out_of_stock && <span style={{ fontSize: 10, background: '#FFF3E0', color: '#E65100', padding: '1px 8px', borderRadius: 100, fontWeight: 700 }}>⚠️ BAIXO</span>}
                </div>
                <div style={{ display: 'flex', gap: 12, marginTop: 4, fontSize: 12, color: '#757575' }}>
                  <span>{p.category}</span>
                  {p.location && <span>📍 {p.location}</span>}
                  {p.extra?.application && <span>🚗 {p.extra.application}</span>}
                  {p.extra?.brand && <span>{p.extra.brand}</span>}
                </div>
              </div>

              {/* Quantidade */}
              <div style={{ textAlign: 'center', minWidth: 80 }}>
                <div style={{ fontSize: 20, fontWeight: 700, color: p.is_out_of_stock ? '#B71C1C' : '#212121' }}>
                  {p.quantity}
                </div>
                <div style={{ fontSize: 11, color: '#9E9E9E' }}>{p.unit} {p.min_quantity > 0 && `(mín: ${p.min_quantity})`}</div>
              </div>

              {/* CMP */}
              <div style={{ textAlign: 'right', minWidth: 100 }}>
                <div style={{ fontSize: 13, fontWeight: 700, color: '#212121' }}>{fmt(p.cost_price)}</div>
                <div style={{ fontSize: 11, color: '#9E9E9E' }}>CMP unitário</div>
                <div style={{ fontSize: 11, color: '#2E7D32', fontWeight: 600 }}>{fmt(p.total_value)} total</div>
              </div>

              {/* Ações */}
              <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                <button onClick={() => setMovModal({ product: p, type: 'in' })} title="Entrada" style={{ background: '#E8F5E9', border: '1px solid #A5D6A7', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', fontSize: 16 }}>↑</button>
                <button onClick={() => setMovModal({ product: p, type: 'out' })} title="Saída" style={{ background: '#FFEBEE', border: '1px solid #EF9A9A', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', fontSize: 16 }}>↓</button>
                <button onClick={() => setMovModal({ product: p, type: 'adjust' })} title="Ajuste" style={{ background: '#F5F7FF', border: '1px solid #C5CAE9', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', fontSize: 16 }}>⚖️</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal movimentação */}
      {movModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 380 }}>
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 16, fontWeight: 700 }}>{MOVEMENT_LABEL[movModal.type]}</h3>
              <button onClick={() => setMovModal(null)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
            </div>
            <p style={{ fontSize: 13, color: '#757575', marginBottom: 16 }}>{movModal.product.name}</p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div>
                <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>
                  {movModal.type === 'adjust' ? 'Nova Quantidade' : 'Quantidade'}
                </label>
                <input type="number" value={movQty} onChange={e => setMovQty(e.target.value)}
                  placeholder="0" style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
              </div>
              {movModal.type === 'in' && (
                <div>
                  <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Custo Unitário (R$)</label>
                  <input type="number" value={movCost} onChange={e => setMovCost(e.target.value)}
                    placeholder="0.00" style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                  {movCost && movQty && (
                    <p style={{ fontSize: 11, color: '#2E7D32', marginTop: 4 }}>
                      Novo CMP será recalculado automaticamente
                    </p>
                  )}
                </div>
              )}
              <div>
                <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Observação</label>
                <input type="text" value={movNotes} onChange={e => setMovNotes(e.target.value)}
                  placeholder="Ex: NF-e 4521, OS #1042..." style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
              </div>
              <button onClick={handleMovement} style={{
                width: '100%', padding: 13, borderRadius: 10,
                background: movModal.type === 'in' ? '#2E7D32' : movModal.type === 'out' ? '#B71C1C' : '#4A148C',
                color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: 'pointer', marginTop: 4,
              }}>
                Confirmar {MOVEMENT_LABEL[movModal.type]}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
