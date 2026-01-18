package handler

import (
	"testing"
)

func TestCleanProfanity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single profane word",
			input:    "This is kerfuffle",
			expected: "This is ****",
		},
		{
			name:     "multiple profane words",
			input:    "What a kerfuffle and sharbert situation",
			expected: "What a **** and **** situation",
		},
		{
			name:     "profane word with punctuation",
			input:    "This is a kerfuffle!",
			expected: "This is a ****!",
		},
		{
			name:     "profane word with mixed case",
			input:    "This is a KERFUFFLE test",
			expected: "This is a **** test",
		},
		{
			name:     "profane word at start",
			input:    "Kerfuffle is happening",
			expected: "**** is happening",
		},
		{
			name:     "profane word at end",
			input:    "What a fornax",
			expected: "What a ****",
		},
		{
			name:     "no profane words",
			input:    "This is a clean message",
			expected: "This is a clean message",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "all profane words",
			input:    "kerfuffle sharbert fornax",
			expected: "**** **** ****",
		},
		{
			name:     "profane word with multiple punctuation",
			input:    "What the kerfuffle?!?",
			expected: "What the ****?!?",
		},
		{
			name:     "profane word in middle of sentence",
			input:    "I think kerfuffle, you know what I mean",
			expected: "I think ****, you know what I mean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanProfanity(tt.input)
			if result != tt.expected {
				t.Errorf("cleanProfanity(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
