package genai

import (
	"bytes"
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestGoTemplate_Type(t *testing.T) {
	template, err := NewGoTemplate("test-01", "content")
	assert.Equal(t, GoTextTemplate, template.Type())
	assert.NoError(t, err)
}

func TestGoTemplate_FormatAsText(t *testing.T) {
	template, err := NewGoTemplate("test-02", "content")
	assert.NoError(t, err)
	data := make(map[string]interface{})
	_, err = template.FormatAsText(data)
	assert.NoError(t, err)
}

func TestGoTemplate_WriteTo(t *testing.T) {
	template, err := NewGoTemplate("test-03", "content")
	assert.NoError(t, err)
	data := make(map[string]interface{})
	writer := &bytes.Buffer{}
	err = template.WriteTo(writer, data)
	assert.NoError(t, err)

}

func TestGetPromptTemplate(t *testing.T) {
	_, err := NewGoTemplate("test-03", "content")
	assert.Nil(t, err)
	template := GetPromptTemplate("test-03")
	assert.NotNil(t, template)
}

func TestGetOrCreate(t *testing.T) {
	template, err := NewGoTemplate("test-05", "content")
	assert.NoError(t, err)
	assert.NotNil(t, template)
}

func TestPrepareData(t *testing.T) {
	data := make(map[string]interface{})
	preparedData := prepareData(data)
	assert.NotNil(t, preparedData)
}
