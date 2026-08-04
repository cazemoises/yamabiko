import { useEffect, useRef, useState } from "react";
import { api, ApiError } from "../../lib/apiClient";

interface SpeakButtonProps {
  text: string;
  lang?: string;
  exerciseId?: string;
  label?: string;
  className?: string;
}

function isJapanese(lang: string): boolean {
  return lang.toLowerCase().startsWith("ja");
}

// hasVoiceForLang compara só o subtag primário (ex: "ja" de "ja-JP") porque
// browsers variam a região da voz instalada (ex: "en-GB" quando pedimos
// "en-US") — exigir match exato de região faria o botão ficar desabilitado à
// toa em sistemas com voz do idioma certo mas região diferente.
function hasVoiceForLang(lang: string): boolean {
  if (typeof window === "undefined" || !window.speechSynthesis) return false;
  const primary = lang.split("-")[0].toLowerCase();
  return window.speechSynthesis.getVoices().some((voice) => voice.lang.toLowerCase().startsWith(primary));
}

// Botão "Ouvir pronúncia esperada". Pra ja-JP, busca o áudio real gerado pelo
// VOICEVOX (GET /exercises/{id}/reference-audio, cacheado em disco no
// core-api) e toca via elemento <audio> — qualidade e disponibilidade
// consistentes entre alunos, ao contrário da Web Speech API (débito antigo
// documentado em BUILD_STATE.md, resolvido por este componente). Pra outros
// idiomas (en-US), o VOICEVOX não fala, então mantém a Web Speech API nativa
// do browser, sem mudança de comportamento.
export function SpeakButton({
  text,
  lang = "ja-JP",
  exerciseId,
  label = "🔊 Ouvir pronúncia esperada",
  className,
}: SpeakButtonProps) {
  const japanese = isJapanese(lang);
  const [available, setAvailable] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const objectUrlRef = useRef<string | null>(null);

  useEffect(() => {
    if (japanese) return; // disponibilidade do VOICEVOX é responsabilidade do backend, não do browser
    if (typeof window === "undefined" || !window.speechSynthesis) {
      setAvailable(false);
      return;
    }
    const update = (): void => setAvailable(hasVoiceForLang(lang));
    update();
    window.speechSynthesis.addEventListener("voiceschanged", update);
    return () => window.speechSynthesis.removeEventListener("voiceschanged", update);
  }, [lang, japanese]);

  useEffect(() => {
    // Libera o object URL anterior ao desmontar/trocar de exercício, senão
    // vaza memória (cada blob gerado fica retido até isso rodar).
    return () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
    };
  }, []);

  async function speakJapanese(): Promise<void> {
    if (!exerciseId || loading) return;
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

  function speakOther(): void {
    if (!available || !text) return;
    window.speechSynthesis.cancel(); // evita empilhar falas se o aluno clicar de novo rápido
    const utterance = new SpeechSynthesisUtterance(text);
    utterance.lang = lang;
    window.speechSynthesis.speak(utterance);
  }

  const disabled = japanese ? !exerciseId || loading : !available;
  const title = japanese
    ? (error ?? undefined)
    : available
      ? undefined
      : "Nenhuma voz nesse idioma disponível neste navegador/sistema";

  return (
    <>
      <button
        type="button"
        className={className ? `speak-button ${className}` : "speak-button"}
        onClick={japanese ? speakJapanese : speakOther}
        disabled={disabled}
        title={title}
      >
        {label}
      </button>
      {japanese && <audio ref={audioRef} data-testid="reference-audio" style={{ display: "none" }} />}
    </>
  );
}
