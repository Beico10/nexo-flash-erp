'use client'

import { useState, useRef, useEffect } from 'react'
import { Camera, Upload, X, Check, AlertCircle, Receipt, Loader2 } from 'lucide-react'

interface ScanResult {
  success: boolean
  message: string
  expense?: {
    id: string
    supplier_name: string
    total_amount: number
    category: string
    ibs_credit: number
    cbs_credit: number
    issue_date: string
    items_count: number
  }
}

export default function QRCodeScanner() {
  const [scanning, setScanning] = useState(false)
  const [processing, setProcessing] = useState(false)
  const [result, setResult] = useState<ScanResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [cameraPermission, setCameraPermission] = useState<'prompt' | 'granted' | 'denied'>('prompt')
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const streamRef = useRef<MediaStream | null>(null)

  // Verificar permissão da câmera
  useEffect(() => {
    navigator.permissions?.query({ name: 'camera' as PermissionName }).then((status) => {
      setCameraPermission(status.state as 'prompt' | 'granted' | 'denied')
      status.onchange = () => setCameraPermission(status.state as 'prompt' | 'granted' | 'denied')
    })
  }, [])

  // Iniciar câmera
  const startCamera = async () => {
    try {
      setError(null)
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' } // Câmera traseira
      })
      
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        streamRef.current = stream
        setScanning(true)
        setCameraPermission('granted')
        
        // Iniciar detecção de QR Code
        detectQRCode()
      }
    } catch (err) {
      console.error('Erro ao acessar câmera:', err)
      setCameraPermission('denied')
      setError('Não foi possível acessar a câmera. Verifique as permissões.')
    }
  }

  // Parar câmera
  const stopCamera = () => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop())
      streamRef.current = null
    }
    setScanning(false)
  }

  // Detectar QR Code usando BarcodeDetector API (Chrome) ou fallback
  const detectQRCode = async () => {
    if (!videoRef.current || !canvasRef.current) return

    const video = videoRef.current
    const canvas = canvasRef.current
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Usar BarcodeDetector se disponível
    if ('BarcodeDetector' in window) {
      const detector = new (window as any).BarcodeDetector({ formats: ['qr_code'] })
      
      const scan = async () => {
        if (!scanning || !video.videoWidth) {
          requestAnimationFrame(scan)
          return
        }

        try {
          const barcodes = await detector.detect(video)
          if (barcodes.length > 0) {
            const qrContent = barcodes[0].rawValue
            stopCamera()
            await processQRCode(qrContent)
            return
          }
        } catch (err) {
          // Continuar tentando
        }
        
        requestAnimationFrame(scan)
      }
      
      scan()
    } else {
      // Fallback: capturar frame e enviar para backend processar
      // Ou usar biblioteca como jsQR
      setError('Seu navegador não suporta leitura de QR Code nativa. Use o upload de arquivo.')
    }
  }

  // Processar QR Code
  const processQRCode = async (qrContent: string) => {
    setProcessing(true)
    setError(null)

    try {
      const token = localStorage.getItem('token')
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/expenses/scan`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ qr_content: qrContent }),
      })

      const data = await res.json()

      if (res.ok) {
        setResult({ success: true, message: data.message, expense: data.expense })
      } else {
        setResult({ success: false, message: data.error || 'Erro ao processar QR Code' })
      }
    } catch (err) {
      setResult({ success: false, message: 'Erro de conexão. Tente novamente.' })
    } finally {
      setProcessing(false)
    }
  }

  // Upload de XML
  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setProcessing(true)
    setError(null)

    try {
      const formData = new FormData()
      formData.append('xml', file)

      const token = localStorage.getItem('token')
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/expenses/upload-xml`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      })

      const data = await res.json()

      if (res.ok) {
        setResult({ success: true, message: data.message, expense: data.expense })
      } else {
        setResult({ success: false, message: data.error || 'Erro ao importar XML' })
      }
    } catch (err) {
      setResult({ success: false, message: 'Erro de conexão. Tente novamente.' })
    } finally {
      setProcessing(false)
    }
  }

  // Resetar para nova leitura
  const resetScanner = () => {
    setResult(null)
    setError(null)
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value)
  }

  return (
    <div className="max-w-lg mx-auto p-4">
      <div className="bg-white rounded-2xl shadow-lg overflow-hidden">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-blue-700 p-6 text-white">
          <div className="flex items-center gap-3">
            <Receipt className="w-8 h-8" />
            <div>
              <h1 className="text-xl font-bold">Registrar Despesa</h1>
              <p className="text-blue-100 text-sm">Escaneie o QR Code da nota fiscal</p>
            </div>
          </div>
        </div>

        <div className="p-6">
          {/* Resultado de sucesso */}
          {result?.success && result.expense && (
            <div className="mb-6">
              <div className="bg-green-50 border border-green-200 rounded-xl p-4 mb-4">
                <div className="flex items-center gap-2 text-green-700 mb-3">
                  <Check className="w-5 h-5" />
                  <span className="font-medium">Despesa registrada!</span>
                </div>
                
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Fornecedor</span>
                    <span className="font-medium text-gray-900">{result.expense.supplier_name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Valor Total</span>
                    <span className="font-bold text-gray-900">{formatCurrency(result.expense.total_amount)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Categoria</span>
                    <span className="text-gray-900 capitalize">{result.expense.category}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Itens</span>
                    <span className="text-gray-900">{result.expense.items_count} produtos</span>
                  </div>
                  
                  {(result.expense.ibs_credit > 0 || result.expense.cbs_credit > 0) && (
                    <div className="pt-2 mt-2 border-t border-green-200">
                      <div className="flex justify-between text-green-700">
                        <span>Crédito de Imposto</span>
                        <span className="font-bold">
                          {formatCurrency(result.expense.ibs_credit + result.expense.cbs_credit)}
                        </span>
                      </div>
                      <p className="text-xs text-green-600 mt-1">
                        IBS: {formatCurrency(result.expense.ibs_credit)} + CBS: {formatCurrency(result.expense.cbs_credit)}
                      </p>
                    </div>
                  )}
                </div>
              </div>

              <button
                onClick={resetScanner}
                className="w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 transition-colors"
              >
                Escanear outra nota
              </button>
            </div>
          )}

          {/* Resultado de erro */}
          {result && !result.success && (
            <div className="mb-6">
              <div className="bg-red-50 border border-red-200 rounded-xl p-4 mb-4">
                <div className="flex items-center gap-2 text-red-700">
                  <AlertCircle className="w-5 h-5" />
                  <span className="font-medium">{result.message}</span>
                </div>
              </div>

              <button
                onClick={resetScanner}
                className="w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 transition-colors"
              >
                Tentar novamente
              </button>
            </div>
          )}

          {/* Scanner */}
          {!result && (
            <>
              {/* Área da câmera */}
              <div className="relative aspect-square bg-gray-100 rounded-xl overflow-hidden mb-4">
                {scanning ? (
                  <>
                    <video
                      ref={videoRef}
                      autoPlay
                      playsInline
                      muted
                      className="w-full h-full object-cover"
                    />
                    <canvas ref={canvasRef} className="hidden" />
                    
                    {/* Overlay com guia */}
                    <div className="absolute inset-0 flex items-center justify-center">
                      <div className="w-48 h-48 border-2 border-white rounded-lg shadow-lg" />
                    </div>
                    
                    {/* Botão fechar */}
                    <button
                      onClick={stopCamera}
                      className="absolute top-3 right-3 p-2 bg-black/50 rounded-full text-white hover:bg-black/70"
                    >
                      <X className="w-5 h-5" />
                    </button>
                    
                    <p className="absolute bottom-4 left-0 right-0 text-center text-white text-sm bg-black/50 py-2">
                      Posicione o QR Code dentro do quadrado
                    </p>
                  </>
                ) : processing ? (
                  <div className="w-full h-full flex flex-col items-center justify-center">
                    <Loader2 className="w-12 h-12 text-blue-600 animate-spin mb-4" />
                    <p className="text-gray-600">Processando nota fiscal...</p>
                  </div>
                ) : (
                  <div className="w-full h-full flex flex-col items-center justify-center p-6 text-center">
                    <Camera className="w-16 h-16 text-gray-400 mb-4" />
                    <p className="text-gray-600 mb-2">
                      Escaneie o QR Code da NFC-e ou NF-e
                    </p>
                    <p className="text-gray-400 text-sm">
                      A despesa será registrada automaticamente
                    </p>
                  </div>
                )}
              </div>

              {/* Erro de permissão */}
              {error && (
                <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-4 mb-4">
                  <div className="flex items-center gap-2 text-yellow-700">
                    <AlertCircle className="w-5 h-5" />
                    <span className="text-sm">{error}</span>
                  </div>
                </div>
              )}

              {/* Botões de ação */}
              <div className="space-y-3">
                {!scanning && (
                  <button
                    onClick={startCamera}
                    disabled={processing}
                    className="w-full py-4 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 transition-colors flex items-center justify-center gap-2 disabled:opacity-50"
                  >
                    <Camera className="w-5 h-5" />
                    Abrir Câmera
                  </button>
                )}

                <div className="relative">
                  <input
                    type="file"
                    accept=".xml"
                    onChange={handleFileUpload}
                    disabled={processing || scanning}
                    className="absolute inset-0 w-full h-full opacity-0 cursor-pointer disabled:cursor-not-allowed"
                  />
                  <button
                    disabled={processing || scanning}
                    className="w-full py-4 bg-gray-100 text-gray-700 rounded-xl font-medium hover:bg-gray-200 transition-colors flex items-center justify-center gap-2 disabled:opacity-50"
                  >
                    <Upload className="w-5 h-5" />
                    Importar XML da Nota
                  </button>
                </div>
              </div>

              {/* Info */}
              <div className="mt-6 p-4 bg-blue-50 rounded-xl">
                <h3 className="font-medium text-blue-900 mb-2">Como funciona?</h3>
                <ul className="text-sm text-blue-700 space-y-1">
                  <li>• Aponte a câmera para o QR Code da nota</li>
                  <li>• O sistema lê e registra a despesa automaticamente</li>
                  <li>• Calcula o crédito de imposto (IBS/CBS)</li>
                  <li>• Gera relatório para abater no IR</li>
                </ul>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
