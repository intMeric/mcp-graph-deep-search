package keyword

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jdkato/prose/v2"
)

type ProseExtractor struct {
	initialized bool
}

func NewProseExtractor() (*ProseExtractor, error) {
	return &ProseExtractor{
		initialized: true,
	}, nil
}

func (pe *ProseExtractor) ExtractKeywords(ctx context.Context, text string, options *Options) ([]string, error) {
	keywords, err := pe.ExtractKeywordsWithScores(ctx, text, options)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(keywords))
	for i, kw := range keywords {
		result[i] = kw.Text
	}
	return result, nil
}

func (pe *ProseExtractor) ExtractKeywordsWithScores(ctx context.Context, text string, options *Options) ([]Keyword, error) {
	if !pe.initialized {
		return nil, fmt.Errorf("prose extractor not initialized")
	}

	if options == nil {
		options = DefaultOptions()
	}

	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, fmt.Errorf("failed to create prose document: %w", err)
	}

	return pe.extractKeywords(doc, options), nil
}

func (pe *ProseExtractor) extractKeywords(doc *prose.Document, options *Options) []Keyword {
	tokenFreq := make(map[string]int)
	stopWords := pe.getStopWords()

	for _, tok := range doc.Tokens() {
		text := strings.ToLower(strings.TrimSpace(tok.Text))

		// Skip empty tokens, short words, or punctuation
		if text == "" || len(text) < options.MinLength || pe.isPunctuation(text) {
			continue
		}

		if options.RemoveStopWords && pe.isStopWord(text, stopWords) {
			continue
		}

		tokenFreq[text]++
	}

	keywords := make([]Keyword, 0, len(tokenFreq))
	for text, freq := range tokenFreq {
		keyword := Keyword{
			Text:      text,
			Frequency: freq,
			Score:     float64(freq),
		}
		keywords = append(keywords, keyword)
	}

	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Score > keywords[j].Score
	})

	if options.MaxKeywords > 0 && len(keywords) > options.MaxKeywords {
		keywords = keywords[:options.MaxKeywords]
	}

	return keywords
}

func (pe *ProseExtractor) getStopWords() map[string]bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
		"through": true, "during": true, "before": true, "after": true, "above": true,
		"below": true, "between": true, "among": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true, "can": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "me": true,
		"my": true, "myself": true, "we": true, "our": true, "ours": true, "ourselves": true,
		"you": true, "your": true, "yours": true, "yourself": true, "yourselves": true,
		"he": true, "him": true, "his": true, "himself": true, "she": true, "her": true,
		"hers": true, "herself": true, "it": true, "its": true, "itself": true, "they": true,
		"them": true, "their": true, "theirs": true, "themselves": true, "what": true,
		"which": true, "who": true, "whom": true, "whose": true, "where": true, "when": true,
		"why": true, "how": true, "all": true, "any": true, "both": true, "each": true,
		"few": true, "more": true, "most": true, "other": true, "some": true, "such": true,
		"no": true, "nor": true, "not": true, "only": true, "own": true, "same": true,
		"so": true, "than": true, "too": true, "very": true, "just": true, "now": true,
	}
	return stopWords
}

func (pe *ProseExtractor) isStopWord(word string, stopWords map[string]bool) bool {
	return stopWords[strings.ToLower(word)]
}

func (pe *ProseExtractor) isPunctuation(text string) bool {
	punctuation := map[string]bool{
		".": true, ",": true, "!": true, "?": true, ":": true, ";": true,
		"(": true, ")": true, "[": true, "]": true, "{": true, "}": true,
		"'": true, "\"": true, "`": true, "-": true, "_": true, "/": true,
		"\\": true, "|": true, "@": true, "#": true, "$": true, "%": true,
		"^": true, "&": true, "*": true, "+": true, "=": true, "~": true,
		"<": true, ">": true,
	}
	return punctuation[text]
}

func (pe *ProseExtractor) Close() error {
	pe.initialized = false
	return nil
}
