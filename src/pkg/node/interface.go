package node

// Interface defines the contract for graph nodes
type Interface interface {
	// Core required fields
	GetMgdsId() string
	GetDisplayName() string
	GetDescription() string

	// Type identification
	GetType() string

	// Minimal generic properties (for metadata)
	GetProperty(key string) (any, bool)
	SetProperty(key string, value any)

	// Serialization
	Serialize() (map[string]any, error)
	Deserialize(data map[string]any) error

	// Validation
	IsValid() bool
}
