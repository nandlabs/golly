package genai

import (
	"context"
	"io"
)

type FinishReason string

const (
	FinishReasonStop             FinishReason = "stop"
	FinishReasonLength           FinishReason = "length"
	FinishReasonToolCall         FinishReason = "tool_call"
	FinishReasonToolResponse     FinishReason = "tool_response"
	FinishReasonFunctionCall     FinishReason = "function_call"
	FinishReasonFunctionResponse FinishReason = "function_response"
	FinishReasonContentFilter    FinishReason = "content_filter"
	FinishReasonUnknown          FinishReason = "unknown"
	FinishReasonError            FinishReason = "error"
	FinishReasonEndTurn          FinishReason = "end_turn"
	FinishReasonInProgress       FinishReason = "null"
)

type ResponseMeta struct {
	// Number of tokens cached for the response
	CachedTokens int `json:"cached_tokens,omitempty" yaml:"cached_tokens,omitempty" toml:"cached_tokens,omitempty"`
	// Number of input tokens used in the request
	InputTokens int `json:"input_tokens,omitempty" yaml:"input_tokens,omitempty" toml:"input_tokens,omitempty"`
	// Number of output tokens generated in the response
	OutputTokens int `json:"output_tokens,omitempty" yaml:"output_tokens,omitempty" toml:"output_tokens,omitempty"`
	// Total number of tokens (input + output)
	TotalTokens int `json:"total_tokens,omitempty" yaml:"total_tokens,omitempty" toml:"total_tokens,omitempty"`
	// Time taken to first response in milliseconds
	TimeToFirst int64 `json:"time_to_first,omitempty" yaml:"time_to_first,omitempty" toml:"time_to_first,omitempty"`
	// Total time taken for the request in milliseconds
	TotalTime int64 `json:"total_time,omitempty" yaml:"total_time,omitempty" toml:"total_time,omitempty"`
	// Any additional data related to the thinking process
	ThinkingData any `json:"thinking_data,omitempty" yaml:"thinking_data,omitempty" toml:"thinking_data,omitempty"`
	// Number of tokens generated during the thinking process
	ThinkingTokens int `json:"thinking_tokens,omitempty" yaml:"thinking_tokens,omitempty" toml:"thinking_tokens,omitempty"`
	// Time taken for the thinking process in milliseconds
	ThinkingTime int64 `json:"thinking_time,omitempty" yaml:"thinking_time,omitempty" toml:"thinking_time,omitempty"`
}
type GroundingInfo struct {
	// Name of the grounding source
	Name string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	// Type of the grounding source
	Source string `json:"source,omitempty" yaml:"source,omitempty" toml:"source,omitempty"`
	// URI of the grounding source
	URI string `json:"uri,omitempty" yaml:"uri,omitempty" toml:"uri,omitempty"`
	// Score or relevance of the grounding source
	Score float64 `json:"score,omitempty" yaml:"score,omitempty" toml:"score,omitempty"`
}

type Candidate struct {
	Index        int             `json:"index,omitempty" yaml:"index,omitempty" toml:"index,omitempty"`
	Message      *Message        `json:"message,omitempty" yaml:"message,omitempty" toml:"message,omitempty"`
	FinishReason FinishReason    `json:"finish_reason,omitempty" yaml:"finish_reason,omitempty" toml:"finish_reason,omitempty"`
	Groundings   []GroundingInfo `json:"groundings,omitempty" yaml:"groundings,omitempty" toml:"groundings,omitempty"`
}

type GenResponse struct {
	Candidates []Candidate  `json:"candidates,omitempty" yaml:"candidates,omitempty" toml:"candidates,omitempty"`
	Meta       ResponseMeta `json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta,omitempty"`
}

// Provider represents a generative AI service provider.

type Provider interface {
	io.Closer
	// Name returns the name of the provider.
	Name() string
	// Description returns a brief description of the provider.
	Description() string
	// Version returns the version of the provider.
	Version() string
	// Models returns the list of model ids supported by the provider.
	Models() []string
	// Generate generates a response based on the provided messages and parameters.
	Generate(ctx context.Context, model string, message *Message, options *Options) (*GenResponse, error)
	// GenerateStream generates a streaming response based on the provided messages and parameters.
	GenerateStream(ctx context.Context, model string, message *Message, options *Options) (<-chan *GenResponse, <-chan error)
}
