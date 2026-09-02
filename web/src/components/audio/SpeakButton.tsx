import { useEffect, useRef, useState } from "react";
import { api, ApiError } from "../../lib/apiClient";

interface SpeakButtonProps {
  exerciseId: string;
  label?: string;
  className?: string;
  /** Troca a classe base "speak-button" por outra — usado quando o botão
   * de ouvir pronúncia precisa de um visual totalmente diferente (ex: o
   * círculo grande de play do Frame 12/ditado) em vez de só adicionar um
   * modificador em cima do botão retangular padrão. */
  baseClassName?: string;
}

// Toca o áudio de referência real do exercício — GET
// /exercises/{id}/reference-audio, cacheado em disco no core-api, gerado por
// VOICEVOX (ja-JP) ou Piper (en-US) conforme o idioma — via um elemento
// <audio>. Substituiu de vez a Web Speech API nativa do browser (primeiro só
// pra ja-JP, depois generalizado pra en-US também), cuja qualidade e
// disponibilidade de voz variavam demais entre alunos e sistemas (débito
// documentado em BUILD_STATE.md).
export function SpeakButton({
  exerciseId,
  label = "🔊 ouvir pronúncia esperada",
  className,
  baseClassName = "speak-button",
}: SpeakButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const objectUrlRef = useRef<string | null>(null);

  useEffect(() => {
    // Libera o object URL anterior ao desmontar/trocar de exercício, senão
    // vaza memória (cada blob gerado fica retido até isso rodar).
    return () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
    };
  }, []);

  async function speak(): Promise<void> {
    if (loading) return;
    setLoading(true);
    setError(null);
    try {
      const blob = await api.getBlob(`/exercises/${exerciseId}/reference-audio`);
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
      const url = URL.createObjectURL(blob);
      objectUrlRef.current = url;
      if (audioRef.current) {
        audioRef.current.src = url;
        await audioRef.current.play();
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao carregar áudio de referência");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <button
        type="button"
        className={className ? `${baseClassName} ${className}` : baseClassName}
        onClick={speak}
        disabled={loading}
        title={error ?? undefined}
      >
        {label}
      </button>
      <audio ref={audioRef} data-testid="reference-audio" style={{ display: "none" }} />
    </>
  );
}
