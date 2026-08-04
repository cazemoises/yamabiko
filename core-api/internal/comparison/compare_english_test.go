package comparison_test

import (
	"testing"

	"github.com/yamabiko/core-api/internal/comparison"
)

// Estes testes cobrem as 5 categorias novas da taxonomia fonética de inglês
// (pedido do usuário) — cada uma reflete um exemplo real de
// seed/english_curriculum_seed.json:phonetic_taxonomy_new_categories.

func TestCompareLang_English_DetectsThSubstituicao_HDeletadoAposT(t *testing.T) {
	// "think" -> "tink": omissão do H do dígrafo TH (θ não existe em PT-BR).
	result := comparison.CompareLang("think", "tink", "en-US")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d: %+v", len(result.PhoneticDiff), result.PhoneticDiff)
	}
	entry := result.PhoneticDiff[0]
	if entry.Pattern != comparison.PatternThSubstituicao {
		t.Fatalf("esperava TH_SUBSTITUICAO, veio %v", entry.Pattern)
	}
}

func TestCompareLang_English_DetectsThSubstituicao_TSubstituidoPorD(t *testing.T) {
	// "this" -> "dis": TH inicial virou D (substituição comum de TH sonoro em PT-BR).
	result := comparison.CompareLang("this", "dis", "en-US")

	found := false
	for _, entry := range result.PhoneticDiff {
		if entry.Pattern == comparison.PatternThSubstituicao {
			found = true
		}
	}
	if !found {
		t.Fatalf("esperava pelo menos 1 entrada TH_SUBSTITUICAO, veio %+v", result.PhoneticDiff)
	}
}

func TestCompareLang_English_DetectsVogalFinalAdicionada(t *testing.T) {
	// "hot" -> "hoti": vogal extra depois da consoante final (sílaba aberta forçada).
	result := comparison.CompareLang("hot", "hoti", "en-US")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d: %+v", len(result.PhoneticDiff), result.PhoneticDiff)
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpInsert {
		t.Fatalf("esperava INSERT, veio %v", entry.Op)
	}
	if entry.Pattern != comparison.PatternVogalFinalAdicionada {
		t.Fatalf("esperava VOGAL_FINAL_ADICIONADA, veio %v", entry.Pattern)
	}
}

func TestCompareLang_English_DetectsVogalReduzidaIgnorada(t *testing.T) {
	// "about" pronunciado com O completo (substituição de vogal, schwa ignorado).
	result := comparison.CompareLang("about", "abaut", "en-US")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d: %+v", len(result.PhoneticDiff), result.PhoneticDiff)
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpSubstitute {
		t.Fatalf("esperava SUBSTITUTE, veio %v", entry.Op)
	}
	if entry.Pattern != comparison.PatternVogalReduzidaIgnorada {
		t.Fatalf("esperava VOGAL_REDUZIDA_IGNORADA, veio %v", entry.Pattern)
	}
}

func TestCompareLang_English_DetectsRAmericanoTrocado(t *testing.T) {
	// "car" -> "cah": R retroflexo do inglês substituído/omitido.
	result := comparison.CompareLang("car", "cah", "en-US")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d: %+v", len(result.PhoneticDiff), result.PhoneticDiff)
	}
	entry := result.PhoneticDiff[0]
	if entry.Pattern != comparison.PatternRAmericanoTrocado {
		t.Fatalf("esperava R_AMERICANO_TROCADO, veio %v", entry.Pattern)
	}
}

func TestCompareLang_English_DetectsConsoanteFinalOmitida(t *testing.T) {
	// "stop" -> "sto": P final (oclusiva) omitido.
	result := comparison.CompareLang("stop", "sto", "en-US")

	if len(result.PhoneticDiff) != 1 {
		t.Fatalf("esperava 1 entrada no diff, veio %d: %+v", len(result.PhoneticDiff), result.PhoneticDiff)
	}
	entry := result.PhoneticDiff[0]
	if entry.Op != comparison.OpDelete {
		t.Fatalf("esperava DELETE, veio %v", entry.Op)
	}
	if entry.Pattern != comparison.PatternConsoanteFinalOmitida {
		t.Fatalf("esperava CONSOANTE_FINAL_OMITIDA, veio %v", entry.Pattern)
	}
}

func TestCompareLang_English_ConsoanteFinalOmitida_NoMeioDaFrase(t *testing.T) {
	// Mesmo padrão, mas a palavra não está no fim da frase — o limite é a
	// fronteira de palavra (espaço), não o fim da string inteira.
	result := comparison.CompareLang("good morning", "goo morning", "en-US")

	found := false
	for _, entry := range result.PhoneticDiff {
		if entry.Pattern == comparison.PatternConsoanteFinalOmitida {
			found = true
		}
	}
	if !found {
		t.Fatalf("esperava CONSOANTE_FINAL_OMITIDA, veio %+v", result.PhoneticDiff)
	}
}

func TestCompareLang_English_PreservaEspacosEntrePalavras(t *testing.T) {
	// Diferente do japonês (moraico, sem espaços), a normalização de inglês não
	// pode remover espaços — eles definem fronteira de palavra, usada pelos
	// classificadores de VOGAL_FINAL_ADICIONADA/CONSOANTE_FINAL_OMITIDA.
	result := comparison.CompareLang("how are you", "how are you", "en-US")

	if result.SimilarityScore != 1.0 {
		t.Fatalf("esperava score 1.0, veio %v", result.SimilarityScore)
	}
}

func TestCompareLang_JapaneseDefault_UnaffectedByEnglishPatterns(t *testing.T) {
	// Garantia de não regressão: CompareLang com language="ja-JP" deve se
	// comportar exatamente como o Compare() original.
	result := comparison.CompareLang("はい", "い", "ja-JP")

	if len(result.PhoneticDiff) != 1 || result.PhoneticDiff[0].Pattern != comparison.PatternHAspiradoOmitido {
		t.Fatalf("esperava H_ASPIRADO_OMITIDO como no Compare() original, veio %+v", result.PhoneticDiff)
	}
}
