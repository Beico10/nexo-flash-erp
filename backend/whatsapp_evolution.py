# =============================================================================
# NEXO ONE — WhatsApp Evolution API
# Verificação real do número de telefone para trial
#
# Evolution API v2.3.7
# Instância: erpnexoone
# =============================================================================

import httpx
import json
import re
import os
from fastapi import APIRouter, Request, Response, HTTPException
from pydantic import BaseModel
from typing import Optional

whatsapp_router = APIRouter(prefix="/api/v1/whatsapp", tags=["whatsapp"])

# ── CONFIGURAÇÃO ──────────────────────────────────────────────────────────────
EVOLUTION_URL      = os.getenv("EVOLUTION_URL", "https://socialgrizzlybear-evolution.cloudfy.live")
EVOLUTION_KEY      = os.getenv("EVOLUTION_KEY", "0iGsRJBF2EgcducrjqkMz3IgQyosoKmR")
EVOLUTION_INSTANCE = os.getenv("EVOLUTION_INSTANCE", "erpnexoone")
NEXO_NUMBER        = os.getenv("NEXO_WHATSAPP", "5511947726031")

HEADERS = {
    "apikey": EVOLUTION_KEY,
    "Content-Type": "application/json",
}

# ── MODELOS ───────────────────────────────────────────────────────────────────

class SendCodeRequest(BaseModel):
    phone: str          # ex: "5511999887766"
    tenant_slug: str    # ex: "mecanica-joao"
    business_type: str  # ex: "mechanic"

class VerifyCodeRequest(BaseModel):
    phone: str
    code: str

# ── ENVIO DE CÓDIGO ───────────────────────────────────────────────────────────

def clean_phone(phone: str) -> str:
    """Remove caracteres não numéricos e garante DDI 55."""
    digits = re.sub(r'\D', '', phone)
    if not digits.startswith('55'):
        digits = '55' + digits
    return digits

def generate_code(phone: str) -> str:
    """Gera código de 6 dígitos baseado no telefone + timestamp."""
    import hashlib, time
    seed = f"{phone}{int(time.time() // 300)}"  # Válido por 5 min
    h = hashlib.sha256(seed.encode()).hexdigest()
    return str(int(h[:8], 16))[-6:].zfill(6)

async def send_whatsapp_message(phone: str, message: str) -> dict:
    """Envia mensagem via Evolution API."""
    url = f"{EVOLUTION_URL}/message/sendText/{EVOLUTION_INSTANCE}"
    body = {
        "number": phone,
        "text": message,
        "delay": 0,
    }
    async with httpx.AsyncClient(timeout=15.0) as client:
        resp = await client.post(url, headers=HEADERS, json=body)
        return resp.json()

# ── ENDPOINTS ─────────────────────────────────────────────────────────────────

@whatsapp_router.post("/send-verification")
async def send_verification(req: SendCodeRequest):
    """
    Envia código de verificação para o WhatsApp do usuário.
    O usuário responde com o código para ativar o trial.
    """
    phone = clean_phone(req.phone)

    if len(phone) < 12 or len(phone) > 13:
        raise HTTPException(400, "Número de telefone inválido")

    code = generate_code(phone)

    # Mensagem de verificação
    nicho_emoji = {
        "mechanic":   "🔧",
        "bakery":     "🍞",
        "industry":   "🏭",
        "logistics":  "🚛",
        "aesthetics": "💇",
        "shoes":      "👟",
    }.get(req.business_type, "💼")

    message = (
        f"{nicho_emoji} *Nexo One ERP* — Verificação\n\n"
        f"Olá! Seu código de acesso ao trial gratuito:\n\n"
        f"*{code}*\n\n"
        f"⏱ Válido por 5 minutos.\n"
        f"Digite este código no sistema para começar.\n\n"
        f"_Nexo One — Gestão inteligente para o seu negócio_"
    )

    try:
        result = await send_whatsapp_message(phone, message)
        return {
            "status": "ok",
            "message": "Código enviado via WhatsApp",
            "phone": f"****{phone[-4:]}",
            # Em dev retorna o código — em produção remover
            "debug_code": code if os.getenv("APP_ENV", "development") == "development" else None,
        }
    except Exception as e:
        raise HTTPException(500, f"Erro ao enviar WhatsApp: {str(e)}")


@whatsapp_router.post("/verify-code")
async def verify_code(req: VerifyCodeRequest):
    """
    Valida o código digitado pelo usuário.
    Retorna sucesso se o código for válido (últimos 5 min).
    """
    phone = clean_phone(req.phone)
    expected = generate_code(phone)

    # Verificar código atual
    if req.code == expected:
        return {"status": "ok", "verified": True, "phone": phone}

    # Verificar código do período anterior (janela de 10 min)
    import hashlib, time
    seed_prev = f"{phone}{int(time.time() // 300) - 1}"
    h_prev = hashlib.sha256(seed_prev.encode()).hexdigest()
    code_prev = str(int(h_prev[:8], 16))[-6:].zfill(6)

    if req.code == code_prev:
        return {"status": "ok", "verified": True, "phone": phone}

    raise HTTPException(400, "Código inválido ou expirado. Solicite um novo.")


@whatsapp_router.post("/webhook")
async def evolution_webhook(request: Request):
    """
    Webhook da Evolution API — recebe mensagens do WhatsApp.
    Configurar na Evolution: POST /api/v1/whatsapp/webhook
    Extrai automaticamente o código se o usuário responder com 6 dígitos.
    """
    try:
        body = await request.json()
    except Exception:
        return Response(status_code=200)

    event = body.get("event", "")

    # Processar apenas mensagens recebidas
    if event != "messages.upsert":
        return Response(status_code=200)

    data = body.get("data", {})
    message_obj = data.get("message", {})

    # Extrair texto da mensagem
    text = (
        message_obj.get("conversation") or
        message_obj.get("extendedTextMessage", {}).get("text") or
        ""
    ).strip()

    # Extrair número do remetente
    from_number = data.get("key", {}).get("remoteJid", "").replace("@s.whatsapp.net", "").replace("@c.us", "")

    if not from_number or not text:
        return Response(status_code=200)

    # Verificar se é um código de 6 dígitos
    code_match = re.search(r'\b(\d{6})\b', text)
    if not code_match:
        return Response(status_code=200)

    code = code_match.group(1)
    expected = generate_code(from_number)

    if code == expected:
        # Código válido — notificar o backend Go via HTTP interno
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                await client.post(
                    "http://127.0.0.1:8002/api/auth/verify/confirm-webhook",
                    json={"phone": from_number, "code": code},
                )
        except Exception:
            pass

        # Responder ao usuário
        await send_whatsapp_message(
            from_number,
            "✅ *Número verificado com sucesso!*\n\n"
            "Seu trial gratuito do Nexo One ERP está ativo.\n"
            "Acesse o sistema e comece a usar agora! 🚀"
        )

    return Response(status_code=200)


@whatsapp_router.post("/send-approval")
async def send_approval_link(
    phone: str,
    customer_name: str,
    os_number: str,
    approval_url: str,
    total_value: float = 0.0
):
    """
    Envia link de aprovação de orçamento (módulo Mecânica).
    """
    phone = clean_phone(phone)

    message = (
        f"🔧 *Orçamento Pronto — {os_number}*\n\n"
        f"Olá, {customer_name}!\n\n"
        f"Seu orçamento está pronto para aprovação.\n"
        f"💰 Valor total: *R$ {total_value:.2f}*\n\n"
        f"👆 Clique para ver e aprovar:\n"
        f"{approval_url}\n\n"
        f"_Qualquer dúvida, estamos à disposição!_"
    )

    try:
        await send_whatsapp_message(phone, message)
        return {"status": "ok", "message": "Link de aprovação enviado"}
    except Exception as e:
        raise HTTPException(500, f"Erro: {str(e)}")


@whatsapp_router.get("/status")
async def check_status():
    """Verifica se a instância Evolution está conectada."""
    url = f"{EVOLUTION_URL}/instance/fetchInstances"
    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.get(url, headers=HEADERS)
            data = resp.json()

            # Buscar instância erpnexoone
            instances = data if isinstance(data, list) else []
            for inst in instances:
                if inst.get("instance", {}).get("instanceName") == EVOLUTION_INSTANCE:
                    state = inst.get("instance", {}).get("state", "unknown")
                    return {
                        "status": "ok",
                        "instance": EVOLUTION_INSTANCE,
                        "connected": state == "open",
                        "state": state,
                    }

            return {"status": "error", "message": "Instância não encontrada"}
    except Exception as e:
        return {"status": "error", "message": str(e)}
