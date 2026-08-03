import os
import tempfile
from contextlib import asynccontextmanager

from fastapi import FastAPI, File, HTTPException, UploadFile

from whisper_engine import get_model, transcribe

ALLOWED_EXTENSIONS = {".wav", ".webm", ".mp3", ".m4a", ".ogg"}


@asynccontextmanager
async def lifespan(app: FastAPI):
    get_model()  # carrega o modelo no startup, não na primeira requisição
    yield


app = FastAPI(title="yamabiko stt-service", lifespan=lifespan)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok"}


@app.post("/transcribe")
async def transcribe_audio(file: UploadFile = File(...)) -> dict:
    ext = os.path.splitext(file.filename or "")[1].lower()
    if ext not in ALLOWED_EXTENSIONS:
        raise HTTPException(status_code=400, detail=f"Formato não suportado: {ext or 'desconhecido'}")

    with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
        tmp.write(await file.read())
        tmp_path = tmp.name

    try:
        return transcribe(tmp_path)
    finally:
        os.unlink(tmp_path)
