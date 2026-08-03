// Package comparison é o núcleo de valor do produto: compara a transcrição do
// Whisper com o gabarito esperado e classifica divergências em padrões de erro
// fonético conhecidos de falantes de PT-BR (Sec. 3 do CLAUDE.md).
package comparison

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

type Verdict string

const (
	VerdictPass    Verdict = "PASS"
	VerdictPartial Verdict = "PARTIAL"
	VerdictFail    Verdict = "FAIL"
)

type ErrorPattern string

const (
	PatternHAspiradoOmitido ErrorPattern = "H_ASPIRADO_OMITIDO"
	PatternVogalEngolida    ErrorPattern = "VOGAL_ENGOLIDA"
	PatternRLTConfusao      ErrorPattern = "R_L_T_CONFUSAO"
	PatternOutro            ErrorPattern = "OUTRO"
)

type DiffOp string

const (
	OpSubstitute DiffOp = "SUBSTITUTE"
	OpDelete     DiffOp = "DELETE" // presente no esperado, ausente na transcrição
	OpInsert     DiffOp = "INSERT" // presente na transcrição, ausente no esperado
)

type DiffEntry struct {
	Op       DiffOp       `json:"op"`
	Position int          `json:"position"`
	Expected string       `json:"expected,omitempty"`
	Actual   string       `json:"actual,omitempty"`
	Pattern  ErrorPattern `json:"pattern,omitempty"`
}

type Result struct {
	SimilarityScore float64     `json:"similarity_score"`
	Verdict         Verdict     `json:"verdict"`
	PhoneticDiff    []DiffEntry `json:"phonetic_diff"`
}

// hRow são as moras do は行 — falantes de PT-BR tendem a omiti-las porque o "h"
// é mudo em português.
var hRow = map[rune]bool{'は': true, 'ひ': true, 'ふ': true, 'へ': true, 'ほ': true}

// pureVowels são as moras vocálicas puras, alvo comum de "engolir" a vogal.
var pureVowels = map[rune]bool{'あ': true, 'い': true, 'う': true, 'え': true, 'お': true}

// rltConfusable agrupa o ら行 com た/だ行: o flap japonês de ら é frequentemente
// percebido/pronunciado por falantes de PT-BR como um L, T ou D.
var rltConfusable = map[rune]bool{
	'ら': true, 'り': true, 'る': true, 'れ': true, 'ろ': true,
	'た': true, 'ち': true, 'つ': true, 'て': true, 'と': true,
	'だ': true, 'ぢ': true, 'づ': true, 'で': true, 'ど': true,
}

// Compare normaliza as duas strings, calcula a distância de Levenshtein a nível
// de rune (mora) e classifica cada divergência num padrão fonético conhecido.
func Compare(expectedRaw, actualRaw string) Result {
	expected := []rune(normalize(expectedRaw))
	actual := []rune(normalize(actualRaw))

	distance, ops := levenshteinOps(expected, actual)

	maxLen := max(len(expected), len(actual))

	score := 1.0
	if maxLen > 0 {
		score = 1 - float64(distance)/float64(maxLen)
	}

	diff := make([]DiffEntry, 0, len(ops))
	position := 0
	for _, o := range ops {
		if o.op == opMatch {
			position++
			continue
		}

		entry := DiffEntry{Op: DiffOp(o.op), Position: position}
		switch o.op {
		case opSubstitute:
			entry.Expected = string(o.expected)
			entry.Actual = string(o.actual)
			entry.Pattern = classifySubstitute(o.expected, o.actual)
		case opDelete:
			entry.Expected = string(o.expected)
			entry.Pattern = classifyDelete(o.expected)
		case opInsert:
			entry.Actual = string(o.actual)
			entry.Pattern = PatternOutro
		}
		diff = append(diff, entry)
		position++
	}

	return Result{
		SimilarityScore: score,
		Verdict:         verdictFor(score),
		PhoneticDiff:    diff,
	}
}

func verdictFor(score float64) Verdict {
	switch {
	case score >= 0.85:
		return VerdictPass
	case score >= 0.6:
		return VerdictPartial
	default:
		return VerdictFail
	}
}

func classifyDelete(expected rune) ErrorPattern {
	switch {
	case hRow[expected]:
		return PatternHAspiradoOmitido
	case pureVowels[expected]:
		return PatternVogalEngolida
	default:
		return PatternOutro
	}
}

func classifySubstitute(expected, actual rune) ErrorPattern {
	if rltConfusable[expected] && rltConfusable[actual] {
		return PatternRLTConfusao
	}
	return PatternOutro
}

// normalize remove todo espaço em branco, aplica NFKC (unifica formas de
// compatibilidade Unicode, ex: katakana de meia-largura -> largura normal) e
// converte katakana pra hiragana — o Whisper às vezes transcreve num script
// diferente do gabarito pra sons foneticamente idênticos (achado real da Fase
// 4/6, ver BUILD_STATE.md), o que não deveria contar como erro de pronúncia.
func normalize(s string) string {
	s = norm.NFKC.String(s)
	s = strings.Join(strings.Fields(s), "")
	return katakanaToHiragana(s)
}

// katakanaToHiragana converte cada rune do bloco katakana (U+30A1-U+30F6) pro
// hiragana correspondente (U+3041-U+3096) via deslocamento fixo de 0x60 — os
// dois blocos são paralelos no Unicode. ー (prolongador de som, U+30FC) fica
// de fora por não ter equivalente direto em hiragana.
func katakanaToHiragana(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'ァ' && r <= 'ヶ' {
			runes[i] = r - 0x60
		}
	}
	return string(runes)
}
