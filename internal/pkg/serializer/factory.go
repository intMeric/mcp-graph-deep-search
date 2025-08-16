package serializer

// SerializerType represents the type of HTML serializer
type SerializerType string

const (
	// GoquerySerializer uses the goquery library for HTML parsing
	GoquerySerializer SerializerType = "goquery"
)

// NewHTMLSerializer creates a new HTML serializer based on the specified type
func NewHTMLSerializer(serializerType SerializerType, options *DocumentOptions) HTMLSerializer {
	switch serializerType {
	case GoquerySerializer:
		return NewGoquerySerializer(options)
	default:
		// Default to goquery serializer
		return NewGoquerySerializer(options)
	}
}
