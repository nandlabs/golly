package genai

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

// PromptTemplate represents a prompt template with an ID, name, template string, and a parsed Go text/template.
// The template string can contain placeholders in the form {{ParameterName}} which will be replaced
// with the corresponding values from the parameters map during formatting.

type PromptTemplate struct {
	Id             string             `json:"id" yaml:"id" toml:"id"`
	Name           string             `json:"name" yaml:"name" toml:"name"`
	Version        string             `json:"version" yaml:"version" toml:"version"`
	Template       string             `json:"template" yaml:"template" toml:"template"`
	parsedTemplate *template.Template `json:"-" yaml:"–" toml:"-"`
	CreatedAt      int64              `json:"created_at" yaml:"created_at" toml:"created_at"`
	UpdatedAt      int64              `json:"updated_at" yaml:"updated_at" toml:"updated_at"`
}

// PromptStore defines the interface for managing PromptTemplates.

type PromptStore interface {
	// Get retrieves a PromptTemplate by its ID.
	// It returns the PromptTemplate and a boolean indicating whether it was found.
	Get(id string) (*PromptTemplate, bool)
	// Add adds a new PromptTemplate to the store.
	// If a template with the same ID already exists, it returns an error.
	Add(pt *PromptTemplate) error
	// Update updates an existing PromptTemplate in the store.
	// It returns an error if the template does not exist.
	Update(pt *PromptTemplate) error
	// List lists all PromptTemplates in the store.
	List() []*PromptTemplate
	// Remove removes a PromptTemplate from the store by its ID.
	// It returns an error if the template does not exist.
	Remove(id string) error
}

// Format formats the prompt template using the provided parameters map.
// It returns the formatted string or an error if formatting fails.

func (pt *PromptTemplate) Format(params map[string]any) (string, error) {
	var sb strings.Builder
	err := pt.parsedTemplate.Execute(&sb, params)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// NewPromptTemplate creates a new PromptTemplate with the given id, name, template string, and parameters.
// It parses the template string into a Go text/template.Template for later formatting.
// If the template string is invalid, it returns an error.
// The template string can contain placeholders in the form {{ParameterName}} which will be replaced with
// the corresponding values from the parameters map during formatting.If the template string is already in Go
// template format (i.e., contains {{.ParamName}}), it will be used as is else it will be converted to Go template format.
//

func NewPromptTemplate(id, name, tmplate string) (pt *PromptTemplate, err error) {
	tmpl := convertToGoTemplate(tmplate)
	parsedTmpl, e := template.New(name).Parse(tmpl)
	if e != nil {
		err = fmt.Errorf("failed to parse template id '%s': %w", id, e)
		return
	}
	pt = &PromptTemplate{
		Id:             id,
		Name:           name,
		Template:       tmplate,
		parsedTemplate: parsedTmpl,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
	return
}

// convertToGoTemplate parses the template string and returns a representation of Go Text templates.
// The input template string can contain placeholders in the form {{ParameterName}}
// which will be replaced with the corresponding representation in go templates.
// If the first parameter is in the form of {{.ParamName}}, it is assumed to be a Go template already and returned as is.
func convertToGoTemplate(tmpl string) string {
	// First check if the template is already in Go template format. If it contains {{. it is assumed to be a Go template.
	if strings.Contains(tmpl, "{{.") {
		return tmpl
	}
	// if not convert all {{ParamName}} to {{.ParamName}}
	var sb strings.Builder
	i := 0
	for {
		start := strings.Index(tmpl[i:], "{{")
		if start == -1 {
			sb.WriteString(tmpl[i:])
			break
		}
		start += i
		sb.WriteString(tmpl[i:start])
		end := strings.Index(tmpl[start:], "}}")
		if end == -1 {
			// no closing braces, treat as literal
			sb.WriteString(tmpl[start:])
			break
		}
		end += start
		paramName := strings.TrimSpace(tmpl[start+2 : end])
		sb.WriteString("{{.")
		sb.WriteString(paramName)
		sb.WriteString("}}")
		i = end + 2
	}
	return sb.String()
}

// InMemoryPromptStore is an in-memory implementation of the PromptStore interface.
// It uses a map to store PromptTemplates by their ID.
type InMemoryPromptStore struct {
	templates map[string]*PromptTemplate
}

func NewInMemoryPromptStore() *InMemoryPromptStore {
	return &InMemoryPromptStore{
		templates: make(map[string]*PromptTemplate),
	}
}
func (ps *InMemoryPromptStore) Get(id string) (*PromptTemplate, bool) {
	pt, ok := ps.templates[id]
	return pt, ok
}

func (ps *InMemoryPromptStore) Add(pt *PromptTemplate) error {
	if _, exists := ps.templates[pt.Id]; exists {
		return fmt.Errorf("prompt template with id '%s' already exists", pt.Id)
	}
	ps.templates[pt.Id] = pt
	return nil
}

func (ps *InMemoryPromptStore) Update(pt *PromptTemplate) error {
	if _, exists := ps.templates[pt.Id]; !exists {
		return fmt.Errorf("prompt template with id '%s' does not exist", pt.Id)
	}
	pt.UpdatedAt = time.Now().Unix()
	ps.templates[pt.Id] = pt
	return nil
}

func (ps *InMemoryPromptStore) List() []*PromptTemplate {
	var pts []*PromptTemplate
	for _, pt := range ps.templates {
		pts = append(pts, pt)
	}
	return pts
}

func (ps *InMemoryPromptStore) Remove(id string) error {
	if _, exists := ps.templates[id]; !exists {
		return fmt.Errorf("prompt template with id '%s' does not exist", id)
	}
	delete(ps.templates, id)
	return nil
}

func CreatePrompt(store PromptStore, id, name, template string) (*PromptTemplate, error) {
	pt, err := NewPromptTemplate(id, name, template)
	if err != nil {
		return nil, err
	}
	err = store.Add(pt)
	if err != nil {
		return nil, err
	}
	return pt, nil
}
