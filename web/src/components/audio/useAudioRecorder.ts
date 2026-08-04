import { useCallback, useEffect, useRef, useState } from "react";

export type RecorderStatus = "idle" | "recording" | "stopped";

interface AudioRecorderState {
  status: RecorderStatus;
  audioBlob: Blob | null;
  error: string | null;
  volume: number;
  start: () => Promise<void>;
  stop: () => void;
  retry: () => Promise<void>;
}

export function useAudioRecorder(): AudioRecorderState {
  const [status, setStatus] = useState<RecorderStatus>("idle");
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [volume, setVolume] = useState(0);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const streamRef = useRef<MediaStream | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const meterFrameRef = useRef<number | null>(null);

  const stopVolumeMeter = useCallback((): void => {
    if (meterFrameRef.current !== null) {
      cancelAnimationFrame(meterFrameRef.current);
      meterFrameRef.current = null;
    }
    audioContextRef.current?.close().catch(() => {});
    audioContextRef.current = null;
    setVolume(0);
  }, []);

  // AnalyserNode sobre o mesmo stream do MediaRecorder — o usuário vê que a
  // gravação começou de verdade e tem noção do volume/ruído antes de enviar.
  const startVolumeMeter = useCallback((stream: MediaStream): void => {
    const audioContext = new AudioContext();
    const source = audioContext.createMediaStreamSource(stream);
    const analyser = audioContext.createAnalyser();
    analyser.fftSize = 256;
    source.connect(analyser);
    audioContextRef.current = audioContext;

    const samples = new Uint8Array(analyser.frequencyBinCount);
    const tick = (): void => {
      analyser.getByteTimeDomainData(samples);
      let sumSquares = 0;
      for (let i = 0; i < samples.length; i++) {
        const normalized = (samples[i] - 128) / 128;
        sumSquares += normalized * normalized;
      }
      const rms = Math.sqrt(sumSquares / samples.length);
      setVolume(Math.min(1, rms * 4));
      meterFrameRef.current = requestAnimationFrame(tick);
    };
    tick();
  }, []);

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
      startVolumeMeter(stream);
    } catch {
      setError("Não foi possível acessar o microfone. Verifique a permissão do navegador.");
    }
  }, [startVolumeMeter]);

  const stop = useCallback((): void => {
    mediaRecorderRef.current?.stop();
    setStatus("stopped");
    stopVolumeMeter();
  }, [stopVolumeMeter]);

  // retry existe pra reduzir "errei -> gravando de novo" a 1 clique só: limpa
  // a prévia da tentativa anterior e já chama start() na sequência, em vez de
  // exigir um clique pra descartar (reset) e outro pra começar a gravar de
  // novo. getUserMedia() dentro de start() não reabre o prompt de permissão —
  // o browser já lembra a concessão feita na 1ª gravação desta sessão.
  const retry = useCallback(async (): Promise<void> => {
    setAudioBlob(null);
    await start();
  }, [start]);

  useEffect(() => stopVolumeMeter, [stopVolumeMeter]);

  return { status, audioBlob, error, volume, start, stop, retry };
}
