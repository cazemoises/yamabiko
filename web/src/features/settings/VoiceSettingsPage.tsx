import { useEffect, useRef, useState } from "react";
import { getVoicePreview, listVoices, type Voice } from "../tts/api";
import { getProfile, updateVoicePreference, type Profile } from "../users/api";
import { ACCENT_PRESETS, useAppearance } from "../users/AppearanceContext";

const LANGUAGES = [
  { value: "ja-JP", label: "Japonês" },
  { value: "en-US", label: "Inglês" },
];

const THEME_OPTIONS = [
  { value: "", label: "Sistema" },
  { value: "light", label: "Claro" },
  { value: "dark", label: "Escuro" },
];

// Não é um frame do design (o Yamabiko.dc.html só tem um painel de Tweaks
// pro preview do PRÓPRIO design, não uma tela de app pro usuário final) —
// composição nossa sobre os mesmos tokens/primitivos (.theme-toggle já
// existia pro toggle de idioma da voz), documentada como decisão em
// BUILD_STATE.md.
function AppearanceSection() {
  const { theme, accentColor, setTheme, setAccentColor } = useAppearance();
  const [customHex, setCustomHex] = useState("");

  function applyCustomHex(): void {
    if (/^#[0-9a-fA-F]{6}$/.test(customHex)) setAccentColor(customHex);
  }

  return (
    <div className="appearance-section">
      <span className="section-title">Aparência</span>

      <div className="theme-toggle" role="group" aria-label="Tema">
        {THEME_OPTIONS.map((opt) => (
          <button
            key={opt.value || "system"}
            type="button"
            className={theme === opt.value ? "theme-toggle-button active" : "theme-toggle-button"}
            aria-pressed={theme === opt.value}
            onClick={() => setTheme(opt.value)}
          >
            {opt.label}
          </button>
        ))}
      </div>

      <div className="accent-swatches">
        {ACCENT_PRESETS.map((preset) => (
          <button
            key={preset.id}
            type="button"
            className={accentColor === preset.value ? "accent-swatch selected" : "accent-swatch"}
            style={{ background: preset.swatch }}
            aria-label={`Acento ${preset.label}`}
            aria-pressed={accentColor === preset.value}
            onClick={() => setAccentColor(preset.value)}
          />
        ))}
        <input
          type="text"
          className="accent-custom-input"
          placeholder="#RRGGBB"
          value={customHex}
          onChange={(e) => setCustomHex(e.target.value)}
          onBlur={applyCustomHex}
          onKeyDown={(e) => e.key === "Enter" && applyCustomHex()}
          aria-label="Cor de acento customizada"
        />
      </div>
    </div>
  );
}

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
        <span className="page-title">Configurações</span>
      </div>

      <AppearanceSection />

      <span className="section-title">Voz de referência</span>
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
