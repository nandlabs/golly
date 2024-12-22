package genai

import (
	"bytes"
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestGoTemplate_Type(t *testing.T) {
	template, err := NewGoTemplate("content")
	assert.Equal(t, GoTextTemplate, template.Type())
	assert.NoError(t, err)
}

func TestGoTemplate_FormatAsText(t *testing.T) {
	template, err := NewGoTemplate("content")
	assert.NoError(t, err)
	data := make(map[string]interface{})
	_, err = template.FormatAsText(data)
	assert.NoError(t, err)
}

func TestGoTemplate_WriteTo(t *testing.T) {
	template, err := NewGoTemplate("content")
	assert.NoError(t, err)
	data := make(map[string]interface{})
	writer := &bytes.Buffer{}
	err = template.WriteTo(writer, data)
	assert.NoError(t, err)
}

func TestGetPromptTemplate(t *testing.T) {
	template, err := NewGoTemplate("content")
	assert.NotNil(t, template)
	assert.NoError(t, err)
}

func TestGetOrCreate(t *testing.T) {
	template, err := GetOrCreatePrompt("test", "content")
	assert.NoError(t, err)
	assert.NotNil(t, template)
}

func TestPrepareData(t *testing.T) {
	data := make(map[string]interface{})
	preparedData := prepareData(data)
	assert.NotNil(t, preparedData)
}
