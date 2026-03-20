'use client'
import { useState } from 'react'
import { Truck, MapPin, Calculator, Loader2 } from 'lucide-react'

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

export default function LogisticsPage() {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)
  const [form, setForm] = useState({ distance_km: '500', weight_kg: '15000', toll_cost: '250', fuel_cost_per_km: '2.10', driver_cost_per_km: '0.80' })

  async function calculate() {
    setLoading(true)
    try {
      const token = getToken()
      const res = await fetch('/api/v1/logistics/freight/calculate', {
        method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ vehicle_type: 'truck', distance_km: parseFloat(form.distance_km), weight_kg: parseFloat(form.weight_kg), toll_cost: parseFloat(form.toll_cost), fuel_cost_per_km: parseFloat(form.fuel_cost_per_km), driver_cost_per_km: parseFloat(form.driver_cost_per_km) }),
      })
      const data = await res.json()
      if (res.ok) setResult(data)
    } catch {} finally { setLoading(false) }
  }

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

  return (
    <div className="space-y-5 animate-fade-in" data-testid="logistics-page">
      <div className="grid lg:grid-cols-2 gap-6">
        <div className="card p-5">
          <h3 className="font-semibold text-slate-800 text-sm mb-4 flex items-center gap-2"><Calculator size={15} className="text-nexo-500" /> Calculadora de Frete</h3>
          <div className="space-y-3">
            {Object.entries({ distance_km: 'Distancia (km)', weight_kg: 'Peso (kg)', toll_cost: 'Pedagio (R$)', fuel_cost_per_km: 'Combustivel/km (R$)', driver_cost_per_km: 'Motorista/km (R$)' }).map(([key, label]) => (
              <div key={key}><label className="label">{label}</label><input className="input" type="number" value={(form as any)[key]} onChange={e => setForm({ ...form, [key]: e.target.value })} /></div>
            ))}
            <button onClick={calculate} disabled={loading} className="btn-primary w-full mt-2">{loading ? <Loader2 size={15} className="animate-spin" /> : <Truck size={15} />} Calcular Frete</button>
          </div>
        </div>
        <div className="card p-5">
          <h3 className="font-semibold text-slate-800 text-sm mb-4 flex items-center gap-2"><MapPin size={15} className="text-nexo-500" /> Resultado DRE Viagem</h3>
          {result ? (
            <div className="space-y-3">
              {Object.entries(result).filter(([k]) => typeof result[k] === 'number').map(([k, v]) => (
                <div key={k} className="flex justify-between items-center py-2 border-b border-slate-100">
                  <span className="text-sm text-slate-600">{k.replace(/_/g, ' ')}</span>
                  <span className="font-semibold text-slate-800">{fmt(v as number)}</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-12 text-slate-300"><Truck size={48} className="mx-auto mb-3" /><p className="text-sm">Insira os dados e calcule</p></div>
          )}
        </div>
      </div>
    </div>
  )
}
