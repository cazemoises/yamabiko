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


def test_transcribe_passes_language_to_engine():
    fake_result = {"transcript": "hi", "language": "en", "confidence": 0.9}
    with patch.object(main, "get_model", return_value=None), patch.object(
        main, "transcribe", return_value=fake_result
    ) as mock_transcribe:
        with TestClient(main.app) as client:
            response = client.post(
                "/transcribe",
                files={"file": ("test.wav", b"fake-audio-bytes", "audio/wav")},
                data={"language": "en-US"},
            )
    assert response.status_code == 200
    mock_transcribe.assert_called_once()
    _, kwargs = mock_transcribe.call_args
    assert kwargs["language"] == "en"


def test_transcribe_defaults_to_japanese_when_language_omitted():
    fake_result = {"transcript": "こんにちは", "language": "ja", "confidence": 0.98}
    with patch.object(main, "get_model", return_value=None), patch.object(
        main, "transcribe", return_value=fake_result
    ) as mock_transcribe:
        with TestClient(main.app) as client:
            client.post(
                "/transcribe",
                files={"file": ("test.wav", b"fake-audio-bytes", "audio/wav")},
            )
    _, kwargs = mock_transcribe.call_args
    assert kwargs["language"] == "ja"


def test_resolve_whisper_language_maps_bcp47_tags():
    assert main.resolve_whisper_language("ja-JP") == "ja"
    assert main.resolve_whisper_language("en-US") == "en"
    assert main.resolve_whisper_language(None) == "ja"
    assert main.resolve_whisper_language("") == "ja"
