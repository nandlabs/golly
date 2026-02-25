package impl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"oss.nandlabs.io/golly/clients"
	"oss.nandlabs.io/golly/genai"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/rest"
)

const (
	// ClaudeProviderName is the name of the Claude provider
	ClaudeProviderName = "claude"
	// ClaudeProviderVersion is the version of the Claude provider
	ClaudeProviderVersion = "1.0.0"
	// ClaudeProviderDescription is the description of the Claude provider
	ClaudeProviderDescription = "Anthropic Claude API provider for Claude model inference"
	// ClaudeDefaultBaseURL is the default base URL for the Anthropic API
	ClaudeDefaultBaseURL = "https://api.anthropic.com"
	// ClaudeDefaultAPIVersion is the default Anthropic API version header
	ClaudeDefaultAPIVersion = "2023-06-01"
	// ClaudeDefaultMaxTokens is the default max tokens if not specified
	ClaudeDefaultMaxTokens = 4096
)

// ClaudeProvider implements the Provider interface for the Anthropic Claude API.
type ClaudeProvider struct {
	client       *rest.Client
	baseURL      string
	apiVersion   string
	models       []string
	description  string
	version      string
	extraHeaders map[string]string
}

// ClaudeProviderConfig contains configuration for the Claude provider.
type ClaudeProviderConfig struct {
	// Auth is the authentication provider for the Claude API.
	// Anthropic uses x-api-key header-based auth, so use:
	//   clients.NewAPIKeyAuth("x-api-key", apiKey)
	// For secrets-store-backed keys, implement a custom AuthProvider
	// that returns AuthTypeAPIKey and fetches the key dynamically.
	Auth clients.AuthProvider
	// APIVersion is the Anthropic API version header (default: "2023-06-01").
	APIVersion string
	// BaseURL is the base URL for the Anthropic API (default: https://api.anthropic.com).
	BaseURL string
	// Models is the list of available models.
	Models []string
	// Description is a custom description.
	Description string
	// Version is a custom version.
	Version string
	// ExtraHeaders are additional HTTP headers to include with every request.
	ExtraHeaders map[string]string
}

// --- Anthropic API request/response types ---

// claudeRequest represents the request structure for the Anthropic Messages API.
type claudeRequest struct {
	Model         string          `json:"model"`
	Messages      []claudeMessage `json:"messages"`
	System        string          `json:"system,omitempty"`
	MaxTokens     int             `json:"max_tokens"`
	Stream        bool            `json:"stream,omitempty"`
	Temperature   *float32        `json:"temperature,omitempty"`
	TopP          *float32        `json:"top_p,omitempty"`
	TopK          *int            `json:"top_k,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Tools         []claudeTool    `json:"tools,omitempty"`
	Metadata      *claudeMetadata `json:"metadata,omitempty"`
}

// claudeMessage represents a message in the Anthropic format.
type claudeMessage struct {
	Role    string               `json:"role"`    // "user" or "assistant"
	Content []claudeContentBlock `json:"content"` // always an array of content blocks
}

// claudeContentBlock is a content block within a message.
type claudeContentBlock struct {
	Type   string             `json:"type"`             // "text", "image", "tool_use", "tool_result"
	Text   string             `json:"text,omitempty"`   // for type="text"
	Source *claudeImageSource `json:"source,omitempty"` // for type="image"
	// Tool use fields
	ID    string `json:"id,omitempty"`    // for type="tool_use"
	Name  string `json:"name,omitempty"`  // for type="tool_use"
	Input any    `json:"input,omitempty"` // for type="tool_use"
	// Tool result fields
	ToolUseID string `json:"tool_use_id,omitempty"` // for type="tool_result"
	Content   any    `json:"content,omitempty"`     // for type="tool_result" (string or []claudeContentBlock)
}

// claudeImageSource represents the source data for an inline image.
type claudeImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // e.g. "image/png"
	Data      string `json:"data"`       // base64-encoded image data
}

// claudeTool represents a tool definition for Claude.
type claudeTool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"input_schema"`
}

// claudeMetadata holds optional request metadata.
type claudeMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// claudeResponse represents the response from the Anthropic Messages API.
type claudeResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"` // "message"
	Role         string               `json:"role"` // "assistant"
	Content      []claudeContentBlock `json:"content"`
	Model        string               `json:"model"`
	StopReason   *string              `json:"stop_reason"` // "end_turn", "max_tokens", "stop_sequence", "tool_use"
	StopSequence *string              `json:"stop_sequence,omitempty"`
	Usage        claudeUsage          `json:"usage"`
}

// claudeUsage represents token usage information.
type claudeUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// claudeErrorResponse represents an error response from the Anthropic API.
type claudeErrorResponse struct {
	Type  string            `json:"type"` // "error"
	Error claudeErrorDetail `json:"error"`
}

// claudeErrorDetail contains error details.
type claudeErrorDetail struct {
	Type    string `json:"type"` // "invalid_request_error", "authentication_error", etc.
	Message string `json:"message"`
}

// --- Streaming event types ---

// claudeStreamEvent represents a streaming SSE event from the Anthropic API.
type claudeStreamEvent struct {
	Type         string              `json:"type"`
	Index        int                 `json:"index,omitempty"`
	ContentBlock *claudeContentBlock `json:"content_block,omitempty"`
	Delta        *claudeStreamDelta  `json:"delta,omitempty"`
	Message      *claudeResponse     `json:"message,omitempty"`
	Usage        *claudeUsage        `json:"usage,omitempty"`
}

// claudeStreamDelta represents the delta content in a streaming event.
type claudeStreamDelta struct {
	Type         string  `json:"type,omitempty"`          // "text_delta", "input_json_delta"
	Text         string  `json:"text,omitempty"`          // for text_delta
	PartialJSON  string  `json:"partial_json,omitempty"`  // for input_json_delta
	StopReason   *string `json:"stop_reason,omitempty"`   // for message_delta
	StopSequence *string `json:"stop_sequence,omitempty"` // for message_delta
}

// --- Constructor functions ---

// NewClaudeProvider creates a new Claude provider with the given API key and REST client options.
// The API key is sent via the x-api-key header as required by the Anthropic API.
func NewClaudeProvider(apiKey string, opts *rest.ClientOpts) *ClaudeProvider {
	return NewClaudeProviderWithConfig(&ClaudeProviderConfig{
		Auth: clients.NewAPIKeyAuth("x-api-key", apiKey),
	}, opts)
}

// NewClaudeProviderWithConfig creates a new Claude provider with custom configuration.
func NewClaudeProviderWithConfig(config *ClaudeProviderConfig, opts *rest.ClientOpts) *ClaudeProvider {
	if opts == nil {
		opts = rest.CliOptsBuilder().Build()
	}
	// Set the auth provider on the client options so the rest client
	// automatically handles the x-api-key header.
	if config.Auth != nil {
		opts.Auth = config.Auth
	}

	client := rest.NewClientWithOptions(opts)

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ClaudeDefaultBaseURL
	}

	apiVersion := config.APIVersion
	if apiVersion == "" {
		apiVersion = ClaudeDefaultAPIVersion
	}

	description := config.Description
	if description == "" {
		description = ClaudeProviderDescription
	}

	version := config.Version
	if version == "" {
		version = ClaudeProviderVersion
	}

	return &ClaudeProvider{
		client:       client,
		baseURL:      baseURL,
		apiVersion:   apiVersion,
		models:       config.Models,
		description:  description,
		version:      version,
		extraHeaders: config.ExtraHeaders,
	}
}

// --- Provider interface implementation ---

// Name returns the name of the provider.
func (c *ClaudeProvider) Name() string {
	return ClaudeProviderName
}

// Description returns a brief description of the provider.
func (c *ClaudeProvider) Description() string {
	return c.description
}

// Version returns the version of the provider.
func (c *ClaudeProvider) Version() string {
	return c.version
}

// Models returns the list of model ids supported by the provider.
func (c *ClaudeProvider) Models() []string {
	return c.models
}

// Generate generates a response based on the provided message and options.
func (c *ClaudeProvider) Generate(ctx context.Context, model string, message *genai.Message, options *genai.Options) (*genai.GenResponse, error) {
	claudeReq := c.buildRequest(model, message, options, false)

	url := fmt.Sprintf("%s/v1/messages", c.baseURL)

	req, err := c.client.NewRequest(url, http.MethodPost)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if _, err = req.WithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to set context: %w", err)
	}
	c.setHeaders(req)
	req.SetBody(claudeReq)
	req.SetContentType(ioutils.MimeApplicationJSON)

	resp, err := c.client.Execute(req)
	if err != nil {
		return nil, fmt.Errorf("claude API request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, c.parseError(resp)
	}

	var claudeResp claudeResponse
	if err := resp.Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.toGenResponse(&claudeResp), nil
}

// GenerateStream generates a streaming response based on the provided message and options.
func (c *ClaudeProvider) GenerateStream(ctx context.Context, model string, message *genai.Message, options *genai.Options) (<-chan *genai.GenResponse, <-chan error) {
	responseChan := make(chan *genai.GenResponse, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		claudeReq := c.buildRequest(model, message, options, true)

		url := fmt.Sprintf("%s/v1/messages", c.baseURL)

		reqBody, err := json.Marshal(claudeReq)
		if err != nil {
			errorChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := c.client.NewRequest(url, http.MethodPost)
		if err != nil {
			errorChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		if _, err = req.WithContext(ctx); err != nil {
			errorChan <- fmt.Errorf("failed to set context: %w", err)
			return
		}
		c.setHeaders(req)
		req.SeBodyReader(bytes.NewReader(reqBody))
		req.SetContentType(ioutils.MimeApplicationJSON)

		resp, err := c.client.Execute(req)
		if err != nil {
			errorChan <- fmt.Errorf("claude streaming API request failed: %w", err)
			return
		}

		httpResp := resp.Raw()
		if httpResp == nil || httpResp.Body == nil {
			errorChan <- fmt.Errorf("invalid response from claude API")
			return
		}
		defer httpResp.Body.Close()

		if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
			errorChan <- fmt.Errorf("claude API error: status %d", resp.StatusCode())
			return
		}

		// Read SSE stream — Anthropic uses "event:" + "data:" lines
		scanner := bufio.NewScanner(httpResp.Body)
		var currentEventType string

		for scanner.Scan() {
			// Check for context cancellation between chunks
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
			}

			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Parse event type
			if strings.HasPrefix(line, "event: ") {
				currentEventType = strings.TrimPrefix(line, "event: ")
				continue
			}

			// Parse data
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			switch currentEventType {
			case "content_block_delta":
				var event claudeStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					errorChan <- fmt.Errorf("failed to parse streaming event: %w", err)
					return
				}
				if event.Delta != nil && event.Delta.Text != "" {
					genResp := c.deltaToGenResponse(event.Delta.Text, event.Index)
					responseChan <- genResp
				}

			case "message_delta":
				var event claudeStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					errorChan <- fmt.Errorf("failed to parse message_delta: %w", err)
					return
				}
				genResp := &genai.GenResponse{
					Candidates: []genai.Candidate{
						{
							Index:        0,
							Message:      genai.NewTextMessage(genai.RoleAssistant, ""),
							FinishReason: c.mapStopReason(event.Delta.StopReason),
						},
					},
				}
				if event.Usage != nil {
					genResp.Meta = genai.ResponseMeta{
						OutputTokens: event.Usage.OutputTokens,
					}
				}
				responseChan <- genResp

			case "message_start":
				var event claudeStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					errorChan <- fmt.Errorf("failed to parse message_start: %w", err)
					return
				}
				if event.Message != nil && event.Message.Usage.InputTokens > 0 {
					genResp := &genai.GenResponse{
						Meta: genai.ResponseMeta{
							InputTokens:  event.Message.Usage.InputTokens,
							CachedTokens: event.Message.Usage.CacheReadInputTokens,
						},
					}
					responseChan <- genResp
				}

			case "message_stop":
				// Stream complete
				return

			case "error":
				errorChan <- fmt.Errorf("claude streaming error: %s", data)
				return

			// content_block_start, content_block_stop, ping — ignored
			default:
				continue
			}
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading streaming response: %w", err)
		}
	}()

	return responseChan, errorChan
}

// Close closes the provider and releases resources.
func (c *ClaudeProvider) Close() error {
	return c.client.Close()
}

// --- Internal helpers ---

// setHeaders sets the required Anthropic headers and any extra headers on a request.
// Authorization (x-api-key) is handled automatically by the rest client's AuthProvider.
func (c *ClaudeProvider) setHeaders(req *rest.Request) {
	req.AddHeader("anthropic-version", c.apiVersion)
	for k, v := range c.extraHeaders {
		req.AddHeader(k, v)
	}
}

// parseError extracts an error message from a failed API response.
func (c *ClaudeProvider) parseError(resp *rest.Response) error {
	var errResp claudeErrorResponse
	if err := resp.Decode(&errResp); err != nil {
		return fmt.Errorf("claude API error: status %d", resp.StatusCode())
	}
	return fmt.Errorf("claude API error [%s]: %s", errResp.Error.Type, errResp.Error.Message)
}

// buildRequest builds an Anthropic Messages API request from genai types.
func (c *ClaudeProvider) buildRequest(model string, message *genai.Message, options *genai.Options, stream bool) *claudeRequest {
	req := &claudeRequest{
		Model:     model,
		MaxTokens: ClaudeDefaultMaxTokens,
	}

	if stream {
		req.Stream = true
	}

	// Extract system instructions (Anthropic requires system as top-level field)
	if options != nil {
		if sysInstr := options.GetString(genai.OptionSystemInstructions); sysInstr != "" {
			req.System = sysInstr
		}
		if options.Has(genai.OptionMaxTokens) {
			req.MaxTokens = options.GetInt(genai.OptionMaxTokens)
		}
		if options.Has(genai.OptionTemperature) {
			v := options.GetFloat32(genai.OptionTemperature)
			req.Temperature = &v
		}
		if options.Has(genai.OptionTopP) {
			v := options.GetFloat32(genai.OptionTopP)
			req.TopP = &v
		}
		if options.Has(genai.OptionTopK) {
			v := options.GetInt(genai.OptionTopK)
			req.TopK = &v
		}
		if stopWords := options.GetStrings(genai.OptionStopWords); len(stopWords) > 0 {
			req.StopSequences = stopWords
		}
	}

	req.Messages = c.convertMessages(message)

	return req
}

// convertMessages builds the Anthropic messages array from a genai.Message.
// Anthropic only supports "user" and "assistant" roles in messages.
// System messages are handled separately via the top-level "system" field.
func (c *ClaudeProvider) convertMessages(message *genai.Message) []claudeMessage {
	var messages []claudeMessage

	// Skip system messages — they are extracted by buildRequest
	if message.Role == genai.RoleSystem {
		return messages
	}

	claudeMsg := claudeMessage{
		Role: c.convertRole(message.Role),
	}

	for _, part := range message.Parts {
		switch {
		case part.Text != nil:
			claudeMsg.Content = append(claudeMsg.Content, claudeContentBlock{
				Type: "text",
				Text: part.Text.Text,
			})

		case part.Bin != nil && ioutils.IsImageMime(part.MimeType):
			claudeMsg.Content = append(claudeMsg.Content, claudeContentBlock{
				Type: "image",
				Source: &claudeImageSource{
					Type:      "base64",
					MediaType: part.MimeType,
					Data:      base64.StdEncoding.EncodeToString(part.Bin.Data),
				},
			})

		case part.File != nil && ioutils.IsImageMime(part.MimeType):
			// Anthropic doesn't support image URLs directly — log a warning.
			// For URL-based images, the caller should download and pass as BinPart.
			claudeMsg.Content = append(claudeMsg.Content, claudeContentBlock{
				Type: "text",
				Text: fmt.Sprintf("[Image URL not directly supported by Claude API: %s]", part.File.URI),
			})

		case part.FuncCall != nil:
			claudeMsg.Content = append(claudeMsg.Content, claudeContentBlock{
				Type:  "tool_use",
				ID:    part.FuncCall.Id,
				Name:  part.FuncCall.FunctionName,
				Input: part.FuncCall.Arguments,
			})

		case part.FuncResponse != nil:
			var resultContent any
			if part.FuncResponse.Text != nil {
				resultContent = *part.FuncResponse.Text
			} else if part.FuncResponse.Data != nil {
				resultContent = string(part.FuncResponse.Data)
			}
			claudeMsg.Content = append(claudeMsg.Content, claudeContentBlock{
				Type:      "tool_result",
				ToolUseID: part.Name,
				Content:   resultContent,
			})
		}
	}

	// Ensure at least one content block exists
	if len(claudeMsg.Content) == 0 {
		claudeMsg.Content = append(claudeMsg.Content, claudeContentBlock{
			Type: "text",
			Text: "",
		})
	}

	messages = append(messages, claudeMsg)

	return messages
}

// convertRole maps genai.Role to Anthropic role strings.
// Anthropic only supports "user" and "assistant" roles in messages.
func (c *ClaudeProvider) convertRole(role genai.Role) string {
	switch role {
	case genai.RoleUser:
		return "user"
	case genai.RoleAssistant:
		return "assistant"
	default:
		return "user"
	}
}

// convertFromClaudeRole maps Anthropic role strings to genai.Role.
func (c *ClaudeProvider) convertFromClaudeRole(role string) genai.Role {
	switch role {
	case "user":
		return genai.RoleUser
	case "assistant":
		return genai.RoleAssistant
	default:
		return genai.RoleAssistant
	}
}

// mapStopReason maps an Anthropic stop_reason to genai.FinishReason.
func (c *ClaudeProvider) mapStopReason(reason *string) genai.FinishReason {
	if reason == nil {
		return genai.FinishReasonInProgress
	}
	switch *reason {
	case "end_turn":
		return genai.FinishReasonEndTurn
	case "max_tokens":
		return genai.FinishReasonLength
	case "stop_sequence":
		return genai.FinishReasonStop
	case "tool_use":
		return genai.FinishReasonToolCall
	default:
		return genai.FinishReasonUnknown
	}
}

// toGenResponse converts a full (non-streaming) Claude response to a genai.GenResponse.
func (c *ClaudeProvider) toGenResponse(resp *claudeResponse) *genai.GenResponse {
	msg := c.claudeContentToGenMessage(resp.Role, resp.Content)

	genResp := &genai.GenResponse{
		Candidates: []genai.Candidate{
			{
				Index:        0,
				Message:      msg,
				FinishReason: c.mapStopReason(resp.StopReason),
			},
		},
		Meta: genai.ResponseMeta{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
			CachedTokens: resp.Usage.CacheReadInputTokens,
		},
	}

	return genResp
}

// deltaToGenResponse creates a GenResponse from a streaming text delta.
func (c *ClaudeProvider) deltaToGenResponse(text string, index int) *genai.GenResponse {
	return &genai.GenResponse{
		Candidates: []genai.Candidate{
			{
				Index:        index,
				Message:      genai.NewTextMessage(genai.RoleAssistant, text),
				FinishReason: genai.FinishReasonInProgress,
			},
		},
	}
}

// claudeContentToGenMessage converts Claude content blocks to a genai.Message.
func (c *ClaudeProvider) claudeContentToGenMessage(role string, content []claudeContentBlock) *genai.Message {
	message := &genai.Message{
		Role:  c.convertFromClaudeRole(role),
		Parts: []genai.Part{},
	}

	for _, block := range content {
		switch block.Type {
		case "text":
			if block.Text != "" {
				message.Parts = append(message.Parts, genai.Part{
					Name:     "text",
					MimeType: ioutils.MimeTextPlain,
					Text: &genai.TextPart{
						Text: block.Text,
					},
				})
			}

		case "tool_use":
			args := make(map[string]interface{})
			if block.Input != nil {
				if inputMap, ok := block.Input.(map[string]interface{}); ok {
					args = inputMap
				} else {
					// Try marshaling and unmarshaling for nested types
					if b, err := json.Marshal(block.Input); err == nil {
						_ = json.Unmarshal(b, &args)
					}
				}
			}
			message.Parts = append(message.Parts, genai.Part{
				Name: block.Name,
				FuncCall: &genai.FuncCallPart{
					Id:           block.ID,
					FunctionName: block.Name,
					Arguments:    args,
				},
			})
		}
	}

	return message
}
