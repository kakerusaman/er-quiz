package quiz

import "testing"

func TestPublicDoesNotModifyOriginal(t *testing.T) {
	d := SampleQuestions()[0]
	_ = d.Public()
	if d.Entities[1].Name == nil || d.Entities[0].Attributes[2].Name == nil {
		t.Fatal("Public() cleared names in the original data")
	}
}
