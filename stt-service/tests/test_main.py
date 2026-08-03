from unittest.mock import patch

from fastapi.testclient import TestClient

import main


def test_health():
    with patch.object(main, "get_model", return_value=None):
        with TestClient(main.app) as client:
            response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_transcribe_returns_result():
    fake_result = {"transcript": "こんにちは", "language": "ja", "confidence": 0.98}
    with patch.object(main, "get_model", return_value=None), patch.object(
        main, "transcribe", return_value=fake_result
    ):
        with TestClient(main.app) as client:
            response = client.post(
                "/transcribe",
                files={"file": ("test.wav", b"fake-audio-bytes", "audio/wav")},
            )
    assert response.status_code == 200
    assert response.json() == fake_result


def test_transcribe_rejects_unsupported_extension():
    with patch.object(main, "get_model", return_value=None):
        with TestClient(main.app) as client:
            response = client.post(
                "/transcribe",
                files={"file": ("test.txt", b"not audio", "text/plain")},
            )
    assert response.status_code == 400
