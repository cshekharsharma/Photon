package visualanalysis

import "testing"

func TestCompareFacesResultHighestSimilarityField(t *testing.T) {
	result := CompareFacesResult{HighestSimilarity: 92.5}
	if result.HighestSimilarity != 92.5 {
		t.Fatalf("expected highest similarity to be preserved")
	}
}
