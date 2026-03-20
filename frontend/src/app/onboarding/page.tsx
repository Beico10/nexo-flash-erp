'use client'

import { useState, useEffect } from 'react'
import { CheckCircle2, Circle, SkipForward, Clock, Gift, ChevronRight, Sparkles } from 'lucide-react'

interface OnboardingStep {
  ID: string
  StepCode: string
  StepOrder: number
  Title: string
  Description: string
  Icon: string
  IsRequired: boolean
  IsSkippable: boolean
  EstimatedTime: number
  RewardText: string
}

interface OnboardingProgress {
  TenantID: string
  CurrentStep: string
  TotalSteps: number
  CompletedSteps: string[]
  Skipped: boolean
}

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

export default function OnboardingPage() {
  const [steps, setSteps] = useState<OnboardingStep[]>([])
  const [progress, setProgress] = useState<OnboardingProgress | null>(null)
  const [percent, setPercent] = useState(0)
  const [rewardDays, setRewardDays] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = getToken()
    if (!token) { window.location.href = '/login'; return }
    fetch('/api/v1/onboarding/progress', { headers: { Authorization: `Bearer ${token}` } })
      .then(r => { if (r.status === 401) { window.location.href = '/login'; return null }; return r.json() })
      .then(d => {
        if (d) {
          setSteps(d.steps || [])
          setProgress(d.progress)
          setPercent(d.percent || 0)
          setRewardDays(d.reward_days || 0)
          setLoading(false)
        }
      })
      .catch(() => setLoading(false))
  }, [])

  const completeStep = async (stepCode: string, skipped: boolean) => {
    const token = getToken()
    const res = await fetch('/api/v1/onboarding/complete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ step_code: stepCode, skipped }),
    })
    if (res.ok) {
      setProgress(prev => {
        if (!prev) return prev
        return { ...prev, CompletedSteps: [...prev.CompletedSteps, stepCode] }
      })
      setPercent(prev => Math.min(100, prev + Math.round(100 / steps.length)))
    }
  }

  const skipOnboarding = async () => {
    const token = getToken()
    await fetch('/api/v1/onboarding/skip', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    })
    window.location.href = '/dashboard'
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600" />
      </div>
    )
  }

  const completedSet = new Set(progress?.CompletedSteps || [])

  return (
    <div className="max-w-3xl mx-auto py-8 px-4">
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900" data-testid="onboarding-title">Primeiros passos</h1>
            <p className="text-sm text-gray-500 mt-1">Complete as etapas para aproveitar ao maximo o Nexo One</p>
          </div>
          <button onClick={skipOnboarding} data-testid="skip-onboarding-btn" className="text-sm text-gray-400 hover:text-gray-600 flex items-center gap-1">
            <SkipForward size={14} /> Pular
          </button>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5 mb-6">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium text-gray-700">Progresso</span>
            <span className="text-sm font-bold text-blue-600">{percent}%</span>
          </div>
          <div className="h-2.5 bg-gray-100 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full transition-all duration-500"
              style={{ width: `${percent}%` }}
              data-testid="onboarding-progress-bar"
            />
          </div>
          {rewardDays > 0 && (
            <p className="text-xs text-green-600 mt-2 flex items-center gap-1">
              <Gift size={12} /> +{rewardDays} dias de trial ganhos!
            </p>
          )}
        </div>
      </div>

      <div className="space-y-3">
        {steps.map((step, idx) => {
          const completed = completedSet.has(step.StepCode)
          const isCurrent = !completed && (idx === 0 || completedSet.has(steps[idx - 1]?.StepCode))

          return (
            <div
              key={step.ID}
              data-testid={`onboarding-step-${step.StepCode}`}
              className={`bg-white rounded-xl border transition-all ${
                completed ? 'border-green-200 bg-green-50/30' :
                isCurrent ? 'border-blue-200 shadow-sm ring-1 ring-blue-100' :
                'border-gray-200 opacity-60'
              } p-5`}
            >
              <div className="flex items-start gap-4">
                <div className="pt-0.5">
                  {completed ? (
                    <CheckCircle2 size={22} className="text-green-500" />
                  ) : (
                    <Circle size={22} className={isCurrent ? 'text-blue-400' : 'text-gray-300'} />
                  )}
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <h3 className={`font-semibold ${completed ? 'text-green-700 line-through' : 'text-gray-900'}`}>
                      {step.Title}
                    </h3>
                    {step.IsRequired && !completed && (
                      <span className="text-[10px] font-bold text-red-500 bg-red-50 px-1.5 py-0.5 rounded">OBRIGATORIO</span>
                    )}
                  </div>
                  <p className="text-sm text-gray-500 mt-1">{step.Description}</p>
                  <div className="flex items-center gap-4 mt-3">
                    <span className="text-xs text-gray-400 flex items-center gap-1">
                      <Clock size={11} /> ~{Math.round(step.EstimatedTime / 60)} min
                    </span>
                    {step.RewardText && (
                      <span className="text-xs text-amber-600 flex items-center gap-1">
                        <Sparkles size={11} /> {step.RewardText}
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex gap-2">
                  {isCurrent && !completed && (
                    <>
                      <button
                        data-testid={`complete-step-${step.StepCode}`}
                        onClick={() => completeStep(step.StepCode, false)}
                        className="px-3 py-1.5 text-xs font-semibold bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-1"
                      >
                        Completar <ChevronRight size={12} />
                      </button>
                      {step.IsSkippable && (
                        <button
                          onClick={() => completeStep(step.StepCode, true)}
                          className="px-3 py-1.5 text-xs text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-50"
                        >
                          Pular
                        </button>
                      )}
                    </>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {percent === 100 && (
        <div className="mt-8 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl p-6 text-white text-center">
          <h2 className="text-xl font-bold mb-2">Parabens! Onboarding completo!</h2>
          <p className="text-blue-100 mb-4">Voce ja esta pronto para usar o Nexo One no dia a dia.</p>
          <button
            onClick={() => window.location.href = '/dashboard'}
            className="px-6 py-2.5 bg-white text-blue-600 font-semibold rounded-lg hover:bg-blue-50"
          >
            Ir para o Dashboard
          </button>
        </div>
      )}
    </div>
  )
}
