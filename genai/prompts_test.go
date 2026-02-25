package genai

import (
	"fmt"
	"strings"
	"testing"

	"oss.nandlabs.io/golly/assertion"
)

// TestNewPromptTemplate validates creation of PromptTemplate with various template formats and error cases.
func TestNewPromptTemplate(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		templateName string
		template     string
		wantErr      bool
	}{
		{
			name:         "Simple template with placeholder",
			id:           "test-1",
			templateName: "greeting",
			template:     "Hello {{Name}}!",
			wantErr:      false,
		},
		{
			name:         "Template with multiple placeholders",
			id:           "test-2",
			templateName: "message",
			template:     "Hello {{Name}}, you have {{Count}} new messages.",
			wantErr:      false,
		},
		{
			name:         "Template already in Go format",
			id:           "test-3",
			templateName: "goTemplate",
			template:     "Hello {{.Name}}!",
			wantErr:      false,
		},
		{
			name:         "Template with no placeholders",
			id:           "test-4",
			templateName: "static",
			template:     "This is a static template.",
			wantErr:      false,
		},
		{
			name:         "Template with spaces in placeholder",
			id:           "test-5",
			templateName: "spaces",
			template:     "Hello {{ Name }}!",
			wantErr:      false,
		},
		{
			name:         "Empty template",
			id:           "test-6",
			templateName: "empty",
			template:     "",
			wantErr:      false,
		},
		{
			name:         "Invalid template syntax",
			id:           "test-7",
			templateName: "invalid",
			template:     "Hello {{.Name",
			wantErr:      true,
		},
		{
			name:         "Template with unclosed braces",
			id:           "test-8",
			templateName: "unclosed",
			template:     "Hello {{Name",
			wantErr:      true, // This causes a parse error since it's treated as an unclosed template action
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := NewPromptTemplate(tt.id, tt.templateName, tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPromptTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !assertion.NotEqual(pt, nil) {
					t.Error("NewPromptTemplate() returned nil without error")
					return
				}
				if !assertion.Equal(pt.Id, tt.id) {
					t.Errorf("NewPromptTemplate() Id = %v, want %v", pt.Id, tt.id)
				}
				if !assertion.Equal(pt.Name, tt.templateName) {
					t.Errorf("NewPromptTemplate() Name = %v, want %v", pt.Name, tt.templateName)
				}
				if !assertion.Equal(pt.Template, tt.template) {
					t.Errorf("NewPromptTemplate() Template = %v, want %v", pt.Template, tt.template)
				}
				if !assertion.NotEqual(pt.parsedTemplate, nil) {
					t.Error("NewPromptTemplate() parsedTemplate is nil")
				}
			}
		})
	}
}

// TestPromptTemplate_Format checks correct formatting/substitution of parameters in templates for various cases.
func TestPromptTemplate_Format(t *testing.T) {
	tests := []struct {
		name     string
		template string
		params   map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "Simple substitution",
			template: "Hello {{Name}}!",
			params: map[string]interface{}{
				"Name": "World",
			},
			want:    "Hello World!",
			wantErr: false,
		},
		{
			name:     "Multiple substitutions",
			template: "Hello {{Name}}, you have {{Count}} new messages.",
			params: map[string]interface{}{
				"Name":  "Alice",
				"Count": 5,
			},
			want:    "Hello Alice, you have 5 new messages.",
			wantErr: false,
		},
		{
			name:     "Go template format",
			template: "Hello {{.Name}}!",
			params: map[string]interface{}{
				"Name": "Bob",
			},
			want:    "Hello Bob!",
			wantErr: false,
		},
		{
			name:     "Missing parameter",
			template: "Hello {{Name}}!",
			params:   map[string]interface{}{},
			want:     "Hello <no value>!",
			wantErr:  false,
		},
		{
			name:     "Empty params",
			template: "Hello World!",
			params:   map[string]interface{}{},
			want:     "Hello World!",
			wantErr:  false,
		},
		{
			name:     "Numeric parameter",
			template: "The answer is {{Answer}}",
			params: map[string]interface{}{
				"Answer": 42,
			},
			want:    "The answer is 42",
			wantErr: false,
		},
		{
			name:     "Boolean parameter",
			template: "Status: {{Active}}",
			params: map[string]interface{}{
				"Active": true,
			},
			want:    "Status: true",
			wantErr: false,
		},
		{
			name:     "Complex object parameter",
			template: "User: {{User.Name}}, Age: {{User.Age}}",
			params: map[string]interface{}{
				"User": map[string]interface{}{
					"Name": "Charlie",
					"Age":  30,
				},
			},
			want:    "User: Charlie, Age: 30",
			wantErr: false,
		},
		{
			name:     "Template with spaces in placeholder",
			template: "Hello {{ Name }}!",
			params: map[string]interface{}{
				"Name": "Dave",
			},
			want:    "Hello Dave!",
			wantErr: false,
		},
		{
			name:     "Multi-line template",
			template: "Hello {{Name}},\n\nYou have {{Count}} messages.\n\nBest regards",
			params: map[string]interface{}{
				"Name":  "Eve",
				"Count": 3,
			},
			want:    "Hello Eve,\n\nYou have 3 messages.\n\nBest regards",
			wantErr: false,
		},
		{
			name:     "Empty string parameter",
			template: "Hello {{Name}}!",
			params: map[string]interface{}{
				"Name": "",
			},
			want:    "Hello !",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := NewPromptTemplate("test-id", "test-template", tt.template)
			if err != nil {
				t.Fatalf("NewPromptTemplate() error = %v", err)
			}

			got, err := pt.Format(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assertion.Equal(got, tt.want) {
				t.Errorf("PromptTemplate.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConvertToGoTemplate ensures conversion of custom template syntax to Go template format for a variety of input patterns.
func TestConvertToGoTemplate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple placeholder",
			input: "Hello {{Name}}!",
			want:  "Hello {{.Name}}!",
		},
		{
			name:  "Multiple placeholders",
			input: "{{Greeting}} {{Name}}, you have {{Count}} messages.",
			want:  "{{.Greeting}} {{.Name}}, you have {{.Count}} messages.",
		},
		{
			name:  "Already in Go format",
			input: "Hello {{.Name}}!",
			want:  "Hello {{.Name}}!",
		},
		{
			name:  "Mixed format (should detect Go format)",
			input: "Hello {{.Name}}, your id is {{Id}}",
			want:  "Hello {{.Name}}, your id is {{Id}}", // Returns as is since it contains {{.
		},
		{
			name:  "No placeholders",
			input: "Hello World!",
			want:  "Hello World!",
		},
		{
			name:  "Placeholder with spaces",
			input: "Hello {{ Name }}!",
			want:  "Hello {{.Name}}!",
		},
		{
			name:  "Unclosed placeholder",
			input: "Hello {{Name",
			want:  "Hello {{Name",
		},
		{
			name:  "Only opening braces",
			input: "Hello {{",
			want:  "Hello {{",
		},
		{
			name:  "Only closing braces",
			input: "Hello }}",
			want:  "Hello }}",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Multiple consecutive placeholders",
			input: "{{First}}{{Second}}{{Third}}",
			want:  "{{.First}}{{.Second}}{{.Third}}",
		},
		{
			name:  "Placeholder at start",
			input: "{{Name}} says hello",
			want:  "{{.Name}} says hello",
		},
		{
			name:  "Placeholder at end",
			input: "Hello {{Name}}",
			want:  "Hello {{.Name}}",
		},
		{
			name:  "Multi-line template",
			input: "Hello {{Name}},\n\nYour count is {{Count}}.",
			want:  "Hello {{.Name}},\n\nYour count is {{.Count}}.",
		},
		{
			name:  "Nested braces (not valid but should handle)",
			input: "Hello {{{Name}}}",
			want:  "Hello {{.{Name}}}",
		},
		{
			name:  "Placeholder with underscores",
			input: "Hello {{User_Name}}!",
			want:  "Hello {{.User_Name}}!",
		},
		{
			name:  "Placeholder with numbers",
			input: "Hello {{Name123}}!",
			want:  "Hello {{.Name123}}!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToGoTemplate(tt.input)
			if !assertion.Equal(got, tt.want) {
				t.Errorf("convertToGoTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPromptTemplate_FormatWithNilParams verifies formatting works when parameters are nil.
func TestPromptTemplate_FormatWithNilParams(t *testing.T) {
	pt, err := NewPromptTemplate("test-id", "test-template", "Hello World!")
	if err != nil {
		t.Fatalf("NewPromptTemplate() error = %v", err)
	}

	got, err := pt.Format(nil)
	if err != nil {
		t.Errorf("PromptTemplate.Format() with nil params error = %v", err)
	}
	if !assertion.Equal(got, "Hello World!") {
		t.Errorf("PromptTemplate.Format() = %v, want %v", got, "Hello World!")
	}
}

// TestPromptTemplate_FormatEdgeCases tests formatting with special characters, newlines, tabs, and unicode.
func TestPromptTemplate_FormatEdgeCases(t *testing.T) {
	t.Run("Template with special characters", func(t *testing.T) {
		pt, err := NewPromptTemplate("test-id", "test", "Hello {{Name}}! @#$%^&*()")
		if err != nil {
			t.Fatalf("NewPromptTemplate() error = %v", err)
		}
		got, err := pt.Format(map[string]interface{}{"Name": "Test"})
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}
		want := "Hello Test! @#$%^&*()"
		if !assertion.Equal(got, want) {
			t.Errorf("Format() = %v, want %v", got, want)
		}
	})

	t.Run("Template with newlines and tabs", func(t *testing.T) {
		pt, err := NewPromptTemplate("test-id", "test", "Line1\n\tLine2\n{{Name}}")
		if err != nil {
			t.Fatalf("NewPromptTemplate() error = %v", err)
		}
		got, err := pt.Format(map[string]interface{}{"Name": "Test"})
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}
		want := "Line1\n\tLine2\nTest"
		if !assertion.Equal(got, want) {
			t.Errorf("Format() = %v, want %v", got, want)
		}
	})

	t.Run("Template with unicode characters", func(t *testing.T) {
		pt, err := NewPromptTemplate("test-id", "test", "你好 {{Name}}! 🎉")
		if err != nil {
			t.Fatalf("NewPromptTemplate() error = %v", err)
		}
		got, err := pt.Format(map[string]interface{}{"Name": "世界"})
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}
		want := "你好 世界! 🎉"
		if !assertion.Equal(got, want) {
			t.Errorf("Format() = %v, want %v", got, want)
		}
	})
}

// TestPromptTemplateStructFields confirms struct fields are set correctly.
func TestPromptTemplateStructFields(t *testing.T) {
	id := "test-id-123"
	name := "test-template-name"
	template := "Hello {{Name}}!"

	pt, err := NewPromptTemplate(id, name, template)
	if err != nil {
		t.Fatalf("NewPromptTemplate() error = %v", err)
	}

	if !assertion.Equal(pt.Id, id) {
		t.Errorf("PromptTemplate.Id = %v, want %v", pt.Id, id)
	}
	if !assertion.Equal(pt.Name, name) {
		t.Errorf("PromptTemplate.Name = %v, want %v", pt.Name, name)
	}
	if !assertion.Equal(pt.Template, template) {
		t.Errorf("PromptTemplate.Template = %v, want %v", pt.Template, template)
	}
	if !assertion.NotEqual(pt.parsedTemplate, nil) {
		t.Error("PromptTemplate.parsedTemplate is nil")
	}
}

// BenchmarkNewPromptTemplate benchmarks creation of PromptTemplate.
func BenchmarkNewPromptTemplate(b *testing.B) {
	template := "Hello {{Name}}, you have {{Count}} new messages from {{Sender}}."
	for i := 0; i < b.N; i++ {
		_, _ = NewPromptTemplate("bench-id", "bench-template", template)
	}
}

// BenchmarkPromptTemplate_Format benchmarks formatting operation.
func BenchmarkPromptTemplate_Format(b *testing.B) {
	pt, err := NewPromptTemplate("bench-id", "bench-template", "Hello {{Name}}, you have {{Count}} new messages.")
	if err != nil {
		b.Fatalf("NewPromptTemplate() error = %v", err)
	}

	params := map[string]interface{}{
		"Name":  "Alice",
		"Count": 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pt.Format(params)
	}
}

// BenchmarkConvertToGoTemplate benchmarks template conversion.
func BenchmarkConvertToGoTemplate(b *testing.B) {
	template := "Hello {{Name}}, you have {{Count}} messages from {{Sender}} about {{Topic}}."
	for i := 0; i < b.N; i++ {
		_ = convertToGoTemplate(template)
	}
}

// TestPromptTemplate_FormatLargeTemplate validates formatting for a large, realistic template and checks all placeholders are replaced.
func TestPromptTemplate_FormatLargeTemplate(t *testing.T) {
	// Test with a larger, more realistic template
	template := `
Dear {{Name}},

Thank you for your {{Action}} on {{Date}}. We have received your request 
regarding {{Topic}} and are processing it with reference number {{RefNumber}}.

Your account status: {{Status}}
Priority level: {{Priority}}

We will contact you at {{Email}} within {{Days}} business days.

Best regards,
{{CompanyName}} Team
`
	pt, err := NewPromptTemplate("large-test", "large-template", template)
	if err != nil {
		t.Fatalf("NewPromptTemplate() error = %v", err)
	}

	params := map[string]interface{}{
		"Name":        "John Doe",
		"Action":      "inquiry",
		"Date":        "2025-11-26",
		"Topic":       "account upgrade",
		"RefNumber":   "REF-2025-001",
		"Status":      "Active",
		"Priority":    "High",
		"Email":       "john@example.com",
		"Days":        3,
		"CompanyName": "ACME Corp",
	}

	got, err := pt.Format(params)
	if err != nil {
		t.Errorf("Format() error = %v", err)
	}

	// Check that all parameters were replaced
	for key := range params {
		placeholder := "{{" + key + "}}"
		if strings.Contains(got, placeholder) {
			t.Errorf("Format() still contains placeholder %s", placeholder)
		}
	}

	// Check some expected content
	if !strings.Contains(got, "John Doe") {
		t.Error("Format() missing expected Name value")
	}
	if !strings.Contains(got, "REF-2025-001") {
		t.Error("Format() missing expected RefNumber value")
	}
}

// InMemoryPromptStore Tests

// TestNewInMemoryPromptStore verifies initialization of the store and its internal map.
func TestNewInMemoryPromptStore(t *testing.T) {
	store := NewInMemoryPromptStore()
	if !assertion.NotEqual(store, nil) {
		t.Error("NewInMemoryPromptStore() returned nil")
	}
	if !assertion.NotEqual(store.templates, nil) {
		t.Error("NewInMemoryPromptStore() templates map is nil")
	}
	if !assertion.Equal(len(store.templates), 0) {
		t.Errorf("NewInMemoryPromptStore() templates map should be empty, got %d items", len(store.templates))
	}
}

// TestInMemoryPromptStore_AddPromptTemplate tests adding new, duplicate, and empty-ID templates.
func TestInMemoryPromptStore_AddPromptTemplate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*InMemoryPromptStore)
		pt      *PromptTemplate
		wantErr bool
		errMsg  string
	}{
		{
			name:  "Add new template successfully",
			setup: func(s *InMemoryPromptStore) {},
			pt: &PromptTemplate{
				Id:       "test-1",
				Name:     "greeting",
				Template: "Hello {{Name}}!",
			},
			wantErr: false,
		},
		{
			name: "Add duplicate template returns error",
			setup: func(s *InMemoryPromptStore) {
				pt, _ := NewPromptTemplate("test-1", "existing", "Hello!")
				s.templates["test-1"] = pt
			},
			pt: &PromptTemplate{
				Id:       "test-1",
				Name:     "duplicate",
				Template: "Goodbye {{Name}}!",
			},
			wantErr: true,
			errMsg:  "prompt template with id 'test-1' already exists",
		},
		{
			name:  "Add multiple different templates",
			setup: func(s *InMemoryPromptStore) {},
			pt: &PromptTemplate{
				Id:       "test-2",
				Name:     "farewell",
				Template: "Goodbye {{Name}}!",
			},
			wantErr: false,
		},
		{
			name:  "Add template with empty ID",
			setup: func(s *InMemoryPromptStore) {},
			pt: &PromptTemplate{
				Id:       "",
				Name:     "empty-id",
				Template: "Test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryPromptStore()
			tt.setup(store)

			err := store.Add(tt.pt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Add() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				// Verify template was added
				retrieved, ok := store.templates[tt.pt.Id]
				if !assertion.Equal(ok, true) {
					t.Error("Add() template not found in store")
				}
				if !assertion.Equal(retrieved, tt.pt) {
					t.Errorf("Add() stored template = %v, want %v", retrieved, tt.pt)
				}
			}
		})
	}
}

// TestInMemoryPromptStore_GetPromptTemplate checks retrieval of existing, non-existing, and empty-ID templates.
func TestInMemoryPromptStore_GetPromptTemplate(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*InMemoryPromptStore) *PromptTemplate
		id     string
		wantOk bool
	}{
		{
			name: "Get existing template",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				pt, _ := NewPromptTemplate("test-1", "greeting", "Hello {{Name}}!")
				s.templates["test-1"] = pt
				return pt
			},
			id:     "test-1",
			wantOk: true,
		},
		{
			name: "Get non-existing template",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				return nil
			},
			id:     "non-existent",
			wantOk: false,
		},
		{
			name: "Get template with empty ID",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				return nil
			},
			id:     "",
			wantOk: false,
		},
		{
			name: "Get one of many templates",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				pt1, _ := NewPromptTemplate("test-1", "greeting", "Hello!")
				pt2, _ := NewPromptTemplate("test-2", "farewell", "Goodbye!")
				pt3, _ := NewPromptTemplate("test-3", "question", "How are you?")
				s.templates["test-1"] = pt1
				s.templates["test-2"] = pt2
				s.templates["test-3"] = pt3
				return pt2
			},
			id:     "test-2",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryPromptStore()
			expectedPt := tt.setup(store)

			got, ok := store.Get(tt.id)
			if !assertion.Equal(ok, tt.wantOk) {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk {
				if !assertion.Equal(got, expectedPt) {
					t.Errorf("Get() = %v, want %v", got, expectedPt)
				}
			} else {
				if !assertion.Equal(got, (*PromptTemplate)(nil)) {
					t.Errorf("Get() should return nil for non-existent template, got %v", got)
				}
			}
		})
	}
}

// TestInMemoryPromptStore_ListPromptTemplates validates listing templates for empty, single, multiple, and many templates.
func TestInMemoryPromptStore_ListPromptTemplates(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*InMemoryPromptStore)
		wantCount int
	}{
		{
			name:      "List empty store",
			setup:     func(s *InMemoryPromptStore) {},
			wantCount: 0,
		},
		{
			name: "List single template",
			setup: func(s *InMemoryPromptStore) {
				pt, _ := NewPromptTemplate("test-1", "greeting", "Hello!")
				s.templates["test-1"] = pt
			},
			wantCount: 1,
		},
		{
			name: "List multiple templates",
			setup: func(s *InMemoryPromptStore) {
				pt1, _ := NewPromptTemplate("test-1", "greeting", "Hello!")
				pt2, _ := NewPromptTemplate("test-2", "farewell", "Goodbye!")
				pt3, _ := NewPromptTemplate("test-3", "question", "How?")
				s.templates["test-1"] = pt1
				s.templates["test-2"] = pt2
				s.templates["test-3"] = pt3
			},
			wantCount: 3,
		},
		{
			name: "List many templates",
			setup: func(s *InMemoryPromptStore) {
				for i := 0; i < 10; i++ {
					id := fmt.Sprintf("test-%d", i)
					pt, _ := NewPromptTemplate(id, fmt.Sprintf("name-%d", i), "Template")
					s.templates[id] = pt
				}
			},
			wantCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryPromptStore()
			tt.setup(store)

			got := store.List()
			if !assertion.Equal(len(got), tt.wantCount) {
				t.Errorf("List() count = %v, want %v", len(got), tt.wantCount)
			}

			// Verify all returned templates exist in the store
			for _, pt := range got {
				if _, exists := store.templates[pt.Id]; !assertion.Equal(exists, true) {
					t.Errorf("List() returned template with id %s not in store", pt.Id)
				}
			}
		})
	}
}

// TestInMemoryPromptStore_RemovePromptTemplate tests removal of existing, non-existing, empty-ID, and one-of-many templates.
func TestInMemoryPromptStore_RemovePromptTemplate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*InMemoryPromptStore)
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Remove existing template",
			setup: func(s *InMemoryPromptStore) {
				pt, _ := NewPromptTemplate("test-1", "greeting", "Hello!")
				s.templates["test-1"] = pt
			},
			id:      "test-1",
			wantErr: false,
		},
		{
			name:    "Remove non-existing template",
			setup:   func(s *InMemoryPromptStore) {},
			id:      "non-existent",
			wantErr: true,
			errMsg:  "prompt template with id 'non-existent' does not exist",
		},
		{
			name:    "Remove with empty ID",
			setup:   func(s *InMemoryPromptStore) {},
			id:      "",
			wantErr: true,
			errMsg:  "prompt template with id '' does not exist",
		},
		{
			name: "Remove one of many templates",
			setup: func(s *InMemoryPromptStore) {
				pt1, _ := NewPromptTemplate("test-1", "greeting", "Hello!")
				pt2, _ := NewPromptTemplate("test-2", "farewell", "Goodbye!")
				pt3, _ := NewPromptTemplate("test-3", "question", "How?")
				s.templates["test-1"] = pt1
				s.templates["test-2"] = pt2
				s.templates["test-3"] = pt3
			},
			id:      "test-2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryPromptStore()
			tt.setup(store)
			initialCount := len(store.templates)

			err := store.Remove(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Remove() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				// Verify template was removed
				_, exists := store.templates[tt.id]
				if !assertion.Equal(exists, false) {
					t.Error("Remove() template still exists in store")
				}
				// Verify count decreased
				if !assertion.Equal(len(store.templates), initialCount-1) {
					t.Errorf("Remove() count = %v, want %v", len(store.templates), initialCount-1)
				}
			}
		})
	}
}

// TestInMemoryPromptStore_UpdatePromptTemplate validates updating existing, non-existing, empty-ID, name-only, version, and one-of-many templates.
func TestInMemoryPromptStore_UpdatePromptTemplate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*InMemoryPromptStore) *PromptTemplate
		update  func(*PromptTemplate) *PromptTemplate
		wantErr bool
		errMsg  string
	}{
		{
			name: "Update existing template",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				pt, _ := NewPromptTemplate("test-1", "greeting", "Hello {{Name}}!")
				s.templates["test-1"] = pt
				return pt
			},
			update: func(pt *PromptTemplate) *PromptTemplate {
				pt.Template = "Hi {{Name}}!"
				pt.Name = "updated-greeting"
				return pt
			},
			wantErr: false,
		},
		{
			name: "Update non-existing template",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				return nil
			},
			update: func(pt *PromptTemplate) *PromptTemplate {
				ptNew, _ := NewPromptTemplate("non-existent", "test", "Hello!")
				return ptNew
			},
			wantErr: true,
			errMsg:  "prompt template with id 'non-existent' does not exist",
		},
		{
			name: "Update template with empty ID",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				return nil
			},
			update: func(pt *PromptTemplate) *PromptTemplate {
				ptNew, _ := NewPromptTemplate("", "empty", "Test")
				return ptNew
			},
			wantErr: true,
			errMsg:  "prompt template with id '' does not exist",
		},
		{
			name: "Update template name only",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				pt, _ := NewPromptTemplate("test-2", "original", "Hello!")
				s.templates["test-2"] = pt
				return pt
			},
			update: func(pt *PromptTemplate) *PromptTemplate {
				pt.Name = "updated-name"
				return pt
			},
			wantErr: false,
		},
		{
			name: "Update template version",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				pt, _ := NewPromptTemplate("test-3", "versioned", "Template")
				pt.Version = "1.0.0"
				s.templates["test-3"] = pt
				return pt
			},
			update: func(pt *PromptTemplate) *PromptTemplate {
				pt.Version = "2.0.0"
				return pt
			},
			wantErr: false,
		},
		{
			name: "Update one of many templates",
			setup: func(s *InMemoryPromptStore) *PromptTemplate {
				pt1, _ := NewPromptTemplate("test-1", "greeting", "Hello!")
				pt2, _ := NewPromptTemplate("test-2", "farewell", "Goodbye!")
				pt3, _ := NewPromptTemplate("test-3", "question", "How?")
				s.templates["test-1"] = pt1
				s.templates["test-2"] = pt2
				s.templates["test-3"] = pt3
				return pt2
			},
			update: func(pt *PromptTemplate) *PromptTemplate {
				pt.Template = "See you later!"
				return pt
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryPromptStore()
			originalPt := tt.setup(store)

			var updatedPt *PromptTemplate
			if originalPt != nil {
				updatedPt = tt.update(originalPt)
			} else {
				updatedPt = tt.update(&PromptTemplate{})
			}

			// Store the original UpdatedAt value before calling UpdatePromptTemplate
			var originalUpdatedAt int64
			if originalPt != nil {
				originalUpdatedAt = originalPt.UpdatedAt
			}

			err := store.Update(updatedPt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Update() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				// Verify template was updated
				retrieved, exists := store.templates[updatedPt.Id]
				if !assertion.Equal(exists, true) {
					t.Error("Update() template not found in store")
				}
				if !assertion.Equal(retrieved.Name, updatedPt.Name) {
					t.Errorf("Update() Name = %v, want %v", retrieved.Name, updatedPt.Name)
				}
				if !assertion.Equal(retrieved.Template, updatedPt.Template) {
					t.Errorf("Update() Template = %v, want %v", retrieved.Template, updatedPt.Template)
				}
				// Verify UpdatedAt was modified (should be greater than or equal to original)
				if originalUpdatedAt > 0 && retrieved.UpdatedAt < originalUpdatedAt {
					t.Errorf("Update() UpdatedAt should be >= original, got %v < %v", retrieved.UpdatedAt, originalUpdatedAt)
				}
			}
		})
	}
}

// TestInMemoryPromptStore_UpdatePromptTemplate_PreservesCreatedAt ensures CreatedAt is preserved and UpdatedAt is updated.
func TestInMemoryPromptStore_UpdatePromptTemplate_PreservesCreatedAt(t *testing.T) {
	store := NewInMemoryPromptStore()

	// Add initial template
	pt, _ := NewPromptTemplate("test-1", "original", "Hello {{Name}}!")
	originalCreatedAt := pt.CreatedAt
	store.Add(pt)

	// Update the template
	pt.Name = "updated"
	pt.Template = "Hi {{Name}}!"
	err := store.Update(pt)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify CreatedAt is preserved
	retrieved, _ := store.Get("test-1")
	if !assertion.Equal(retrieved.CreatedAt, originalCreatedAt) {
		t.Errorf("UpdatePromptTemplate() CreatedAt changed from %v to %v", originalCreatedAt, retrieved.CreatedAt)
	}

	// Verify UpdatedAt changed (should be greater than or equal to CreatedAt)
	if retrieved.UpdatedAt < originalCreatedAt {
		t.Errorf("UpdatePromptTemplate() UpdatedAt (%v) should be >= CreatedAt (%v)", retrieved.UpdatedAt, originalCreatedAt)
	}
}

// TestInMemoryPromptStore_Integration is a full CRUD integration test covering add, get, list, update, remove, and error cases.
func TestInMemoryPromptStore_Integration(t *testing.T) {
	store := NewInMemoryPromptStore()

	// Test 1: Store should be empty initially
	list := store.List()
	if !assertion.Equal(len(list), 0) {
		t.Errorf("Initial list should be empty, got %d items", len(list))
	}

	// Test 2: Add first template
	pt1, _ := NewPromptTemplate("greeting", "Greeting Template", "Hello {{Name}}!")
	err := store.Add(pt1)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}

	// Test 3: Verify template was added
	retrieved, ok := store.Get("greeting")
	if !assertion.Equal(ok, true) {
		t.Error("Get() could not find added template")
	}
	if !assertion.Equal(retrieved.Id, "greeting") {
		t.Errorf("Get() Id = %v, want greeting", retrieved.Id)
	}

	// Test 4: Add second template
	pt2, _ := NewPromptTemplate("farewell", "Farewell Template", "Goodbye {{Name}}!")
	err = store.Add(pt2)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}

	// Test 5: List should have 2 templates
	list = store.List()
	if !assertion.Equal(len(list), 2) {
		t.Errorf("List should have 2 templates, got %d", len(list))
	}

	// Test 6: Try to add duplicate
	pt3, _ := NewPromptTemplate("greeting", "Duplicate", "Test")
	err = store.Add(pt3)
	if err == nil {
		t.Error("Add() should return error for duplicate ID")
	}

	// Test 7: Update farewell template
	pt2.Template = "See you later {{Name}}!"
	pt2.Version = "2.0.0"
	err = store.Update(pt2)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Test 7.5: Verify template was updated
	retrieved, ok = store.Get("farewell")
	if !assertion.Equal(ok, true) {
		t.Error("Get() should find updated template")
	}
	if !assertion.Equal(retrieved.Template, "See you later {{Name}}!") {
		t.Errorf("Updated template content = %v, want 'See you later {{Name}}!'", retrieved.Template)
	}
	if !assertion.Equal(retrieved.Version, "2.0.0") {
		t.Errorf("Updated template version = %v, want '2.0.0'", retrieved.Version)
	}

	// Test 7.6: Try to update non-existent template
	pt4, _ := NewPromptTemplate("non-existent", "Test", "Test")
	err = store.Update(pt4)
	if err == nil {
		t.Error("Update() should return error for non-existent ID")
	}

	// Test 8: Remove first template
	err = store.Remove("greeting")
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}

	// Test 9: Verify template was removed
	_, ok = store.Get("greeting")
	if !assertion.Equal(ok, false) {
		t.Error("Get() should not find removed template")
	}

	// Test 10: List should have 1 template
	list = store.List()
	if !assertion.Equal(len(list), 1) {
		t.Errorf("List should have 1 template, got %d", len(list))
	}

	// Test 11: Try to remove non-existent template
	err = store.Remove("non-existent")
	if err == nil {
		t.Error("Remove() should return error for non-existent ID")
	}
}

// TestCreatePrompt tests creation of prompts via the store, including duplicate and invalid template cases.
func TestCreatePrompt(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*InMemoryPromptStore)
		id       string
		ptName   string
		template string
		wantErr  bool
	}{
		{
			name:     "Create new prompt successfully",
			setup:    func(s *InMemoryPromptStore) {},
			id:       "test-1",
			ptName:   "greeting",
			template: "Hello {{Name}}!",
			wantErr:  false,
		},
		{
			name: "Create duplicate prompt returns error",
			setup: func(s *InMemoryPromptStore) {
				pt, _ := NewPromptTemplate("test-1", "existing", "Hello!")
				s.templates["test-1"] = pt
			},
			id:       "test-1",
			ptName:   "duplicate",
			template: "Goodbye!",
			wantErr:  true,
		},
		{
			name:     "Create prompt with invalid template",
			setup:    func(s *InMemoryPromptStore) {},
			id:       "test-2",
			ptName:   "invalid",
			template: "Hello {{.Name",
			wantErr:  true,
		},
		{
			name:     "Create prompt with complex template",
			setup:    func(s *InMemoryPromptStore) {},
			id:       "test-3",
			ptName:   "complex",
			template: "Hello {{Name}}, you have {{Count}} messages from {{Sender}}.",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryPromptStore()
			tt.setup(store)

			pt, err := CreatePrompt(store, tt.id, tt.ptName, tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !assertion.NotEqual(pt, nil) {
					t.Error("CreatePrompt() returned nil without error")
				}
				if !assertion.Equal(pt.Id, tt.id) {
					t.Errorf("CreatePrompt() Id = %v, want %v", pt.Id, tt.id)
				}
				if !assertion.Equal(pt.Name, tt.ptName) {
					t.Errorf("CreatePrompt() Name = %v, want %v", pt.Name, tt.ptName)
				}
				// Verify template was added to store
				retrieved, ok := store.Get(tt.id)
				if !assertion.Equal(ok, true) {
					t.Error("CreatePrompt() template not found in store")
				}
				if !assertion.Equal(retrieved, pt) {
					t.Error("CreatePrompt() template in store doesn't match returned template")
				}
			}
		})
	}
}

// TestInMemoryPromptStore_ConcurrentAccess is a basic concurrent access test for safe reads.
func TestInMemoryPromptStore_ConcurrentAccess(t *testing.T) {
	// Note: This is a basic concurrent access test
	// For production use, consider adding proper synchronization
	store := NewInMemoryPromptStore()

	// Add initial template
	pt, _ := NewPromptTemplate("test-1", "initial", "Hello!")
	store.Add(pt)

	// Perform concurrent reads (safe for current implementation)
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = store.Get("test-1")
			_ = store.List()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify store is still valid
	retrieved, ok := store.Get("test-1")
	if !assertion.Equal(ok, true) {
		t.Error("Template should still exist after concurrent reads")
	}
	if !assertion.Equal(retrieved.Id, "test-1") {
		t.Error("Template data corrupted after concurrent access")
	}
}

// BenchmarkInMemoryPromptStore_AddPromptTemplate benchmarks add operation.
func BenchmarkInMemoryPromptStore_AddPromptTemplate(b *testing.B) {
	store := NewInMemoryPromptStore()
	pt, _ := NewPromptTemplate("bench-id", "bench-template", "Hello {{Name}}!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.templates = make(map[string]*PromptTemplate) // Reset for each iteration
		_ = store.Add(pt)
	}
}

// BenchmarkInMemoryPromptStore_Get benchmarks get operation.
func BenchmarkInMemoryPromptStore_Get(b *testing.B) {
	store := NewInMemoryPromptStore()
	pt, _ := NewPromptTemplate("bench-id", "bench-template", "Hello {{Name}}!")
	store.Add(pt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get("bench-id")
	}
}

// BenchmarkInMemoryPromptStore_Update benchmarks update operation.
func BenchmarkInMemoryPromptStore_Update(b *testing.B) {
	store := NewInMemoryPromptStore()
	pt, _ := NewPromptTemplate("bench-id", "bench-template", "Hello {{Name}}!")
	store.Add(pt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pt.Template = "Updated template"
		_ = store.Update(pt)
	}
}

// BenchmarkInMemoryPromptStore_List benchmarks list operation.
func BenchmarkInMemoryPromptStore_List(b *testing.B) {
	store := NewInMemoryPromptStore()
	// Add 100 templates
	for i := 0; i < 100; i++ {
		pt, _ := NewPromptTemplate(fmt.Sprintf("id-%d", i), fmt.Sprintf("name-%d", i), "Template")
		store.Add(pt)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.List()
	}
}

// BenchmarkCreatePrompt benchmarks prompt creation via store.
func BenchmarkCreatePrompt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		store := NewInMemoryPromptStore()
		_, _ = CreatePrompt(store, "bench-id", "bench-template", "Hello {{Name}}!")
	}
}
