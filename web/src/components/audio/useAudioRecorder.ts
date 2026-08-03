import { useCallback, useRef, useState } from "react";

export type RecorderStatus = "idle" | "recording" | "stopped";

interface AudioRecorderState {
  status: RecorderStatus;
  audioBlob: Blob | null;
  error: string | null;
  start: () => Promise<void>;
  stop: () => void;
  reset: () => void;
}

export function useAudioRecorder(): AudioRecorderState {
  const [status, setStatus] = useState<RecorderStatus>("idle");
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
  const [error, setError] = useState<string | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const streamRef = useRef<MediaStream | null>(null);

  const start = useCallback(async (): Promise<void> => {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;
      chunksRef.current = [];

      const recorder = new MediaRecorder(stream);
      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunksRef.current.push(event.data);
        }
      };
      recorder.onstop = () => {
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        setAudioBlob(blob);
        streamRef.current?.getTracks().forEach((track) => track.stop());
      };

      mediaRecorderRef.current = recorder;
      recorder.start();
      setStatus("recording");
    } catch {
      setError("Não foi possível acessar o microfone. Verifique a permissão do navegador.");
    }
  }, []);

  const stop = useCallback((): void => {
    mediaRecorderRef.current?.stop();
    setStatus("stopped");
  }, []);

  const reset = useCallback((): void => {
    setAudioBlob(null);
    setStatus("idle");
  }, []);

  return { status, audioBlob, error, start, stop, reset };
}
