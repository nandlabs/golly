package data

// Schema used to define structure of the data
// This is a subset of OpenAPI 3.0 schema
type Schema struct {
	Id               string             `json:"id,omitempty" yaml:"id,omitempty"`
	Ref              string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Schema           string             `json:"-" yaml:"-"`
	Description      string             `json:"description,omitempty" yaml:"description,omitempty"`
	Type             string             `json:"type,omitempty" yaml:"type,omitempty"`
	Nullable         bool               `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Format           *string            `json:"format,omitempty" yaml:"format,omitempty"`
	Title            string             `json:"title,omitempty" yaml:"title,omitempty"`
	Default          any                `json:"default,omitempty" yaml:"default,omitempty"`
	Maximum          *float64           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum *float64           `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          *float64           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum *float64           `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength        *int               `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength        *int               `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern          *string            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems         *int               `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems         *int               `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems      bool               `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MultipleOf       *float64           `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Enum             []any              `json:"enum,omitempty" yaml:"enum,omitempty"`
	MaxProperties    *int               `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties    *int               `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required         []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Items            *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	AllOf            []*Schema          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf            []*Schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf            []*Schema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not              *Schema            `json:"not,omitempty" yaml:"not,omitempty"`
	Properties       map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalItems  *Schema            `json:"additionalItems,omitempty" yaml:"additionalItems,omitempty"`
	Xml              *Xml               `json:"xml,omitempty" yaml:"xml,omitempty"`
	ReadOnly         bool               `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly        bool               `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Example          any                `json:"example,omitempty" yaml:"example,omitempty"`
	Examples         []any              `json:"examples,omitempty" yaml:"examples,omitempty"`
	Deprecated       bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

type Xml struct {
	Name      *string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace *string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    *string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Attribute *bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   *bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}
