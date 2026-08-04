import { useEffect, useState } from "react";

interface SpeakButtonProps {
  text: string;
  lang?: string;
  label?: string;
  className?: string;
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

// Botão "Ouvir pronúncia esperada" via Web Speech API nativa do browser — sem
// dependência externa/custo de TTS, mas a lista de vozes só vem populada de
// verdade depois do evento "voiceschanged" em alguns browsers (Chrome), daí o
// listener em vez de checar getVoices() só uma vez no mount.
export function SpeakButton({ text, lang = "ja-JP", label = "🔊 Ouvir pronúncia esperada", className }: SpeakButtonProps) {
  const [available, setAvailable] = useState(false);

  useEffect(() => {
    if (typeof window === "undefined" || !window.speechSynthesis) {
      setAvailable(false);
      return;
    }
    const update = (): void => setAvailable(hasVoiceForLang(lang));
    update();
    window.speechSynthesis.addEventListener("voiceschanged", update);
    return () => window.speechSynthesis.removeEventListener("voiceschanged", update);
  }, [lang]);

  function speak(): void {
    if (!available || !text) return;
    window.speechSynthesis.cancel(); // evita empilhar falas se o aluno clicar de novo rápido
    const utterance = new SpeechSynthesisUtterance(text);
    utterance.lang = lang;
    window.speechSynthesis.speak(utterance);
  }

  return (
    <button
      type="button"
      className={className ? `speak-button ${className}` : "speak-button"}
      onClick={speak}
      disabled={!available}
      title={available ? undefined : "Nenhuma voz nesse idioma disponível neste navegador/sistema"}
    >
      {label}
    </button>
  );
}
