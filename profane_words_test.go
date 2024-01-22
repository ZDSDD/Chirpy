package main

import (
	"testing"
)

func TestCleanBody(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "This is a clean chirp.",
			expected: "This is a clean chirp.",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			input:    ".kerfuffle KERFUFFLE KERFuffle kerfuffle kerfuffle kerfuffle! foRnaX",
			expected: ".kerfuffle **** **** **** **** kerfuffle! ****",
		},
	}

	for _, c := range cases {
		actual := cleanBody(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("lengths don't match: '%v' vs '%v'", actual, c.expected)
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%v) == %v, expected %v", c.input, actual, c.expected)
			}
		}
	}
}
