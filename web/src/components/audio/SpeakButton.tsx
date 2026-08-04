import { useEffect, useState } from "react";

interface SpeakButtonProps {
  text: string;
  label?: string;
  className?: string;
}

function hasJapaneseVoice(): boolean {
  if (typeof window === "undefined" || !window.speechSynthesis) return false;
  return window.speechSynthesis.getVoices().some((voice) => voice.lang.toLowerCase().startsWith("ja"));
}

// Botão "Ouvir pronúncia esperada" via Web Speech API nativa do browser — sem
// dependência externa/custo de TTS, mas a lista de vozes só vem populada de
// verdade depois do evento "voiceschanged" em alguns browsers (Chrome), daí o
// listener em vez de checar getVoices() só uma vez no mount.
export function SpeakButton({ text, label = "🔊 Ouvir pronúncia esperada", className }: SpeakButtonProps) {
  const [available, setAvailable] = useState(false);

  useEffect(() => {
    if (typeof window === "undefined" || !window.speechSynthesis) {
      setAvailable(false);
      return;
    }
    const update = (): void => setAvailable(hasJapaneseVoice());
    update();
    window.speechSynthesis.addEventListener("voiceschanged", update);
    return () => window.speechSynthesis.removeEventListener("voiceschanged", update);
  }, []);

  function speak(): void {
    if (!available || !text) return;
    window.speechSynthesis.cancel(); // evita empilhar falas se o aluno clicar de novo rápido
    const utterance = new SpeechSynthesisUtterance(text);
    utterance.lang = "ja-JP";
    window.speechSynthesis.speak(utterance);
  }

  return (
    <button
      type="button"
      className={className ? `speak-button ${className}` : "speak-button"}
      onClick={speak}
      disabled={!available}
      title={available ? undefined : "Nenhuma voz em japonês disponível neste navegador/sistema"}
    >
      {label}
    </button>
  );
}
