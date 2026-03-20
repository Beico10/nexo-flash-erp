import subprocess
import signal
import sys
import os
import httpx
import asyncio
from fastapi import FastAPI, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv

load_dotenv("/app/backend/.env")

app = FastAPI(title="Nexo One Proxy")

GO_BINARY = "/app/nexo-one"
GO_PORT = 8002
go_process = None

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.on_event("startup")
async def startup():
    global go_process
    env = os.environ.copy()
    env["PORT"] = str(GO_PORT)
    env["JWT_SECRET"] = os.environ.get("JWT_SECRET", "nexo-one-dev-secret-2026")
    go_process = subprocess.Popen(
        [GO_BINARY],
        env=env,
        stdout=sys.stdout,
        stderr=sys.stderr,
    )


@app.on_event("shutdown")
async def shutdown():
    global go_process
    if go_process:
        go_process.send_signal(signal.SIGTERM)
        go_process.wait(timeout=10)


# ── AI Co-Piloto Endpoint ─────────────────────────────────
@app.post("/api/v1/copilot/suggest")
async def copilot_suggest(request: Request):
    from emergentintegrations.llm.chat import LlmChat, UserMessage

    auth = request.headers.get("authorization", "")
    if not auth.startswith("Bearer "):
        return Response(content='{"error":"nao autorizado"}', status_code=401, media_type="application/json")

    body = await request.json()
    context_data = body.get("context", "")
    question = body.get("question", "")
    session_id = body.get("session_id", "copilot-default")

    api_key = os.environ.get("EMERGENT_LLM_KEY", "")
    if not api_key:
        return Response(content='{"error":"LLM key nao configurada"}', status_code=500, media_type="application/json")

    system_msg = """Voce e o Co-Piloto do Nexo One ERP, um assistente inteligente para negocios brasileiros.
Voce ajuda com: gestao financeira, fiscal (IBS/CBS 2026), estoque, agendamentos, logistica, producao.
Responda sempre em portugues brasileiro, de forma direta e acionavel.
Quando receber dados do sistema, analise e de sugestoes proativas.
Use emojis moderadamente. Seja conciso."""

    try:
        chat = LlmChat(
            api_key=api_key,
            session_id=session_id,
            system_message=system_msg
        ).with_model("gemini", "gemini-3-flash-preview")

        prompt = question
        if context_data:
            prompt = f"Dados do sistema:\n{context_data}\n\nPergunta: {question}"

        user_message = UserMessage(text=prompt)
        response = await chat.send_message(user_message)

        return Response(
            content=__import__("json").dumps({"suggestion": response, "model": "gemini-3-flash"}, ensure_ascii=False),
            status_code=200,
            media_type="application/json",
        )
    except Exception as e:
        return Response(
            content=__import__("json").dumps({"error": str(e)}, ensure_ascii=False),
            status_code=500,
            media_type="application/json",
        )


@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"])
async def proxy(request: Request, path: str):
    target = f"http://127.0.0.1:{GO_PORT}/{path}"
    if request.url.query:
        target += f"?{request.url.query}"

    body = await request.body()
    headers = dict(request.headers)
    headers.pop("host", None)

    async with httpx.AsyncClient(timeout=30.0) as client:
        try:
            resp = await client.request(
                method=request.method,
                url=target,
                headers=headers,
                content=body,
            )
        except httpx.ConnectError:
            return Response(
                content='{"error":"backend Go nao iniciado"}',
                status_code=502,
                media_type="application/json",
            )

    resp_headers = dict(resp.headers)
    resp_headers.pop("transfer-encoding", None)
    resp_headers.pop("content-encoding", None)
    resp_headers.pop("content-length", None)

    return Response(
        content=resp.content,
        status_code=resp.status_code,
        headers=resp_headers,
        media_type=resp.headers.get("content-type"),
    )
