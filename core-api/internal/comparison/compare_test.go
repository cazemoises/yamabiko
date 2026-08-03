package comparison_test

import (
	"strings"
	"testing"

	"github.com/yamabiko/core-api/internal/comparison"
)

func TestCompare_ExactMatch_ReturnsPassWithEmptyDiff(t *testing.T) {
	result := comparison.Compare("こんにちは", "こんにちは")

	if result.SimilarityScore != 1.0 {
		t.Fatalf("esperava score 1.0, veio %v", result.SimilarityScore)
	}
	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS, veio %v", result.Verdict)
	}
	if len(result.PhoneticDiff) != 0 {
		t.Fatalf("esperava diff vazio, veio %v", result.PhoneticDiff)
	}
}

func TestCompare_NormalizesWhitespace(t *testing.T) {
	result := comparison.Compare("こんにちは", " こん に ちは ")

	if result.SimilarityScore != 1.0 {
		t.Fatalf("esperava score 1.0 após normalizar espaços, veio %v", result.SimilarityScore)
	}
	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS, veio %v", result.Verdict)
	}
}

func TestCompare_NormalizesHalfWidthKana(t *testing.T) {
	// "ｱﾘｶﾞﾄｳ" em katakana meia-largura deve normalizar (NFKC) pra "アリガトウ".
	result := comparison.Compare("アリガトウ", "ｱﾘｶﾞﾄｳ")

	if result.SimilarityScore != 1.0 {
		t.Fatalf("esperava score 1.0 após normalizar kana meia-largura, veio %v", result.SimilarityScore)
	}
}

func TestCompare_DetectsHAspiradoOmitido(t *testing.T) {
	// "はい" -> "い": omissão do は (aspirado), erro clássico de falante de PT-BR.
	result := comparison.Compare("はい", "い")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d", len(result.PhoneticDiff))
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpDelete {
		t.Fatalf("esperava DELETE, veio %v", entry.Op)
	}
	if entry.Expected != "は" {
		t.Fatalf("esperava expected 'は', veio %q", entry.Expected)
	}
	if entry.Pattern != comparison.PatternHAspiradoOmitido {
		t.Fatalf("esperava H_ASPIRADO_OMITIDO, veio %v", entry.Pattern)
	}
}

func TestCompare_DetectsVogalEngolida(t *testing.T) {
	// "おはよう" -> "おはよ": omissão da vogal final う.
	result := comparison.Compare("おはよう", "おはよ")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d", len(result.PhoneticDiff))
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpDelete {
		t.Fatalf("esperava DELETE, veio %v", entry.Op)
	}
	if entry.Expected != "う" {
		t.Fatalf("esperava expected 'う', veio %q", entry.Expected)
	}
	if entry.Pattern != comparison.PatternVogalEngolida {
		t.Fatalf("esperava VOGAL_ENGOLIDA, veio %v", entry.Pattern)
	}
}

func TestCompare_DetectsRLTConfusao(t *testing.T) {
	// "から" -> "かた": ら confundido com た, confusão típica R/L/T.
	result := comparison.Compare("から", "かた")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d", len(result.PhoneticDiff))
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpSubstitute {
		t.Fatalf("esperava SUBSTITUTE, veio %v", entry.Op)
	}
	if entry.Expected != "ら" || entry.Actual != "た" {
		t.Fatalf("esperava ら->た, veio %q->%q", entry.Expected, entry.Actual)
	}
	if entry.Pattern != comparison.PatternRLTConfusao {
		t.Fatalf("esperava R_L_T_CONFUSAO, veio %v", entry.Pattern)
	}
}

func TestCompare_UnrecognizedSubstitution_MarksOutro(t *testing.T) {
	// "ねこ" -> "ねき": substituição não coberta pelos padrões conhecidos.
	result := comparison.Compare("ねこ", "ねき")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d", len(result.PhoneticDiff))
	}
	entry := result.PhoneticDiff[0]
	if entry.Pattern != comparison.PatternOutro {
		t.Fatalf("esperava OUTRO, veio %v", entry.Pattern)
	}
}

func TestCompare_ExtraInsertedMora_MarksOutro(t *testing.T) {
	// "ねこ" -> "ねこあ": mora extra inserida (alucinação do Whisper), sem padrão fonético definido.
	result := comparison.Compare("ねこ", "ねこあ")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d", len(result.PhoneticDiff))
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpInsert {
		t.Fatalf("esperava INSERT, veio %v", entry.Op)
	}
	if entry.Actual != "あ" {
		t.Fatalf("esperava actual 'あ', veio %q", entry.Actual)
	}
	if entry.Pattern != comparison.PatternOutro {
		t.Fatalf("esperava OUTRO, veio %v", entry.Pattern)
	}
}

func TestCompare_VerdictPass_AtThreshold(t *testing.T) {
	expected := strings.Repeat("あ", 20)
	actual := "ねねね" + strings.Repeat("あ", 17) // 3 substituições em 20 -> score 0.85

	result := comparison.Compare(expected, actual)

	if result.SimilarityScore != 0.85 {
		t.Fatalf("esperava score 0.85, veio %v", result.SimilarityScore)
	}
	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS no limiar 0.85, veio %v", result.Verdict)
	}
}

func TestCompare_VerdictPartial_AtLowerThreshold(t *testing.T) {
	expected := strings.Repeat("あ", 10)
	actual := "ねねねね" + strings.Repeat("あ", 6) // 4 substituições em 10 -> score 0.6

	result := comparison.Compare(expected, actual)

	if result.SimilarityScore != 0.6 {
		t.Fatalf("esperava score 0.6, veio %v", result.SimilarityScore)
	}
	if result.Verdict != comparison.VerdictPartial {
		t.Fatalf("esperava PARTIAL no limiar 0.6, veio %v", result.Verdict)
	}
}

func TestCompare_VerdictFail_BelowLowerThreshold(t *testing.T) {
	result := comparison.Compare("から", "かた") // score 0.5

	if result.Verdict != comparison.VerdictFail {
		t.Fatalf("esperava FAIL, veio %v", result.Verdict)
	}
}

func TestCompare_BothEmpty_ReturnsPassWithFullScore(t *testing.T) {
	result := comparison.Compare("", "")

	if result.SimilarityScore != 1.0 {
		t.Fatalf("esperava score 1.0 pra dois vazios, veio %v", result.SimilarityScore)
	}
	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS, veio %v", result.Verdict)
	}
}

func TestCompare_EmptyActual_ReturnsFail(t *testing.T) {
	result := comparison.Compare("こんにちは", "")

	if result.SimilarityScore != 0.0 {
		t.Fatalf("esperava score 0.0, veio %v", result.SimilarityScore)
	}
	if result.Verdict != comparison.VerdictFail {
		t.Fatalf("esperava FAIL, veio %v", result.Verdict)
	}
}
