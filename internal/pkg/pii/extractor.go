package pii

import (
	"context"
	"fmt"

	piiextractor "github.com/intMeric/pii-extractor"
)

type SimpleExtractor struct {
	// The library uses a direct interface, no need to store the extractor
}

func NewExtractor() (Extractor, error) {
	return &SimpleExtractor{}, nil
}

func (se *SimpleExtractor) ExtractPII(ctx context.Context, text string) (*Result, error) {
	// Create a new RegexExtractor instance using the modern API as shown in the example
	extractor := piiextractor.NewDefaultRegexExtractor()

	result, err := extractor.Extract(text)
	if err != nil {
		return nil, fmt.Errorf("failed to extract PII: %w", err)
	}

	return se.convertResult(result), nil
}

func (se *SimpleExtractor) convertResult(result *piiextractor.PiiExtractionResult) *Result {
	entities := make([]Entity, 0, len(result.Entities))

	for _, entity := range result.Entities {
		entities = append(entities, Entity{
			Type:     se.convertPIIType(entity.Type),
			Value:    entity.GetValue(),
			Count:    entity.GetCount(),
			Contexts: entity.GetContexts(),
		})
	}

	// Convert stats map from PiiType to string
	stats := make(map[string]int)
	for piiType, count := range result.Stats {
		stats[se.piiTypeToString(piiType)] = count
	}

	return &Result{
		Total:    result.Total,
		Entities: entities,
		Stats:    stats,
	}
}

func (se *SimpleExtractor) convertPIIType(piiType piiextractor.PiiType) PIIType {
	switch piiType {
	case piiextractor.PiiTypeEmail:
		return PIITypeEmail
	case piiextractor.PiiTypePhone:
		return PIITypePhone
	case piiextractor.PiiTypeCreditCard:
		return PIITypeCreditCard
	case piiextractor.PiiTypeSSN:
		return PIITypeSSN
	case piiextractor.PiiTypeIPAddress:
		return PIITypeIPAddress
	case piiextractor.PiiTypeIBAN:
		return PIITypeIBAN
	default:
		return PIITypeOther
	}
}

func (se *SimpleExtractor) piiTypeToString(piiType piiextractor.PiiType) string {
	switch piiType {
	case piiextractor.PiiTypeEmail:
		return "email"
	case piiextractor.PiiTypePhone:
		return "phone"
	case piiextractor.PiiTypeCreditCard:
		return "credit_card"
	case piiextractor.PiiTypeSSN:
		return "ssn"
	case piiextractor.PiiTypeIPAddress:
		return "ip_address"
	case piiextractor.PiiTypeIBAN:
		return "iban"
	default:
		return "other"
	}
}

func (se *SimpleExtractor) Close() error {
	return nil
}
