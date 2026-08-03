package comparison

type editOp string

const (
	opMatch      editOp = "MATCH"
	opSubstitute editOp = "SUBSTITUTE"
	opDelete     editOp = "DELETE"
	opInsert     editOp = "INSERT"
)

type alignmentOp struct {
	op       editOp
	expected rune
	actual   rune
}

// levenshteinOps calcula a distância de edição a nível de rune entre expected e
// actual e reconstrói a sequência de operações (match/substitute/delete/insert)
// que produz essa distância mínima.
func levenshteinOps(expected, actual []rune) (int, []alignmentOp) {
	n, m := len(expected), len(actual)

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 0; i <= n; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if expected[i-1] == actual[j-1] {
				dp[i][j] = dp[i-1][j-1]
				continue
			}
			dp[i][j] = 1 + min3(dp[i-1][j-1], dp[i-1][j], dp[i][j-1])
		}
	}

	ops := make([]alignmentOp, 0, n+m)
	i, j := n, m
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && expected[i-1] == actual[j-1]:
			ops = append(ops, alignmentOp{op: opMatch, expected: expected[i-1], actual: actual[j-1]})
			i--
			j--
		case i > 0 && j > 0 && dp[i][j] == dp[i-1][j-1]+1:
			ops = append(ops, alignmentOp{op: opSubstitute, expected: expected[i-1], actual: actual[j-1]})
			i--
			j--
		case i > 0 && dp[i][j] == dp[i-1][j]+1:
			ops = append(ops, alignmentOp{op: opDelete, expected: expected[i-1]})
			i--
		default:
			ops = append(ops, alignmentOp{op: opInsert, actual: actual[j-1]})
			j--
		}
	}

	for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
		ops[l], ops[r] = ops[r], ops[l]
	}

	return dp[n][m], ops
}

func min3(a, b, c int) int {
	return min(a, min(b, c))
}
