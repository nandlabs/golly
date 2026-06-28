package genai

import (
	"testing"
)

// FuzzPromptTemplate_NewAndFormat asserts that constructing a PromptTemplate
// and then formatting it with an arbitrary parameter value never panics.
// The template engine accepts both {placeholder} and Go-template syntaxes;
// either malformed input must error rather than crash.
func FuzzPromptTemplate_NewAndFormat(f *testing.F) {
	type seed struct {
		template string
		key      string
		value    string
	}
	seeds := []seed{
		{"hello {name}", "name", "world"},
		{"{a}+{b}={c}", "a", "1"},
		{"plain text — no placeholders", "x", "y"},
		{"unbalanced { brace", "x", "y"},
		{"{nested {brace}}", "brace", "B"},
		{"", "x", "y"},
		{"{empty}", "empty", ""},
		{"go template {{.Name}}", "Name", "G"},
		{"mixed {hello} {{.Name}}", "hello", "H"},
		{"{a} {a} {a}", "a", "repeat"},
		{"{name}", "name", "\x00\x01\x02"},
		{"{key with space}", "key with space", "v"},
		{"{name", "name", "v"}, // unclosed
		{"name}", "name", "v"}, // unopened
	}
	for _, s := range seeds {
		f.Add(s.template, s.key, s.value)
	}
	f.Fuzz(func(t *testing.T, tmpl, key, val string) {
		pt, err := NewPromptTemplate("fuzz", "fuzz-template", tmpl)
		if err != nil {
			return // malformed template — must error, not panic
		}
		_, _ = pt.Format(map[string]any{key: val})
	})
}
