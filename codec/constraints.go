package codec

// StructMeta Struct
type StructMeta struct {
	Fields map[string]string
}

// BaseConstraints struct captures the basic information for a field
type BaseConstraints struct {
	// Name of the field
	Name string
	// Dimension holds the field dimension
	Dimension int
	// Required flag indicating if the field is a required field.
	Required bool
	// TargetNames stores map for known format types. This allows
	TargetNames map[string]string
	// TargetConfig stores configuration that is required by the target format . for Eg. Attribute config for XML etc.
	TargetConfig map[string]string
	// Sequence specifies the order  of the fields in the source/target format
	Sequence int
	//SkipField indicates that if the value of the field is absent/nil then skip the field while writing to data
	//This is similar to omitempty
	SkipField bool
}

// StrConstraints Struct
type StrConstraints struct {
	BaseConstraints
	DefaultVal *string
	Pattern    *string
	Format     *string
	MinLength  *int
	MaxLength  *int
}

// IntConstraints Struct
type IntConstraints struct {
	BaseConstraints
	DefaultVal *int
	Min        *int //The value is inclusive
	Max        *int //The value is inclusive
}

// UIntConstraints Struct
type UIntConstraints struct {
	BaseConstraints
	DefaultVal *uint
	Min        *uint //The value is inclusive
	Max        *uint //The value is inclusive
}

// F32Constraints Struct
type F32Constraints struct {
	BaseConstraints
	DefaultVal *float32
	Min        *float32 //The value is inclusive
	Max        *float32 //The value is inclusive
}

// F64Constraints Struct
type F64Constraints struct {
	BaseConstraints
	DefaultVal *float64
	Min        *float64 //The value is inclusive
	Max        *float64 //The value is inclusive
}

// BoolConstraints Struct
type BoolConstraints struct {
	BaseConstraints
	DefaultVal *bool
}
