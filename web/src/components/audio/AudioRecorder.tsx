import { useMemo } from "react";
import { useAudioRecorder } from "./useAudioRecorder";

interface AudioRecorderProps {
  onRecorded: (blob: Blob) => void;
  disabled?: boolean;
}

export function AudioRecorder({ onRecorded, disabled }: AudioRecorderProps) {
  const { status, audioBlob, error, volume, start, stop, retry } = useAudioRecorder();

  const previewUrl = useMemo(() => (audioBlob ? URL.createObjectURL(audioBlob) : null), [audioBlob]);

  if (audioBlob && previewUrl) {
    return (
      <div className="audio-recorder">
        <audio controls src={previewUrl} />
        <div className="audio-recorder-actions">
          <button type="button" onClick={() => onRecorded(audioBlob)} disabled={disabled}>
            Enviar
          </button>
          {/* 1 clique só: limpa a prévia e já começa a gravar de novo (retry),
              em vez de exigir um 2º clique em "🎙 Gravar" depois de descartar. */}
          <button type="button" onClick={() => void retry()} disabled={disabled}>
            Tentar de novo
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="audio-recorder">
      {status === "recording" ? (
        <>
          <button type="button" className="recording" onClick={stop}>
            ● Parar gravação
          </button>
          <div
            className="volume-meter"
            data-testid="volume-meter"
            data-volume={volume}
            role="progressbar"
            aria-label="Volume do microfone"
            aria-valuemin={0}
            aria-valuemax={1}
            aria-valuenow={volume}
          >
            <div className="volume-meter-bar" style={{ width: `${volume * 100}%` }} />
          </div>
        </>
      ) : (
        <button type="button" onClick={start} disabled={disabled}>
          🎙 Gravar
        </button>
      )}
      {error && <p className="error">{error}</p>}
    </div>
  );
}
