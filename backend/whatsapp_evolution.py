# =============================================================================
# NEXO ONE — Integração WhatsApp via Evolution API
# 
# Configuração:
#   EVOLUTION_URL: URL base da Evolution API (ex: https://evolution.exemplo.com)
#   EVOLUTION_KEY: API Key da Evolution
#   EVOLUTION_INSTANCE: Nome da instância (ex: erpnexoone)
#   NEXO_WHATSAPP: Número do WhatsApp do Nexo (ex: 5511999999999)
#
# Webhook na Evolution:
#   URL: https://SEU_DOMINIO/api/v1/whatsapp/webhook
#   Evento: messages.upsert
# =============================================================================

import os
import httpx
import json
import re
from datetime import datetime
from typing import Optional, Dict, Any, List
from fastapi import APIRouter, HTTPException, Request, BackgroundTasks
from pydantic import BaseModel

whatsapp_router = APIRouter(prefix="/api/v1/whatsapp", tags=["whatsapp"])

# ── CONFIGURAÇÃO ──────────────────────────────────────────────────────────────

def load_env():
    """Carrega variáveis do .env se não estiverem no ambiente."""
    env_file = "/app/backend/.env"
    if os.path.exists(env_file):
        with open(env_file) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    if key not in os.environ or not os.environ[key]:
                        os.environ[key] = value

load_env()

EVOLUTION_URL = os.environ.get("EVOLUTION_URL", "")
EVOLUTION_KEY = os.environ.get("EVOLUTION_KEY", "")
EVOLUTION_INSTANCE = os.environ.get("EVOLUTION_INSTANCE", "")
NEXO_WHATSAPP = os.environ.get("NEXO_WHATSAPP", "")

TIMEOUT = 30.0

# ── MODELOS ───────────────────────────────────────────────────────────────────

class SendMessageRequest(BaseModel):
    to: str  # Número destino (5511999999999)
    message: str
    
class SendMediaRequest(BaseModel):
    to: str
    media_url: str
    caption: str = ""
    media_type: str = "image"  # image, video, audio, document

class VerificationRequest(BaseModel):
    phone: str
    code: str = ""  # Se vazio, gera novo código

# ── CLIENTE EVOLUTION API ─────────────────────────────────────────────────────

async def evolution_request(
    method: str,
    endpoint: str,
    data: Optional[Dict] = None
) -> Dict[str, Any]:
    """Faz requisição para Evolution API."""
    if not EVOLUTION_URL or not EVOLUTION_KEY:
        raise HTTPException(500, "Evolution API não configurada (EVOLUTION_URL/EVOLUTION_KEY)")
    
    url = f"{EVOLUTION_URL.rstrip('/')}/{endpoint}"
    headers = {
        "apikey": EVOLUTION_KEY,
        "Content-Type": "application/json"
    }
    
    async with httpx.AsyncClient(timeout=TIMEOUT) as client:
        try:
            if method == "GET":
                resp = await client.get(url, headers=headers)
            elif method == "POST":
                resp = await client.post(url, headers=headers, json=data or {})
            elif method == "DELETE":
                resp = await client.delete(url, headers=headers)
            else:
                raise HTTPException(400, f"Método não suportado: {method}")
            
            if resp.status_code >= 400:
                return {"error": True, "status": resp.status_code, "detail": resp.text}
            
            return resp.json() if resp.text else {"success": True}
            
        except httpx.ConnectError:
            raise HTTPException(502, "Não foi possível conectar à Evolution API")
        except Exception as e:
            raise HTTPException(500, f"Erro Evolution API: {str(e)}")

def format_phone(phone: str) -> str:
    """Formata número para padrão WhatsApp (5511999999999)."""
    digits = re.sub(r'\D', '', phone)
    if len(digits) == 11:  # DDD + 9 dígitos
        digits = "55" + digits
    elif len(digits) == 10:  # DDD + 8 dígitos (fixo)
        digits = "55" + digits
    return digits

# ── ARMAZENAMENTO CÓDIGOS VERIFICAÇÃO (em memória - usar Redis em prod) ───────

verification_codes: Dict[str, Dict] = {}

import random
import string

def generate_code() -> str:
    """Gera código de 6 dígitos."""
    return ''.join(random.choices(string.digits, k=6))

# ── ENDPOINTS ─────────────────────────────────────────────────────────────────

@whatsapp_router.get("/status")
async def get_status():
    """Verifica status da conexão WhatsApp."""
    if not EVOLUTION_INSTANCE:
        return {"connected": False, "error": "EVOLUTION_INSTANCE não configurada"}
    
    result = await evolution_request("GET", f"instance/connectionState/{EVOLUTION_INSTANCE}")
    
    # Evolution v2 retorna {instance: {state: "open"}}
    state = result.get("instance", {}).get("state") or result.get("state", "unknown")
    
    return {
        "instance": EVOLUTION_INSTANCE,
        "connected": state == "open",
        "state": state,
        "detail": result
    }


@whatsapp_router.post("/send")
async def send_message(req: SendMessageRequest):
    """Envia mensagem de texto via WhatsApp."""
    if not EVOLUTION_INSTANCE:
        raise HTTPException(500, "EVOLUTION_INSTANCE não configurada")
    
    phone = format_phone(req.to)
    
    data = {
        "number": phone,
        "text": req.message
    }
    
    result = await evolution_request(
        "POST",
        f"message/sendText/{EVOLUTION_INSTANCE}",
        data
    )
    
    return {
        "success": not result.get("error"),
        "to": phone,
        "message_id": result.get("key", {}).get("id"),
        "detail": result
    }


@whatsapp_router.post("/send-media")
async def send_media(req: SendMediaRequest):
    """Envia mídia (imagem, vídeo, documento) via WhatsApp."""
    if not EVOLUTION_INSTANCE:
        raise HTTPException(500, "EVOLUTION_INSTANCE não configurada")
    
    phone = format_phone(req.to)
    
    data = {
        "number": phone,
        "mediatype": req.media_type,
        "media": req.media_url,
        "caption": req.caption
    }
    
    result = await evolution_request(
        "POST",
        f"message/sendMedia/{EVOLUTION_INSTANCE}",
        data
    )
    
    return {
        "success": not result.get("error"),
        "to": phone,
        "media_type": req.media_type,
        "detail": result
    }


@whatsapp_router.post("/verify/send")
async def send_verification(req: VerificationRequest):
    """Envia código de verificação via WhatsApp."""
    phone = format_phone(req.phone)
    code = generate_code()
    
    # Armazena código (expira em 10 minutos)
    verification_codes[phone] = {
        "code": code,
        "created_at": datetime.now().isoformat(),
        "attempts": 0
    }
    
    message = f"🔐 *Nexo One - Código de Verificação*\n\nSeu código é: *{code}*\n\nVálido por 10 minutos.\nNão compartilhe este código."
    
    # Envia via WhatsApp
    send_req = SendMessageRequest(to=phone, message=message)
    result = await send_message(send_req)
    
    return {
        "success": result.get("success"),
        "phone": phone,
        "message": "Código enviado via WhatsApp",
        "expires_in_minutes": 10
    }


@whatsapp_router.post("/verify/check")
async def check_verification(req: VerificationRequest):
    """Verifica código recebido."""
    phone = format_phone(req.phone)
    
    stored = verification_codes.get(phone)
    if not stored:
        return {"valid": False, "error": "Código não encontrado ou expirado"}
    
    # Verifica tentativas
    if stored["attempts"] >= 5:
        del verification_codes[phone]
        return {"valid": False, "error": "Muitas tentativas. Solicite novo código."}
    
    # Verifica expiração (10 minutos)
    created = datetime.fromisoformat(stored["created_at"])
    if (datetime.now() - created).seconds > 600:
        del verification_codes[phone]
        return {"valid": False, "error": "Código expirado"}
    
    # Verifica código
    stored["attempts"] += 1
    
    if stored["code"] == req.code:
        del verification_codes[phone]
        return {"valid": True, "phone": phone, "message": "Verificação bem-sucedida"}
    
    return {
        "valid": False,
        "error": "Código incorreto",
        "attempts_remaining": 5 - stored["attempts"]
    }


@whatsapp_router.post("/webhook")
async def webhook_receiver(request: Request, background_tasks: BackgroundTasks):
    """
    Recebe webhooks da Evolution API.
    Configure na Evolution: messages.upsert
    """
    try:
        payload = await request.json()
    except:
        return {"received": True}
    
    event = payload.get("event")
    instance = payload.get("instance")
    
    # Log do evento
    print(f"[WhatsApp Webhook] {event} from {instance}")
    
    # Processa mensagens recebidas
    if event == "messages.upsert":
        data = payload.get("data", {})
        message = data.get("message", {})
        key = data.get("key", {})
        
        # Ignora mensagens enviadas por nós
        if key.get("fromMe"):
            return {"received": True, "ignored": "fromMe"}
        
        sender = key.get("remoteJid", "").replace("@s.whatsapp.net", "")
        text = message.get("conversation") or message.get("extendedTextMessage", {}).get("text", "")
        
        if text:
            print(f"[WhatsApp] Mensagem de {sender}: {text}")
            # Aqui você pode adicionar lógica de resposta automática
            # background_tasks.add_task(process_incoming_message, sender, text)
    
    return {"received": True, "event": event}


@whatsapp_router.get("/config")
async def get_config():
    """Retorna configuração atual (sem expor chaves)."""
    return {
        "evolution_url": EVOLUTION_URL[:30] + "..." if EVOLUTION_URL else "não configurado",
        "evolution_key": "***" + EVOLUTION_KEY[-4:] if len(EVOLUTION_KEY) > 4 else "não configurado",
        "evolution_instance": EVOLUTION_INSTANCE or "não configurado",
        "nexo_whatsapp": NEXO_WHATSAPP or "não configurado",
        "webhook_url": "/api/v1/whatsapp/webhook"
    }
