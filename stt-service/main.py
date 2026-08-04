import os
import tempfile
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, File, Form, HTTPException, UploadFile

from whisper_engine import get_model, transcribe

ALLOWED_EXTENSIONS = {".wav", ".webm", ".mp3", ".m4a", ".ogg"}
DEFAULT_WHISPER_LANGUAGE = "ja"


def resolve_whisper_language(language: Optional[str]) -> str:
    """Converte tags tipo 'ja-JP'/'en-US' (usadas em exercises.language) pro
    código de 2 letras que o faster-whisper espera. Sem `language` (chamador
    antigo) cai no default ja, preservando o comportamento histórico."""
    if not language:
        return DEFAULT_WHISPER_LANGUAGE
    return language.split("-")[0].lower()


@asynccontextmanager
async def lifespan(app: FastAPI):
    get_model()  # carrega o modelo no startup, não na primeira requisição
    yield


app = FastAPI(title="yamabiko stt-service", lifespan=lifespan)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok"}


@app.post("/transcribe")
async def transcribe_audio(file: UploadFile = File(...), language: Optional[str] = Form(None)) -> dict:
    ext = os.path.splitext(file.filename or "")[1].lower()
    if ext not in ALLOWED_EXTENSIONS:
        raise HTTPException(status_code=400, detail=f"Formato não suportado: {ext or 'desconhecido'}")

    with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
        tmp.write(await file.read())
        tmp_path = tmp.name

    try:
        return transcribe(tmp_path, language=resolve_whisper_language(language))
    finally:
        os.unlink(tmp_path)
