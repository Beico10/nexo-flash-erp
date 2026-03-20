#!/usr/bin/env python3
"""
Micro-serviço de IA para o Nexo One ERP.
Expõe um endpoint HTTP que o backend Go pode chamar.
"""
import os
import asyncio
from fastapi import FastAPI, Request, Response
from pydantic import BaseModel
from typing import Optional
import json

app = FastAPI(title="Nexo One AI Service")

# Cache de sessões
sessions = {}

class ChatRequest(BaseModel):
    question: str
    session_id: str = "default"
    system_prompt: Optional[str] = None
    context: Optional[str] = None

@app.post("/chat")
async def chat(req: ChatRequest):
    from emergentintegrations.llm.chat import LlmChat, UserMessage
    
    api_key = os.environ.get("EMERGENT_LLM_KEY", "")
    if not api_key:
        return {"error": "EMERGENT_LLM_KEY nao configurada"}
    
    system_msg = req.system_prompt or """Voce e o Co-Piloto do Nexo One ERP, um assistente inteligente para negocios brasileiros.
Voce ajuda com: gestao financeira, fiscal (IBS/CBS 2026), estoque, agendamentos, logistica, producao.
Responda sempre em portugues brasileiro, de forma direta e acionavel.
Quando receber dados do sistema, analise e de sugestoes proativas.
Use emojis moderadamente. Seja conciso."""
    
    try:
        chat_instance = LlmChat(
            api_key=api_key,
            session_id=req.session_id,
            system_message=system_msg
        ).with_model("gemini", "gemini-3-flash-preview")
        
        prompt = req.question
        if req.context:
            prompt = f"Dados do sistema:\n{req.context}\n\nPergunta: {req.question}"
        
        user_message = UserMessage(text=prompt)
        response = await chat_instance.send_message(user_message)
        
        return {"suggestion": response, "model": "gemini-3-flash-preview", "session_id": req.session_id}
    except Exception as e:
        return {"error": str(e)}

@app.post("/clear")
async def clear_session(req: dict):
    session_id = req.get("session_id", "default")
    if session_id in sessions:
        del sessions[session_id]
    return {"message": "sessao limpa"}

@app.get("/health")
async def health():
    return {"status": "ok", "service": "nexo-ai"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=8003)
