const LANGUAGES = [
  { value: "ja-JP", label: "JA" },
  { value: "en-US", label: "EN" },
];

interface LanguageToggleProps {
  language: string;
  onChange: (language: string) => void;
}

// Toggle compacto JA/EN usado no topo das telas de navegação (Home,
// Cenários, Exercícios avulsos, Configurações de voz) — visual do design
// (2 botões dentro de uma pílula com borda), não confundir com o
// language-toggle maior de ExercisesListPage que usa rótulo + bandeira.
export function LanguageToggle({ language, onChange }: LanguageToggleProps) {
  return (
    <div className="topbar-lang-toggle" role="group" aria-label="Idioma">
      {LANGUAGES.map((lang) => (
        <button
          key={lang.value}
          type="button"
          className={lang.value === language ? "active" : undefined}
          aria-pressed={lang.value === language}
          onClick={() => onChange(lang.value)}
        >
          {lang.label}
        </button>
      ))}
    </div>
  );
}
