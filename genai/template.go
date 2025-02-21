package genai

import (
	"io"
	"strings"
	"sync"
	"text/template"
)

// TemplateType is the type of the template
type TemplateType string

// templateCache is a cache of templates
var templateCache map[string]PromptTemplate

// cacheMutex is a mutex for the cache. This is preferred over using a sync.Map because it is faster in most cases
var cacheMutex sync.RWMutex

const (
	// GoTextTemplate represents a Go text template
	GoTextTemplate TemplateType = "go-text"
)

// PromptTemplate is the interface that represents a prompt template
type PromptTemplate interface {
	// Id returns the id of the template. This is expected to be unique.
	Id() string
	// Type returns the type of the template
	Type() TemplateType
	// FormatAsText formats the template as text
	FormatAsText(map[string]any) (string, error)
	//WriteTo writes the template to a writer
	WriteTo(io.Writer, map[string]any) error
}

// goTemplate is a template that uses the Go template format
type goTemplate struct {
	//template is the Go template string
	template *template.Template
	//id is the id of the template
	id string
}

func (t *goTemplate) Id() string {
	return t.id
}

// Type returns the format type of the template.
func (t *goTemplate) Type() TemplateType {
	return GoTextTemplate
}

// FormatAsText returns the string format of the template.
func (t *goTemplate) FormatAsText(data map[string]any) (s string, err error) {
	sb := new(strings.Builder)
	err = t.WriteTo(sb, data)
	if err == nil {
		s = sb.String()
	}
	return
}

// WriteTo writes the template to the writer
func (t *goTemplate) WriteTo(w io.Writer, data map[string]any) error {
	//prepare the data
	d := prepareData(data)
	//execute the template
	return t.template.Execute(w, d)
}

// prepareData prepares the data for the template
// the input data is a map and the value can be a function
func prepareData(data map[string]any) map[string]any {
	d := make(map[string]any)
	//for each key-value pair in the data
	for k, v := range data {
		//if the value is a function
		if f, ok := v.(func() any); ok {
			//set the value to the result of the function
			d[k] = f()
		} else {
			//set the value to the original value
			d[k] = v
		}
	}
	//return the data
	return d
}

// GetPromptTemplate returns a prompt template from the cache
func GetPromptTemplate(id string) PromptTemplate {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return templateCache[id]
}

// NewGoTemplate returns a prompt template from the cache if it matches the id or creates a new one if it does not exist
func NewGoTemplate(id, content string) (PromptTemplate, error) {
	//Check if the template is already in the cache
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if _, ok := templateCache[id]; ok {
		LOGGER.WarnF("Replacing template with id %s", id)
	}

	//Create a new template
	tmpl, err := template.New(id).Parse(content)
	if err != nil {
		return nil, err
	}
	//Add the template to the cache
	templateCache[id] = &goTemplate{template: tmpl, id: id}
	return templateCache[id], nil

}

// init initializes the cache
func init() {
	cacheMutex = sync.RWMutex{}
	templateCache = make(map[string]PromptTemplate)
}
