package plagiarism

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// TextProcessor handles text preprocessing for plagiarism detection
type TextProcessor struct {
	stopWords map[string]bool
}

// NewTextProcessor creates a new text processor
func NewTextProcessor() *TextProcessor {
	stopWords := map[string]bool{
		"и": true, "в": true, "на": true, "с": true, "по": true, "для": true,
		"от": true, "до": true, "из": true, "к": true, "о": true, "об": true,
		"что": true, "как": true, "так": true, "но": true, "а": true, "или": true,
		"же": true, "бы": true, "ли": true, "не": true, "ни": true, "то": true,
		"это": true, "этот": true, "эта": true, "эти": true, "тот": true, "та": true,
		"те": true, "он": true, "она": true, "оно": true, "они": true, "мы": true,
		"вы": true, "я": true, "ты": true, "его": true, "её": true, "их": true,
		"наш": true, "ваш": true, "мой": true, "твой": true, "свой": true,
		"который": true, "которая": true, "которое": true, "которые": true,
		"где": true, "когда": true, "почему": true, "зачем": true, "куда": true,
		"откуда": true, "сколько": true, "чем": true, "чего": true, "кого": true,
		"кому": true, "кем": true, "чему": true, "чём": true, "при": true,
		"под": true, "над": true, "за": true, "перед": true, "между": true,
		"через": true, "без": true, "против": true, "вместо": true, "кроме": true,
		"после": true, "во": true, "со": true, "ко": true,
	}

	return &TextProcessor{
		stopWords: stopWords,
	}
}

// CleanText removes punctuation, HTML tags, and normalizes text
func (tp *TextProcessor) CleanText(text string) string {
	// Удаляем HTML теги
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	text = htmlRegex.ReplaceAllString(text, " ")

	// Удаляем лишние пробелы и переводы строк
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// Приводим к нижнему регистру
	text = strings.ToLower(text)

	// Удаляем пунктуацию, оставляем только буквы и пробелы
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}

	// Убираем лишние пробелы
	text = spaceRegex.ReplaceAllString(result.String(), " ")
	return strings.TrimSpace(text)
}

// RemoveStopWords removes stop words from text
func (tp *TextProcessor) RemoveStopWords(text string) string {
	words := strings.Fields(text)
	var filteredWords []string

	for _, word := range words {
		if !tp.stopWords[word] && len(word) > 2 { // Также убираем слова короче 3 символов
			filteredWords = append(filteredWords, word)
		}
	}

	return strings.Join(filteredWords, " ")
}

// SimpleStem performs basic stemming for Russian words
func (tp *TextProcessor) SimpleStem(word string) string {
	// Простой стемминг для русского языка
	// Удаляем распространенные окончания
	suffixes := []string{
		"ость", "ение", "ание", "ение", "ние", "ие", "ые", "ий", "ая", "ое",
		"ем", "ам", "ах", "ми", "ов", "ев", "ей", "ой", "ый", "ая", "ее",
		"ет", "ит", "ут", "ют", "ал", "ил", "ел", "ла", "ло", "ли",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(word, suffix) && len(word) > len(suffix)+2 {
			return word[:len(word)-len(suffix)]
		}
	}

	return word
}

// StemText applies stemming to all words in text
func (tp *TextProcessor) StemText(text string) string {
	words := strings.Fields(text)
	var stemmedWords []string

	for _, word := range words {
		stemmed := tp.SimpleStem(word)
		stemmedWords = append(stemmedWords, stemmed)
	}

	return strings.Join(stemmedWords, " ")
}

// ProcessText performs full text preprocessing
func (tp *TextProcessor) ProcessText(text string) string {
	cleaned := tp.CleanText(text)
	withoutStopWords := tp.RemoveStopWords(cleaned)
	stemmed := tp.StemText(withoutStopWords)

	return stemmed
}

// GenerateShingles creates n-grams from processed text
func (tp *TextProcessor) GenerateShingles(text string, n int) []string {
	words := strings.Fields(text)
	if len(words) < n {
		return []string{strings.Join(words, " ")}
	}

	var shingles []string
	for i := 0; i <= len(words)-n; i++ {
		shingle := strings.Join(words[i:i+n], " ")
		shingles = append(shingles, shingle)
	}

	return shingles
}

// HashShingles creates MD5 hashes for shingles
func (tp *TextProcessor) HashShingles(shingles []string) []string {
	var hashes []string
	for _, shingle := range shingles {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(shingle)))
		hashes = append(hashes, hash)
	}
	return hashes
}

// CalculateTextStatistics calculates basic text statistics
func (tp *TextProcessor) CalculateTextStatistics(originalText string) *TextStatistics {
	characterCount := len(originalText)

	words := strings.Fields(originalText)
	wordCount := len(words)

	// Подсчет абзацев (по двойным переводам строк)
	paragraphs := strings.Split(originalText, "\n\n")
	paragraphCount := len(paragraphs)
	if paragraphCount == 1 && strings.TrimSpace(paragraphs[0]) == "" {
		paragraphCount = 0
	}

	// Подсчет предложений (по точкам, восклицательным и вопросительным знакам)
	sentenceRegex := regexp.MustCompile(`[.!?]+`)
	sentences := sentenceRegex.Split(originalText, -1)
	sentenceCount := len(sentences) - 1 // Последний элемент обычно пустой
	if sentenceCount < 0 {
		sentenceCount = 0
	}

	return &TextStatistics{
		ParagraphCount: paragraphCount,
		WordCount:      wordCount,
		CharacterCount: characterCount,
		SentenceCount:  sentenceCount,
	}
}

// TextStatistics represents text analysis statistics
type TextStatistics struct {
	ParagraphCount int `json:"paragraph_count"`
	WordCount      int `json:"word_count"`
	CharacterCount int `json:"character_count"`
	SentenceCount  int `json:"sentence_count"`
}
