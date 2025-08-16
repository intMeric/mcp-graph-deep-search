package text_analysis

import (
	"context"
	"encoding/json"

	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/pii"
)

type TextAnalysisService interface {
	AnalyzeText(ctx context.Context, text string) (*TextAnalysisResult, error)
	Close() error
}

type TextAnalysisResult struct {
	Text      string            `json:"text"`
	PIIResult *pii.Result       `json:"pii_result"`
	Keywords  []keyword.Keyword `json:"keywords"`
}

func (r *TextAnalysisResult) HasPII() bool {
	return r.PIIResult != nil && !r.PIIResult.IsEmpty()
}

func (r *TextAnalysisResult) HasKeywords() bool {
	return len(r.Keywords) > 0
}

func (r *TextAnalysisResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *TextAnalysisResult) ToJSONString() (string, error) {
	data, err := r.ToJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FromJSON(data []byte) (*TextAnalysisResult, error) {
	var result TextAnalysisResult
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func FromJSONString(jsonString string) (*TextAnalysisResult, error) {
	return FromJSON([]byte(jsonString))
}
