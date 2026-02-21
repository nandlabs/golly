// Package genai provides types and utilities for generic AI messaging and roles.
package genai

import (
	"fmt"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/textutils"
)

// Role represents the type of participant in a conversation (system, user, assistant).
type Role int

// String returns the string representation of a Role.
func (r Role) String() string {
	switch r {
	case RoleSystem:
		return "system"
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	default:
		return "unknown"
	}
}

// RoleSystem indicates a system message (instructions, context).
// RoleUser indicates a user message (input, query).
// RoleAssistant indicates an assistant message (AI response).
const (
	RoleSystem Role = iota
	RoleUser
	RoleAssistant
)

// Message represents a single message in a conversation, with a role and content parts.
type Message struct {
	Role  Role   `json:"role" yaml:"role" toml:"role"`    // The role of the sender (system, user, assistant)
	Parts []Part `json:"parts" yaml:"parts" toml:"parts"` // The content parts of the message
}

// Part represents a named section of a message, which can be text, file, binary, function call, or response.
type Part struct {
	Name         string                 `json:"name" yaml:"name" toml:"name"`                                                          // Name of the part
	Text         *TextPart              `json:"text,omitempty" yaml:"text,omitempty" toml:"text,omitempty"`                            // Text content
	File         *FilePart              `json:"file,omitempty" yaml:"file,omitempty" toml:"file,omitempty"`                            // File reference
	Bin          *BinPart               `json:"bin,omitempty" yaml:"bin,omitempty" toml:"bin,omitempty"`                               // Binary data
	FuncCall     *FuncCallPart          `json:"func_call,omitempty" yaml:"func_call,omitempty" toml:"func_call,omitempty"`             // Function call details
	FuncResponse *FuncResponsePart      `json:"func_response,omitempty" yaml:"func_response,omitempty" toml:"func_response,omitempty"` // Function response details
	MimeType     string                 `json:"mime" yaml:"mime" toml:"mime"`                                                          // MIME type of the part
	Attributes   map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"`          // Additional attributes
}

// TextPart contains plain text content for a message part.
type TextPart struct {
	Text string `json:"text" yaml:"text" toml:"text"` // The text content

}

// FilePart contains file reference information for a message part.
type FilePart struct {
	URI string `json:"uri" yaml:"uri" toml:"uri"` // URI of the file
}

// BinPart contains binary data for a message part.
type BinPart struct {
	Data []byte `json:"data" yaml:"data" toml:"data"` // Binary data
}

// FuncCallPart contains details of a function call within a message part.
type FuncCallPart struct {
	Id           string                 `json:"id" yaml:"id" toml:"id"`                                  // Unique ID for the function call
	FunctionName string                 `json:"function_name" yaml:"function_name" toml:"function_name"` // Name of the function
	Arguments    map[string]interface{} `json:"arguments" yaml:"arguments" toml:"arguments"`             // Arguments for the function call
}

// FuncResponsePart contains the response from a function call, including content type and data.
type FuncResponsePart struct {
	Text    *string `json:"text" yaml:"text" toml:"text"`             // Textual response
	FileURI *string `json:"file_uri" yaml:"file_uri" toml:"file_uri"` // URI of a file response
	Data    []byte  `json:"data" yaml:"data" toml:"data"`             // Binary response data
}

// NewTextMessage creates a new Message with the specified role and text content.
func NewTextMessage(role Role, text string) *Message {
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     "text",
				MimeType: ioutils.MimeTextPlain,
				Text: &TextPart{
					Text: text,
				},
			},
		},
	}
}

// NewJsonMessage creates a new Message with the specified role and JSON-encoded content.
func NewJsonMessage(role Role, name string, v interface{}) (*Message, error) {
	c := codec.JsonCodec()
	jsonStr, err := c.EncodeToString(v)
	if err != nil {
		return nil, err
	}
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     name,
				MimeType: ioutils.MimeApplicationJSON,
				Text: &TextPart{
					Text: jsonStr,
				},
			},
		},
	}, nil
}

// AddJsonPart adds a JSON part to an existing Message.
func AddJsonPart(msg *Message, name string, v interface{}) error {
	c := codec.JsonCodec()
	jsonStr, err := c.EncodeToString(v)
	if err != nil {
		return err
	}
	msg.Parts = append(msg.Parts, Part{
		Name:     name,
		MimeType: ioutils.MimeApplicationJSON,
		Text: &TextPart{
			Text: jsonStr,
		},
	})
	return nil
}

// NewYamlMessage creates a new Message with the specified role and YAML-encoded content.
func NewYamlMessage(role Role, name string, v interface{}) (*Message, error) {
	c := codec.YamlCodec()
	yamlStr, err := c.EncodeToString(v)
	if err != nil {
		return nil, err
	}
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     name,
				MimeType: ioutils.MimeTextYAML,
				Text: &TextPart{
					Text: yamlStr,
				},
			},
		},
	}, nil
}

// AddYamlPart adds a YAML part to an existing Message.
func AddYamlPart(msg *Message, name string, v interface{}) error {
	c := codec.YamlCodec()
	yamlStr, err := c.EncodeToString(v)
	if err != nil {
		return err
	}
	msg.Parts = append(msg.Parts, Part{
		Name:     name,
		MimeType: ioutils.MimeTextYAML,
		Text: &TextPart{
			Text: yamlStr,
		},
	})
	return nil
}

// AddTextPart adds a text part to an existing Message.
func AddTextPart(msg *Message, name, text string) {
	msg.Parts = append(msg.Parts, Part{
		Name:     name,
		MimeType: ioutils.MimeTextPlain,
		Text: &TextPart{
			Text: text,
		},
	})
}

// NewFileMessage creates a new Message with the specified role and file URI.
func NewFileMessage(role Role, name, fileURI, mimeType string) *Message {
	mime := mimeType
	if mime == textutils.EmptyStr {
		mime = ioutils.MimeApplicationOctetStream
	}
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     name,
				MimeType: mime,
				File: &FilePart{
					URI: fileURI,
				},
			},
		},
	}
}

// AddFilePart adds a file part to an existing Message.
func AddFilePart(msg *Message, name, fileURI, mimeType string) {
	mime := mimeType
	if mime == textutils.EmptyStr {
		mime = ioutils.MimeApplicationOctetStream
	}
	msg.Parts = append(msg.Parts, Part{
		Name:     name,
		MimeType: mime,
		File: &FilePart{
			URI: fileURI,
		},
	})
}

// NewBinMessage adds a binary part to an existing Message.
func NewBinMessage(role Role, name string, data []byte, mimeType string) *Message {
	mime := mimeType
	if mime == textutils.EmptyStr {
		mime = ioutils.MimeApplicationOctetStream
	}
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     name,
				MimeType: mime,
				Bin: &BinPart{
					Data: data,
				},
			},
		},
	}
}

// AddBinPart adds a binary part to an existing Message.
func AddBinPart(msg *Message, name string, data []byte, mimeType string) {
	mime := mimeType
	if mime == textutils.EmptyStr {
		mime = ioutils.MimeApplicationOctetStream
	}
	msg.Parts = append(msg.Parts, Part{
		Name:     name,
		MimeType: mime,
		Bin: &BinPart{
			Data: data,
		},
	})
}

func NewMsgFromPrompt(role Role, name, prompt string) *Message {
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     name,
				MimeType: ioutils.MimeTextPlain,
				Text: &TextPart{
					Text: prompt,
				},
			},
		},
	}
}

func NewMsgFromPromptTemplate(role Role, pt *PromptTemplate, variables map[string]any) (*Message, error) {
	prompt, err := pt.Format(variables)
	if err != nil {
		return nil, err
	}
	return &Message{
		Role: role,
		Parts: []Part{
			{
				Name:     pt.Name,
				MimeType: ioutils.MimeTextPlain,
				Text: &TextPart{
					Text: prompt,
				},
			},
		},
	}, nil
}

func NewMsgFromPromptId(role Role, store PromptStore, promptId string, variables map[string]any) (*Message, error) {
	pt, exists := store.Get(promptId)
	if !exists {
		return nil, fmt.Errorf("prompt template with id '%s' not found", promptId)
	}
	return NewMsgFromPromptTemplate(role, pt, variables)
}
