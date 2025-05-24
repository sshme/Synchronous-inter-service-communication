package plagiarism

import (
	"testing"
)

func TestTextProcessor_CleanText(t *testing.T) {
	processor := NewTextProcessor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML tags removal",
			input:    "<p>Привет <b>мир</b>!</p>",
			expected: "привет мир",
		},
		{
			name:     "Punctuation removal",
			input:    "Привет, мир! Как дела?",
			expected: "привет мир как дела",
		},
		{
			name:     "Multiple spaces",
			input:    "Привет    мир   !",
			expected: "привет мир",
		},
		{
			name:     "Mixed case",
			input:    "ПрИвЕт МиР",
			expected: "привет мир",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.CleanText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanText() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTextProcessor_RemoveStopWords(t *testing.T) {
	processor := NewTextProcessor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove common stop words",
			input:    "это быстрая коричневая лиса",
			expected: "быстрая коричневая лиса",
		},
		{
			name:     "Remove prepositions",
			input:    "книга на столе в комнате",
			expected: "книга столе комнате",
		},
		{
			name:     "Keep meaningful words",
			input:    "программирование алгоритмы структуры данных",
			expected: "программирование алгоритмы структуры данных",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.RemoveStopWords(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveStopWords() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTextProcessor_GenerateShingles(t *testing.T) {
	processor := NewTextProcessor()

	tests := []struct {
		name     string
		input    string
		n        int
		expected []string
	}{
		{
			name:  "4-grams from sentence",
			input: "быстрая коричневая лиса прыгает через забор",
			n:     4,
			expected: []string{
				"быстрая коричневая лиса прыгает",
				"коричневая лиса прыгает через",
				"лиса прыгает через забор",
			},
		},
		{
			name:     "3-grams from short text",
			input:    "один два три",
			n:        3,
			expected: []string{"один два три"},
		},
		{
			name:     "Text shorter than n",
			input:    "один два",
			n:        4,
			expected: []string{"один два"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.GenerateShingles(tt.input, tt.n)
			if len(result) != len(tt.expected) {
				t.Errorf("GenerateShingles() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, shingle := range result {
				if shingle != tt.expected[i] {
					t.Errorf("GenerateShingles()[%d] = %v, want %v", i, shingle, tt.expected[i])
				}
			}
		})
	}
}

func TestTextProcessor_CalculateTextStatistics(t *testing.T) {
	processor := NewTextProcessor()

	text := `Первый абзац. Это второе предложение!

Второй абзац. Здесь предложение.

Третий абзац.`

	stats := processor.CalculateTextStatistics(text)

	if stats.ParagraphCount != 3 {
		t.Errorf("ParagraphCount = %v, want %v", stats.ParagraphCount, 3)
	}

	if stats.WordCount != 11 {
		t.Errorf("WordCount = %v, want %v", stats.WordCount, 21)
	}

	if stats.SentenceCount != 5 {
		t.Errorf("SentenceCount = %v, want %v", stats.SentenceCount, 6)
	}

	if stats.CharacterCount != len(text) {
		t.Errorf("CharacterCount = %v, want %v", stats.CharacterCount, len(text))
	}
}

func TestTextProcessor_ProcessText(t *testing.T) {
	processor := NewTextProcessor()

	input := "<p>Это <b>быстрая</b> коричневая лиса, которая прыгает через забор!</p>"
	result := processor.ProcessText(input)

	if result == "" {
		t.Error("ProcessText() returned empty string")
	}

	if contains(result, "<") || contains(result, ">") {
		t.Error("ProcessText() did not remove HTML tags")
	}

	if result != toLower(result) {
		t.Error("ProcessText() did not convert to lowercase")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	runes := []rune(s)
	result := make([]rune, len(runes))
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else if r >= 'А' && r <= 'Я' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}
