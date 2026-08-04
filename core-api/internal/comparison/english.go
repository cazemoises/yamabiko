package comparison

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func isEnglish(language string) bool {
	return strings.HasPrefix(strings.ToLower(language), "en")
}

// englishVowels são as vogais simples do alfabeto latino usadas pelos
// heurísticos de VOGAL_FINAL_ADICIONADA e VOGAL_REDUZIDA_IGNORADA — 'y' fica
// de fora por ter comportamento ambíguo (consoante em "yes", vogal em "gym").
var englishVowels = map[rune]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true}

func isEnglishConsonant(r rune) bool {
	return r >= 'a' && r <= 'z' && !englishVowels[r]
}

// normalizeEnglish aplica NFKC, lower-case e remove pontuação (mantendo
// apóstrofo, que faz parte de contrações como "I'm"/"don't" e afeta a
// pronúncia) — ao contrário do japonês, os espaços são preservados (só
// colapsados quando repetidos) porque os classificadores de fronteira de
// palavra abaixo dependem deles pra saber onde uma palavra termina.
func normalizeEnglish(s string) string {
	s = norm.NFKC.String(s)
	s = strings.ToLower(s)

	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\'':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteRune(' ')
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func runeAt(s []rune, idx int) (rune, bool) {
	if idx < 0 || idx >= len(s) {
		return 0, false
	}
	return s[idx], true
}

// atWordBoundary é true quando `idx` está no fim da string ou aponta pra um
// espaço — usado pra detectar posição "final de palavra" em `fullExpected`.
func atWordBoundary(fullExpected []rune, idx int) bool {
	r, ok := runeAt(fullExpected, idx)
	return !ok || r == ' '
}

// touchesThDigraph é true quando `r` (o rune de `fullExpected` em `idx`) faz
// parte do dígrafo "th" — cobre tanto o T quanto o H, porque o Levenshtein
// pode escolher deletar/substituir qualquer um dos dois lados dependendo do
// alinhamento de menor distância.
func touchesThDigraph(fullExpected []rune, idx int, r rune) bool {
	switch r {
	case 'h':
		prev, ok := runeAt(fullExpected, idx-1)
		return ok && prev == 't'
	case 't':
		next, ok := runeAt(fullExpected, idx+1)
		return ok && next == 'h'
	default:
		return false
	}
}

// classifyDeleteEN classifica uma mora deletada usando o contexto de
// `fullExpected` ao redor de `idx` (posição do rune deletado).
func classifyDeleteEN(expected rune, fullExpected []rune, idx int) ErrorPattern {
	switch {
	case touchesThDigraph(fullExpected, idx, expected):
		return PatternThSubstituicao
	case expected == 'r':
		return PatternRAmericanoTrocado
	case isEnglishConsonant(expected) && atWordBoundary(fullExpected, idx+1):
		return PatternConsoanteFinalOmitida
	default:
		return PatternOutro
	}
}

func classifySubstituteEN(expected, actual rune, fullExpected []rune, idx int) ErrorPattern {
	switch {
	case touchesThDigraph(fullExpected, idx, expected):
		return PatternThSubstituicao
	case expected == 'r':
		return PatternRAmericanoTrocado
	case englishVowels[expected] && englishVowels[actual]:
		return PatternVogalReduzidaIgnorada
	default:
		return PatternOutro
	}
}

// classifyInsertEN detecta vogal extra inserida logo após uma consoante final
// de palavra (ex: "hot" -> "hoti") — `idx` é a posição em `fullExpected` onde
// a inserção ocorre (insert não avança o ponteiro do expected).
func classifyInsertEN(actual rune, fullExpected []rune, idx int) ErrorPattern {
	prev, hasPrev := runeAt(fullExpected, idx-1)
	if englishVowels[actual] && hasPrev && isEnglishConsonant(prev) && atWordBoundary(fullExpected, idx) {
		return PatternVogalFinalAdicionada
	}
	return PatternOutro
}
