# Servidor wrapper - Inicia binário Go Nexo One + Serviço de IA
import subprocess
import signal
import sys
import os
from fastapi import FastAPI, Request, Response
from contextlib import asynccontextmanager
import httpx
import asyncio
from router_module import router_api

GO_BINARY = "/app/nexo-one"
go_process = None
ai_process = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global go_process, ai_process
    
    # Carrega variáveis de ambiente
    env = os.environ.copy()
    env_file = "/app/backend/.env"
    if os.path.exists(env_file):
        with open(env_file) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    env[key] = value
    
    # Inicia o serviço de IA (Python)
    print("Iniciando servico de IA na porta 8003...")
    ai_process = subprocess.Popen(
        [sys.executable, "-m", "uvicorn", "ai_service:app", "--host", "127.0.0.1", "--port", "8003"],
        env=env,
        cwd="/app/backend",
        stdout=sys.stdout,
        stderr=sys.stderr,
    )
    
    # Aguarda o serviço de IA iniciar
    await asyncio.sleep(2)
    
    # Verifica se o binário Go existe
    if not os.path.exists(GO_BINARY):
        print(f"ERRO: Binario Go nao encontrado em {GO_BINARY}")
        print("Execute: cd /app && go build -o /app/nexo-one ./cmd/api/")
        yield
        return
    
    env["PORT"] = "8002"
    
    # Inicia o binário Go
    print(f"Iniciando Nexo One Go na porta 8002...")
    go_process = subprocess.Popen(
        [GO_BINARY],
        env=env,
        stdout=sys.stdout,
        stderr=sys.stderr,
    )
    
    yield
    
    # Shutdown
    if ai_process:
        print("Encerrando servico de IA...")
        ai_process.terminate()
        try:
            ai_process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            ai_process.kill()
    
    if go_process:
        print("Encerrando binario Go...")
        go_process.send_signal(signal.SIGTERM)
        try:
            go_process.wait(timeout=10)
        except subprocess.TimeoutExpired:
            go_process.kill()

app = FastAPI(lifespan=lifespan, title="Nexo One Proxy")
app.include_router(router_api)

@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"])
async def proxy(request: Request, path: str):
    target = f"http://127.0.0.1:8002/{path}"
    if request.url.query:
        target += f"?{request.url.query}"
    
    body = await request.body()
    headers = dict(request.headers)
    headers.pop("host", None)
    
    async with httpx.AsyncClient(timeout=60.0) as client:
        try:
            resp = await client.request(
                method=request.method,
                url=target,
                headers=headers,
                content=body,
            )
        except httpx.ConnectError:
            return Response(
                content='{"error":"backend Go nao iniciado - reconstrua com: cd /app && go build -o /app/nexo-one ./cmd/api/"}',
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
