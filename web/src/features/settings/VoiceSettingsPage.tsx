import { useEffect, useRef, useState } from "react";
import { getVoicePreview, listVoices, type Voice } from "../tts/api";
import { getProfile, updateVoicePreference, type Profile } from "../users/api";

const LANGUAGES = [
  { value: "ja-JP", label: "🇯🇵 Japonês" },
  { value: "en-US", label: "🇺🇸 Inglês" },
];

export function VoiceSettingsPage() {
  const [language, setLanguage] = useState("ja-JP");
  const [voices, setVoices] = useState<Voice[]>([]);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [previewingId, setPreviewingId] = useState<string | null>(null);
  const [savingId, setSavingId] = useState<string | null>(null);

  const audioRef = useRef<HTMLAudioElement | null>(null);
  const objectUrlRef = useRef<string | null>(null);

  useEffect(() => {
    getProfile()
      .then(setProfile)
      .catch(() => setError("Erro ao carregar perfil"));
  }, []);

  useEffect(() => {
    setLoading(true);
    listVoices(language)
      .then(setVoices)
      .catch(() => setError("Erro ao carregar vozes"))
      .finally(() => setLoading(false));
  }, [language]);

  useEffect(() => {
    // Libera o object URL do preview anterior ao desmontar — mesmo cuidado
    // de SpeakButton, senão cada preview ouvido fica retido em memória.
    return () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
    };
  }, []);

  const selectedVoiceId = language === "ja-JP" ? profile?.preferred_voice_ja : profile?.preferred_voice_en;

  async function preview(voiceId: string): Promise<void> {
    if (previewingId) return;
    setPreviewingId(voiceId);
    setError(null);
    try {
      const blob = await getVoicePreview(voiceId);
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
      const url = URL.createObjectURL(blob);
      objectUrlRef.current = url;
      if (audioRef.current) {
        audioRef.current.src = url;
        await audioRef.current.play();
      }
    } catch {
      setError("Erro ao carregar preview de voz");
    } finally {
      setPreviewingId(null);
    }
  }

  async function select(voiceId: string): Promise<void> {
    setSavingId(voiceId);
    setError(null);
    try {
      await updateVoicePreference(language, voiceId);
      setProfile((prev) =>
        prev === null
          ? prev
          : {
              ...prev,
              ...(language === "ja-JP" ? { preferred_voice_ja: voiceId } : { preferred_voice_en: voiceId }),
            },
      );
    } catch {
      setError("Erro ao salvar preferência de voz");
    } finally {
      setSavingId(null);
    }
  }

  return (
    <div className="voice-settings-page">
      <h1>Escolher voz</h1>
      <div className="language-toggle" role="group" aria-label="Idioma da voz">
        {LANGUAGES.map((lang) => (
          <button
            key={lang.value}
            type="button"
            className={lang.value === language ? "language-toggle-button active" : "language-toggle-button"}
            aria-pressed={lang.value === language}
            onClick={() => setLanguage(lang.value)}
          >
            {lang.label}
          </button>
        ))}
      </div>

      {error && <p className="error">{error}</p>}
      {loading && <p>Carregando vozes...</p>}

      {!loading && (
        <ul className="voice-list">
          {voices.map((voice) => {
            const isSelected = voice.id === selectedVoiceId;
            return (
              <li key={voice.id} className={isSelected ? "voice-row voice-row-selected" : "voice-row"}>
                <span className="voice-name">{voice.name}</span>
                <div className="voice-row-actions">
                  <button
                    type="button"
                    className="voice-preview-button"
                    onClick={() => preview(voice.id)}
                    disabled={previewingId === voice.id}
                  >
                    {previewingId === voice.id ? "Carregando..." : "▶ Ouvir"}
                  </button>
                  <button
                    type="button"
                    className={isSelected ? "voice-select-button voice-select-button-active" : "voice-select-button"}
                    onClick={() => select(voice.id)}
                    disabled={isSelected || savingId === voice.id}
                  >
                    {isSelected ? "✓ Selecionada" : "Selecionar"}
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      )}

      <audio ref={audioRef} data-testid="voice-preview-audio" style={{ display: "none" }} />
    </div>
  );
}
