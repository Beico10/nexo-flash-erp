import subprocess
import signal
import sys
import os
import httpx
from fastapi import FastAPI, Request, Response
from fastapi.middleware.cors import CORSMiddleware

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
