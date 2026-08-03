import { useMemo } from "react";
import { useAudioRecorder } from "./useAudioRecorder";

interface AudioRecorderProps {
  onRecorded: (blob: Blob) => void;
  disabled?: boolean;
}

export function AudioRecorder({ onRecorded, disabled }: AudioRecorderProps) {
  const { status, audioBlob, error, start, stop, reset } = useAudioRecorder();

  const previewUrl = useMemo(() => (audioBlob ? URL.createObjectURL(audioBlob) : null), [audioBlob]);

  if (audioBlob && previewUrl) {
    return (
      <div className="audio-recorder">
        <audio controls src={previewUrl} />
        <div className="audio-recorder-actions">
          <button type="button" onClick={() => onRecorded(audioBlob)} disabled={disabled}>
            Enviar
          </button>
          <button type="button" onClick={reset} disabled={disabled}>
            Gravar de novo
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="audio-recorder">
      {status === "recording" ? (
        <button type="button" className="recording" onClick={stop}>
          ● Parar gravação
        </button>
      ) : (
        <button type="button" onClick={start} disabled={disabled}>
          🎙 Gravar
        </button>
      )}
      {error && <p className="error">{error}</p>}
    </div>
  );
}
