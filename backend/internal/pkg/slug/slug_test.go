package slug

import "testing"

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Hello World":       "hello-world",
		"  Trim Me  ":       "trim-me",
		"Multiple   Spaces": "multiple-spaces",
		"Special!@#Chars":   "special-chars",
		"Already-slugged":   "already-slugged",
		"Ünïcödé Title":     "n-c-d-title",
	}

	for input, want := range cases {
		if got := Make(input); got != want {
			t.Errorf("Make(%q) = %q, want %q", input, got, want)
		}
	}
}
