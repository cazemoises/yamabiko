import { useEffect, useRef, useState } from "react";
import { getVoicePreview, listVoices, type Voice } from "../tts/api";
import { getProfile, updateVoicePreference, type Profile } from "../users/api";

const LANGUAGES = [
  { value: "ja-JP", label: "Japonês" },
  { value: "en-US", label: "Inglês" },
];

// Frame 8 (Configurações de Voz) — cada linha tem o nome da voz (já
// descritivo, ex: "Voz Feminina Grave" — o catálogo curado não tem um
// campo separado de gênero/tom pra mostrar como subtítulo, diferente do
// mock "Yuki — Feminina · tom claro" do design), botão de preview e um
// indicador circular de seleção (anel vazio -> bolinha preenchida com
// check).
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
    <div className="page">
      <div className="page-header">
        <span className="page-title">Voz de referência</span>
      </div>

      <div className="theme-toggle" role="group" aria-label="Idioma da voz">
        {LANGUAGES.map((lang) => (
          <button
            key={lang.value}
            type="button"
            className={lang.value === language ? "theme-toggle-button active" : "theme-toggle-button"}
            aria-pressed={lang.value === language}
            onClick={() => setLanguage(lang.value)}
          >
            {lang.label}
          </button>
        ))}
      </div>
      <p className="page-subtitle">
        Usada nos áudios de pronúncia em {language === "ja-JP" ? "japonês" : "inglês"}
      </p>

      {error && <p className="error">{error}</p>}
      {loading && <p className="center-message">Carregando vozes...</p>}

      {!loading && (
        <ul className="plain-list voice-list">
          {voices.map((voice) => {
            const isSelected = voice.id === selectedVoiceId;
            return (
              <li key={voice.id} className={isSelected ? "voice-row voice-row-selected" : "voice-row"}>
                <div className="voice-row-info">
                  <span className="voice-row-name">{voice.name}</span>
                </div>
                <div className="voice-row-actions">
                  <button
                    type="button"
                    className="voice-preview-button"
                    onClick={() => preview(voice.id)}
                    disabled={previewingId === voice.id}
                  >
                    {previewingId === voice.id ? "..." : "▶ Ouvir"}
                  </button>
                  <button
                    type="button"
                    className={isSelected ? "voice-select-button selected" : "voice-select-button"}
                    aria-label={isSelected ? "Voz selecionada" : `Selecionar ${voice.name}`}
                    onClick={() => select(voice.id)}
                    disabled={isSelected || savingId === voice.id}
                  >
                    {isSelected && <CheckIcon />}
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

function CheckIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14">
      <line x1="3" y1="7.2" x2="5.5" y2="9.7" stroke="var(--accent-on)" strokeWidth="1.8" />
      <line x1="5.5" y1="9.7" x2="11" y2="4" stroke="var(--accent-on)" strokeWidth="1.8" />
    </svg>
  );
}
