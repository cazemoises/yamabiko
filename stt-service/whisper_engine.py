"""Engine de transcrição. Params validados em laboratório (mic_test.py / transcribe.py)."""
import os
from typing import Optional

os.environ.setdefault("HF_HUB_DISABLE_SYMLINKS_WARNING", "1")

from faster_whisper import WhisperModel

MODEL_SIZE = "large-v3-turbo"
DEVICE = "cpu"
COMPUTE_TYPE = "int8"

_model: Optional[WhisperModel] = None


def get_model() -> WhisperModel:
    global _model
    if _model is None:
        _model = WhisperModel(MODEL_SIZE, device=DEVICE, compute_type=COMPUTE_TYPE)
    return _model


def transcribe(audio_path: str) -> dict:
    model = get_model()
    segments, info = model.transcribe(
        audio_path,
        beam_size=5,
        language="ja",
        task="transcribe",
        vad_filter=True,
        vad_parameters=dict(min_silence_duration_ms=500),
        condition_on_previous_text=False,
    )
    text = " ".join(segment.text for segment in segments).strip()
    return {
        "transcript": text,
        "language": info.language,
        "confidence": info.language_probability,
    }
