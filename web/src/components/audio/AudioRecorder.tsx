import { useEffect, useMemo } from "react";
import { useAudioRecorder } from "./useAudioRecorder";

interface AudioRecorderProps {
  onRecorded: (blob: Blob) => void;
  disabled?: boolean;
  /** Começa gravando assim que montado, sem exigir o clique em "🎙 Gravar" —
   * usado pelo botão "Tentar de novo" da tela de resultado (ExercisePage),
   * que força um remount deste componente via `key`. */
  autoStart?: boolean;
}

export function AudioRecorder({ onRecorded, disabled, autoStart }: AudioRecorderProps) {
  const { status, audioBlob, error, volume, start, stop, retry } = useAudioRecorder();

  useEffect(() => {
    if (autoStart) void start();
    // Só no mount (key={attemptNumber} força um AudioRecorder novo a cada
    // retry) — repetir por causa de `start` mudar de identidade a cada
    // render reiniciaria a gravação sozinho.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const previewUrl = useMemo(() => (audioBlob ? URL.createObjectURL(audioBlob) : null), [audioBlob]);

  if (audioBlob && previewUrl) {
    return (
      <div className="audio-recorder">
        <audio controls src={previewUrl} style={{ width: "100%", maxWidth: 300 }} />
        <div className="audio-recorder-actions" style={{ width: "100%" }}>
          <button type="button" className="btn-primary" onClick={() => onRecorded(audioBlob)} disabled={disabled}>
            Enviar
          </button>
          {/* 1 clique só: limpa a prévia e já começa a gravar de novo (retry),
              em vez de exigir um 2º clique em "🎙 Gravar" depois de descartar. */}
          <button type="button" className="btn-secondary" onClick={() => void retry()} disabled={disabled}>
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
          <button type="button" className="record-button-ring recording" onClick={stop} aria-label="Parar gravação">
            <span className="record-button-core">
              <span className="record-button-core-icon" />
            </span>
          </button>
          <span className="recording-label">● Parar gravação</span>
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
        <button type="button" className="record-button-ring" onClick={start} disabled={disabled} aria-label="Gravar">
          <span className="record-button-core">
            <span className="record-button-core-icon" />
          </span>
        </button>
      )}
      {status !== "recording" && <span className="recording-label">🎙 Gravar</span>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}
