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
	PatternHAspiradoOmitido   ErrorPattern = "H_ASPIRADO_OMITIDO"
	PatternVogalEngolida      ErrorPattern = "VOGAL_ENGOLIDA"
	PatternRLTConfusao        ErrorPattern = "R_L_T_CONFUSAO"
	PatternSonorizacaoConfusa ErrorPattern = "SONORIZACAO_CONFUSA"

	// Taxonomia de erros pra falantes de PT-BR aprendendo inglês (Sec.
	// phonetic_taxonomy_new_categories de seed/english_curriculum_seed.json).
	PatternThSubstituicao        ErrorPattern = "TH_SUBSTITUICAO"
	PatternVogalFinalAdicionada  ErrorPattern = "VOGAL_FINAL_ADICIONADA"
	PatternVogalReduzidaIgnorada ErrorPattern = "VOGAL_REDUZIDA_IGNORADA"
	PatternRAmericanoTrocado     ErrorPattern = "R_AMERICANO_TROCADO"
	PatternConsoanteFinalOmitida ErrorPattern = "CONSOANTE_FINAL_OMITIDA"

	PatternOutro ErrorPattern = "OUTRO"
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

// rRow é o ら行 — o flap japonês de ら é frequentemente percebido/pronunciado
// por falantes de PT-BR como um L, T ou D.
var rRow = map[rune]bool{'ら': true, 'り': true, 'る': true, 'れ': true, 'ろ': true}

// tdRow é o た/だ行 (surdo e sonoro juntos) — o outro lado da confusão com ら.
// Fica separado de rRow porque nem toda substituição dentro de tdRow envolve
// ら (ex: た<->だ é sonorização pura, ver dakutenPairs/isRLTConfusion abaixo).
var tdRow = map[rune]bool{
	'た': true, 'ち': true, 'つ': true, 'て': true, 'と': true,
	'だ': true, 'ぢ': true, 'づ': true, 'で': true, 'ど': true,
}

// dakutenPairs mapeia cada mora surda pra sua contraparte sonorizada (dakuten)
// nos 4 pares fonológicos k/g, s/z, t/d, h/b — a diferença entre elas é só a
// vibração das cordas vocais (sonorização), um padrão de confusão real e
// sistemático pra falantes de PT-BR aprendendo japonês (não ruído aleatório).
var dakutenPairs = map[rune]rune{
	'か': 'が', 'き': 'ぎ', 'く': 'ぐ', 'け': 'げ', 'こ': 'ご',
	'さ': 'ざ', 'し': 'じ', 'す': 'ず', 'せ': 'ぜ', 'そ': 'ぞ',
	'た': 'だ', 'ち': 'ぢ', 'つ': 'づ', 'て': 'で', 'と': 'ど',
	'は': 'ば', 'ひ': 'び', 'ふ': 'ぶ', 'へ': 'べ', 'ほ': 'ぼ',
}

// Compare é o atalho ja-JP-only histórico (equivalente a CompareLang com
// language="ja-JP") — mantido pra não quebrar chamadores existentes.
func Compare(expectedRaw, actualRaw string) Result {
	return CompareLang(expectedRaw, actualRaw, "ja-JP")
}

// CompareLang normaliza as duas strings de acordo com `language`, calcula a
// distância de Levenshtein a nível de rune e classifica cada divergência num
// padrão fonético conhecido pra falantes de PT-BR aprendendo esse idioma.
func CompareLang(expectedRaw, actualRaw, language string) Result {
	expected := []rune(normalize(expectedRaw, language))
	actual := []rune(normalize(actualRaw, language))

	distance, ops := levenshteinOps(expected, actual)

	maxLen := max(len(expected), len(actual))

	score := 1.0
	if maxLen > 0 {
		score = 1 - float64(distance)/float64(maxLen)
	}

	diff := make([]DiffEntry, 0, len(ops))
	position := 0
	// expIdx é o índice em `expected` do rune sendo processado (match,
	// substitute e delete consomem `expected` na ordem original; insert não
	// avança) — dá aos classificadores de inglês contexto de vizinhança
	// (ex: dígrafo "th", fronteira de palavra) que a mora isolada não tem.
	expIdx := 0
	for _, o := range ops {
		if o.op == opMatch {
			position++
			expIdx++
			continue
		}

		entry := DiffEntry{Op: DiffOp(o.op), Position: position}
		switch o.op {
		case opSubstitute:
			entry.Expected = string(o.expected)
			entry.Actual = string(o.actual)
			entry.Pattern = classifySubstitute(o.expected, o.actual, expected, expIdx, language)
			expIdx++
		case opDelete:
			entry.Expected = string(o.expected)
			entry.Pattern = classifyDelete(o.expected, expected, expIdx, language)
			expIdx++
		case opInsert:
			entry.Actual = string(o.actual)
			entry.Pattern = classifyInsert(o.actual, expected, expIdx, language)
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

// classifyDelete/classifySubstitute/classifyInsert despacham pro classificador
// específico do idioma — japonês opera mora a mora (sem precisar de contexto
// vizinho), inglês precisa da vizinhança em `fullExpected`/`idx` (dígrafo TH,
// fronteira de palavra) pra classificar corretamente.
func classifyDelete(expected rune, fullExpected []rune, idx int, language string) ErrorPattern {
	if isEnglish(language) {
		return classifyDeleteEN(expected, fullExpected, idx)
	}
	return classifyDeleteJA(expected)
}

func classifySubstitute(expected, actual rune, fullExpected []rune, idx int, language string) ErrorPattern {
	if isEnglish(language) {
		return classifySubstituteEN(expected, actual, fullExpected, idx)
	}
	return classifySubstituteJA(expected, actual)
}

func classifyInsert(actual rune, fullExpected []rune, idx int, language string) ErrorPattern {
	if isEnglish(language) {
		return classifyInsertEN(actual, fullExpected, idx)
	}
	return PatternOutro
}

func classifyDeleteJA(expected rune) ErrorPattern {
	switch {
	case hRow[expected]:
		return PatternHAspiradoOmitido
	case pureVowels[expected]:
		return PatternVogalEngolida
	default:
		return PatternOutro
	}
}

func classifySubstituteJA(expected, actual rune) ErrorPattern {
	switch {
	case isRLTConfusion(expected, actual):
		return PatternRLTConfusao
	case isDakutenPair(expected, actual):
		return PatternSonorizacaoConfusa
	case pureVowels[expected] && pureVowels[actual]:
		// Confusão entre duas vogais puras (ex: え/い) é o mesmo problema de
		// qualidade vocálica da VOGAL_ENGOLIDA, só que via substituição em vez
		// de omissão completa da mora.
		return PatternVogalEngolida
	default:
		return PatternOutro
	}
}

// isRLTConfusion é true só quando ら (ou outra mora de rRow) está de um dos
// lados — sem isso, た<->だ (nenhum dos dois em rRow) cairia aqui também, o
// que misturaria a confusão com ら com a sonorização pura de isDakutenPair.
func isRLTConfusion(expected, actual rune) bool {
	return (rRow[expected] && (rRow[actual] || tdRow[actual])) ||
		(rRow[actual] && (rRow[expected] || tdRow[expected]))
}

func isDakutenPair(expected, actual rune) bool {
	return dakutenPairs[expected] == actual || dakutenPairs[actual] == expected
}

// normalize despacha pra normalização específica do idioma. Japonês é
// moraico (sem espaços entre "palavras"), então remove todo espaço em branco;
// inglês é alfabético e os classificadores de fronteira de palavra (Sec.
// english.go) dependem dos espaços sobreviverem, então só colapsa espaços
// repetidos em um único.
func normalize(s, language string) string {
	if isEnglish(language) {
		return normalizeEnglish(s)
	}
	return normalizeJapanese(s)
}

// normalizeJapanese aplica NFKC (unifica formas de compatibilidade Unicode,
// ex: katakana de meia-largura -> largura normal), remove todo espaço em
// branco e converte katakana pra hiragana — o Whisper às vezes transcreve num
// script diferente do gabarito pra sons foneticamente idênticos (achado real
// da Fase 4/6, ver BUILD_STATE.md), o que não deveria contar como erro de
// pronúncia.
func normalizeJapanese(s string) string {
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
