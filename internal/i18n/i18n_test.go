package i18n

import "testing"

func TestNormalizeAndMessages(t *testing.T) {
	for _, testCase := range []struct {
		input string
		want  Language
	}{
		{"pt", Portuguese},
		{"Português", Portuguese},
		{"en-US", English},
		{"English", English},
	} {
		if got := Normalize(testCase.input); got != testCase.want {
			t.Fatalf("Normalize(%q) = %q, want %q", testCase.input, got, testCase.want)
		}
	}
	if !IsValid("en") || IsValid("es") {
		t.Fatal("validação de idioma incorreta")
	}
	if got := T(English, "sync_fetch", "main"); got != "⬇️ Checking updates from origin/main..." {
		t.Fatalf("tradução: %q", got)
	}
}
