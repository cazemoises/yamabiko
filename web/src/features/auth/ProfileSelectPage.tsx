import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";
import { fetchPinProfiles, type PinProfile } from "./profilesApi";
import { ApiError } from "../../lib/apiClient";

const PIN_LENGTH = 6;

function accentStyle(accentColor: string | undefined): React.CSSProperties {
  const accent = accentColor === "mono" ? "var(--text)" : (accentColor ?? "var(--accent-base)");
  return { "--profile-accent": accent } as React.CSSProperties;
}

function formatCountdown(secondsRemaining: number): string {
  const minutes = Math.floor(secondsRemaining / 60);
  const seconds = secondsRemaining % 60;
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}

export function ProfileSelectPage() {
  const { pinLogin } = useAuth();
  const navigate = useNavigate();

  const [profiles, setProfiles] = useState<PinProfile[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selected, setSelected] = useState<PinProfile | null>(null);
  const [pin, setPin] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [attemptsRemaining, setAttemptsRemaining] = useState<number | null>(null);
  const [lockedUntil, setLockedUntil] = useState<number | null>(null);
  const [secondsRemaining, setSecondsRemaining] = useState(0);

  useEffect(() => {
    fetchPinProfiles()
      .then(setProfiles)
      .catch(() => setLoadError("Não foi possível carregar os perfis."));
  }, []);

  useEffect(() => {
    if (lockedUntil === null) return;
    const tick = () => {
      const remaining = Math.max(0, Math.ceil((lockedUntil - Date.now()) / 1000));
      setSecondsRemaining(remaining);
      if (remaining === 0) {
        setLockedUntil(null);
        setError(null);
      }
    };
    tick();
    const interval = setInterval(tick, 1000);
    return () => clearInterval(interval);
  }, [lockedUntil]);

  const submittedRef = useRef(false);
  useEffect(() => {
    if (!selected || pin.length !== PIN_LENGTH || submittedRef.current) return;
    submittedRef.current = true;
    setSubmitting(true);
    setError(null);

    pinLogin(selected.id, pin)
      .then(() => navigate("/"))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 429) {
          const retryAfterSeconds = Number(err.body?.retry_after_seconds ?? 0);
          setLockedUntil(Date.now() + retryAfterSeconds * 1000);
          setError("Conta temporariamente bloqueada por tentativas incorretas.");
          setAttemptsRemaining(null);
        } else if (err instanceof ApiError) {
          const remaining = err.body?.attempts_remaining;
          setAttemptsRemaining(typeof remaining === "number" ? remaining : null);
          setError("PIN incorreto.");
        } else {
          setError("Erro ao autenticar. Tente novamente.");
        }
        setPin("");
      })
      .finally(() => {
        setSubmitting(false);
        submittedRef.current = false;
      });
  }, [pin, selected, pinLogin, navigate]);

  function selectProfile(profile: PinProfile): void {
    setSelected(profile);
    setPin("");
    setError(null);
    setAttemptsRemaining(null);
    setLockedUntil(null);
  }

  function backToProfiles(): void {
    setSelected(null);
    setPin("");
    setError(null);
    setAttemptsRemaining(null);
    setLockedUntil(null);
  }

  function pressDigit(digit: string): void {
    if (submitting || lockedUntil !== null) return;
    setPin((prev) => (prev.length < PIN_LENGTH ? prev + digit : prev));
  }

  function pressBackspace(): void {
    if (submitting || lockedUntil !== null) return;
    setPin((prev) => prev.slice(0, -1));
  }

  useEffect(() => {
    if (!selected) return;
    function onKeyDown(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        backToProfiles();
        return;
      }
      if (/^[0-9]$/.test(event.key)) {
        pressDigit(event.key);
      } else if (event.key === "Backspace") {
        pressBackspace();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, submitting, lockedUntil]);

  if (loadError) {
    return <p className="center-message">{loadError}</p>;
  }

  if (!profiles) {
    return <p className="center-message">Carregando perfis...</p>;
  }

  if (!selected) {
    return (
      <div className="profile-select">
        <h1>やまびこ</h1>
        {profiles.length === 0 ? (
          <p className="center-message">
            Nenhum perfil com PIN configurado ainda. <Link to="/login/password">Entrar com email e senha</Link>.
          </p>
        ) : (
          <div className="profile-grid">
            {profiles.map((profile) => (
              <button
                key={profile.id}
                type="button"
                className="profile-card"
                style={accentStyle(profile.accent_color)}
                onClick={() => selectProfile(profile)}
              >
                <span className="profile-card-avatar">{profile.display_name.charAt(0).toUpperCase()}</span>
                <span className="profile-card-name">{profile.display_name}</span>
              </button>
            ))}
          </div>
        )}
        <Link to="/login/password" className="profile-select-fallback">
          usar outro método
        </Link>
      </div>
    );
  }

  return (
    <div className="profile-select" style={accentStyle(selected.accent_color)}>
      <button type="button" className="pin-back" onClick={backToProfiles} aria-label="Voltar">
        ←
      </button>
      <span className="profile-card-avatar pin-avatar">{selected.display_name.charAt(0).toUpperCase()}</span>
      <h1>{selected.display_name}</h1>

      <div className="pin-dots" aria-hidden="true">
        {Array.from({ length: PIN_LENGTH }, (_, i) => (
          <span key={i} className={i < pin.length ? "pin-dot filled" : "pin-dot"} />
        ))}
      </div>

      {lockedUntil !== null ? (
        <p className="error">Bloqueado — tente novamente em {formatCountdown(secondsRemaining)}</p>
      ) : (
        error && (
          <p className="error">
            {error}
            {attemptsRemaining !== null && ` ${attemptsRemaining} tentativa(s) restante(s).`}
          </p>
        )
      )}

      <div className="pin-keypad">
        {["1", "2", "3", "4", "5", "6", "7", "8", "9"].map((digit) => (
          <button
            key={digit}
            type="button"
            className="pin-key"
            disabled={submitting || lockedUntil !== null}
            onClick={() => pressDigit(digit)}
          >
            {digit}
          </button>
        ))}
        <span />
        <button
          type="button"
          className="pin-key"
          disabled={submitting || lockedUntil !== null}
          onClick={() => pressDigit("0")}
        >
          0
        </button>
        <button
          type="button"
          className="pin-key pin-key-backspace"
          disabled={submitting || lockedUntil !== null || pin.length === 0}
          onClick={pressBackspace}
          aria-label="Apagar"
        >
          ⌫
        </button>
      </div>

      <Link to="/login/password" className="profile-select-fallback">
        usar outro método
      </Link>
    </div>
  );
}
