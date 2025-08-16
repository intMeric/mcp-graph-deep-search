package text_analysis

import (
	"context"
	"strings"

	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/pii"
	"mgds/internal/pkg/serializer"
)

type textAnalysisService struct {
	piiExtractor     pii.Extractor
	keywordExtractor keyword.Extractor
	serializer       serializer.HTMLSerializer
}

func NewTextAnalysisService(piiExtractor pii.Extractor, keywordExtractor keyword.Extractor) TextAnalysisService {
	return &textAnalysisService{
		piiExtractor:     piiExtractor,
		keywordExtractor: keywordExtractor,
		serializer:       serializer.NewGoquerySerializer(nil),
	}
}

func (s *textAnalysisService) AnalyzeText(ctx context.Context, text string) (*TextAnalysisResult, error) {
	result := &TextAnalysisResult{
		Text: text,
	}

	// Serialize HTML content to extract plain text if needed
	var plainText string
	if isHTML(text) {
		doc, err := s.serializer.ParseHTML(ctx, text)
		if err != nil {
			return nil, err
		}
		plainText = doc.Text
	} else {
		plainText = text
	}

	piiResult, err := s.piiExtractor.ExtractPII(ctx, plainText)
	if err != nil {
		return nil, err
	}
	result.PIIResult = piiResult

	keywords, err := s.keywordExtractor.ExtractKeywordsWithScores(ctx, plainText, keyword.DefaultOptions())
	if err != nil {
		return nil, err
	}
	result.Keywords = keywords

	return result, nil
}

func (s *textAnalysisService) Close() error {
	if err := s.piiExtractor.Close(); err != nil {
		return err
	}
	if err := s.keywordExtractor.Close(); err != nil {
		return err
	}
	if err := s.serializer.Close(); err != nil {
		return err
	}
	return nil
}

// isHTML checks if the text appears to be HTML content
func isHTML(text string) bool {
	text = strings.TrimSpace(text)
	return strings.HasPrefix(text, "<") && strings.Contains(text, ">")
}
