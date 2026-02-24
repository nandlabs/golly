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
	// OpenAIProviderName is the name of the OpenAI provider
	OpenAIProviderName = "openai"
	// OpenAIProviderVersion is the version of the OpenAI provider
	OpenAIProviderVersion = "1.0.0"
	// OpenAIProviderDescription is the description of the OpenAI provider
	OpenAIProviderDescription = "OpenAI API provider for GPT model inference"
	// OpenAIDefaultBaseURL is the default base URL for the OpenAI API
	OpenAIDefaultBaseURL = "https://api.openai.com/v1"
)

// OpenAIProvider implements the Provider interface for the OpenAI API
type OpenAIProvider struct {
	client       *rest.Client
	baseURL      string
	orgID        string
	models       []string
	description  string
	version      string
	extraHeaders map[string]string
}

// OpenAIProviderConfig contains configuration for the OpenAI provider
type OpenAIProviderConfig struct {
	// Auth is the authentication provider for the OpenAI API.
	// Use clients.NewBearerAuth(apiKey) for standard OpenAI API keys,
	// clients.NewAPIKeyAuth("api-key", key) for Azure OpenAI,
	// or provide a custom AuthProvider for dynamic key retrieval from secrets stores.
	Auth clients.AuthProvider
	// OrgID is the optional OpenAI organization ID
	OrgID string
	// BaseURL is the base URL for OpenAI API (default: https://api.openai.com/v1)
	BaseURL string
	// Models is the list of available models
	Models []string
	// Description is a custom description
	Description string
	// Version is a custom version
	Version string
	// ExtraHeaders are additional HTTP headers to include with every request.
	// Use this for provider-specific headers not covered by the auth mechanism.
	ExtraHeaders map[string]string
}

// --- OpenAI API request/response types ---

// openAIChatRequest represents the request structure for the OpenAI Chat Completions API
type openAIChatRequest struct {
	Model            string            `json:"model"`
	Messages         []openAIMessage   `json:"messages"`
	Stream           bool              `json:"stream"`
	StreamOptions    *streamOptions    `json:"stream_options,omitempty"`
	Temperature      *float32          `json:"temperature,omitempty"`
	TopP             *float32          `json:"top_p,omitempty"`
	N                *int              `json:"n,omitempty"`
	Stop             []string          `json:"stop,omitempty"`
	MaxTokens        *int              `json:"max_tokens,omitempty"`
	PresencePenalty  *float64          `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64          `json:"frequency_penalty,omitempty"`
	Seed             *int              `json:"seed,omitempty"`
	Tools            []openAITool      `json:"tools,omitempty"`
	ResponseFormat   *openAIRespFormat `json:"response_format,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// openAIMessage represents a message in the OpenAI format
type openAIMessage struct {
	Role       string           `json:"role"`
	Content    any              `json:"content"` // string or []openAIContentPart
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"` // for role=tool responses
	Name       string           `json:"name,omitempty"`
}

// openAIContentPart is a content part in a multi-part message
type openAIContentPart struct {
	Type     string          `json:"type"` // "text" or "image_url"
	Text     string          `json:"text,omitempty"`
	ImageURL *openAIImageURL `json:"image_url,omitempty"`
}

// openAIImageURL represents an image URL in a content part
type openAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "auto", "low", "high"
}

// openAITool represents a tool (function) definition
type openAITool struct {
	Type     string             `json:"type"` // "function"
	Function openAIToolFunction `json:"function"`
}

// openAIToolFunction represents a function definition within a tool
type openAIToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

// openAIToolCall represents a tool call returned by the model
type openAIToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // "function"
	Function openAIToolCallFunction `json:"function"`
}

// openAIToolCallFunction holds the function name and arguments in a tool call
type openAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// openAIRespFormat specifies the response format
type openAIRespFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// openAIChatResponse represents the response from the OpenAI Chat Completions API
type openAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
}

// openAIChoice represents a single choice in the response
type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	Delta        openAIMessage `json:"delta"`         // used in streaming
	FinishReason *string       `json:"finish_reason"` // "stop", "length", "tool_calls", "content_filter", null
}

// openAIUsage represents token usage information
type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// openAIErrorResponse represents an error response from the API
type openAIErrorResponse struct {
	Error openAIErrorDetail `json:"error"`
}

// openAIErrorDetail contains error details
type openAIErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// --- Constructor functions ---

// NewOpenAIProvider creates a new OpenAI provider with the given API key and REST client options.
// The API key is used as a bearer token for authentication via the rest client's AuthProvider.
func NewOpenAIProvider(apiKey string, opts *rest.ClientOpts) *OpenAIProvider {
	return NewOpenAIProviderWithConfig(&OpenAIProviderConfig{
		Auth: clients.NewBearerAuth(apiKey),
	}, opts)
}

// NewOpenAIProviderWithConfig creates a new OpenAI provider with custom configuration.
func NewOpenAIProviderWithConfig(config *OpenAIProviderConfig, opts *rest.ClientOpts) *OpenAIProvider {
	if opts == nil {
		opts = rest.CliOptsBuilder().Build()
	}
	// Set the auth provider on the client options so the rest client
	// automatically handles Authorization headers
	if config.Auth != nil {
		opts.Auth = config.Auth
	}

	client := rest.NewClientWithOptions(opts)

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = OpenAIDefaultBaseURL
	}

	description := config.Description
	if description == "" {
		description = OpenAIProviderDescription
	}

	version := config.Version
	if version == "" {
		version = OpenAIProviderVersion
	}

	return &OpenAIProvider{
		client:       client,
		baseURL:      baseURL,
		orgID:        config.OrgID,
		models:       config.Models,
		description:  description,
		version:      version,
		extraHeaders: config.ExtraHeaders,
	}
}

// --- Provider interface implementation ---

// Name returns the name of the provider.
func (o *OpenAIProvider) Name() string {
	return OpenAIProviderName
}

// Description returns a brief description of the provider.
func (o *OpenAIProvider) Description() string {
	return o.description
}

// Version returns the version of the provider.
func (o *OpenAIProvider) Version() string {
	return o.version
}

// Models returns the list of model ids supported by the provider.
func (o *OpenAIProvider) Models() []string {
	return o.models
}

// Generate generates a response based on the provided message and options.
func (o *OpenAIProvider) Generate(ctx context.Context, model string, message *genai.Message, options *genai.Options) (*genai.GenResponse, error) {
	oaiReq := o.buildRequest(model, message, options, false)

	url := fmt.Sprintf("%s/chat/completions", o.baseURL)

	req, err := o.client.NewRequest(url, http.MethodPost)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if _, err = req.WithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to set context: %w", err)
	}
	o.setHeaders(req)
	req.SetBody(oaiReq)
	req.SetContentType(ioutils.MimeApplicationJSON)

	resp, err := o.client.Execute(req)
	if err != nil {
		return nil, fmt.Errorf("openai API request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, o.parseError(resp)
	}

	var chatResp openAIChatResponse
	if err := resp.Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return o.toGenResponse(&chatResp), nil
}

// GenerateStream generates a streaming response based on the provided message and options.
func (o *OpenAIProvider) GenerateStream(ctx context.Context, model string, message *genai.Message, options *genai.Options) (<-chan *genai.GenResponse, <-chan error) {
	responseChan := make(chan *genai.GenResponse, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		oaiReq := o.buildRequest(model, message, options, true)

		url := fmt.Sprintf("%s/chat/completions", o.baseURL)

		reqBody, err := json.Marshal(oaiReq)
		if err != nil {
			errorChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := o.client.NewRequest(url, http.MethodPost)
		if err != nil {
			errorChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		if _, err = req.WithContext(ctx); err != nil {
			errorChan <- fmt.Errorf("failed to set context: %w", err)
			return
		}
		o.setHeaders(req)
		req.SeBodyReader(bytes.NewReader(reqBody))
		req.SetContentType(ioutils.MimeApplicationJSON)

		resp, err := o.client.Execute(req)
		if err != nil {
			errorChan <- fmt.Errorf("openai streaming API request failed: %w", err)
			return
		}

		httpResp := resp.Raw()
		if httpResp == nil || httpResp.Body == nil {
			errorChan <- fmt.Errorf("invalid response from openai API")
			return
		}
		defer httpResp.Body.Close()

		if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
			errorChan <- fmt.Errorf("openai API error: status %d", resp.StatusCode())
			return
		}

		// Read SSE stream
		scanner := bufio.NewScanner(httpResp.Body)
		for scanner.Scan() {
			// Check for context cancellation between chunks
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
			}

			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// SSE format: "data: {...}" or "data: [DONE]"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				break
			}

			var streamResp openAIChatResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				errorChan <- fmt.Errorf("failed to parse streaming response: %w", err)
				return
			}

			genResp := o.streamChunkToGenResponse(&streamResp)
			responseChan <- genResp
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading streaming response: %w", err)
		}
	}()

	return responseChan, errorChan
}

// Close closes the provider and releases resources.
func (o *OpenAIProvider) Close() error {
	return o.client.Close()
}

// --- Internal helpers ---

// setHeaders sets the optional org header and any extra headers on a request.
// Authorization is handled automatically by the rest client's AuthProvider.
func (o *OpenAIProvider) setHeaders(req *rest.Request) {
	if o.orgID != "" {
		req.AddHeader("OpenAI-Organization", o.orgID)
	}
	for k, v := range o.extraHeaders {
		req.AddHeader(k, v)
	}
}

// parseError extracts an error message from a failed API response.
func (o *OpenAIProvider) parseError(resp *rest.Response) error {
	var errResp openAIErrorResponse
	if err := resp.Decode(&errResp); err != nil {
		return fmt.Errorf("openai API error: status %d", resp.StatusCode())
	}
	return fmt.Errorf("openai API error [%s]: %s", errResp.Error.Type, errResp.Error.Message)
}

// buildRequest builds an OpenAI chat completions request from genai types.
func (o *OpenAIProvider) buildRequest(model string, message *genai.Message, options *genai.Options, stream bool) *openAIChatRequest {
	req := &openAIChatRequest{
		Model:    model,
		Messages: o.convertMessages(message, options),
		Stream:   stream,
	}

	if stream {
		req.StreamOptions = &streamOptions{IncludeUsage: true}
	}

	if options != nil {
		if options.Has(genai.OptionTemperature) {
			v := options.GetFloat32(genai.OptionTemperature)
			req.Temperature = &v
		}
		if options.Has(genai.OptionTopP) {
			v := options.GetFloat32(genai.OptionTopP)
			req.TopP = &v
		}
		if options.Has(genai.OptionMaxTokens) {
			v := options.GetInt(genai.OptionMaxTokens)
			req.MaxTokens = &v
		}
		if options.Has(genai.OptionCandidateCount) {
			v := options.GetInt(genai.OptionCandidateCount)
			req.N = &v
		}
		if options.Has(genai.OptionSeed) {
			v := options.GetInt(genai.OptionSeed)
			req.Seed = &v
		}
		if options.Has(genai.OptionFrequencyPenalty) {
			v := options.GetFloat64(genai.OptionFrequencyPenalty)
			req.FrequencyPenalty = &v
		}
		if options.Has(genai.OptionPresencePenalty) {
			v := options.GetFloat64(genai.OptionPresencePenalty)
			req.PresencePenalty = &v
		}
		if stopWords := options.GetStrings(genai.OptionStopWords); len(stopWords) > 0 {
			req.Stop = stopWords
		}
		if outputMime := options.GetString(genai.OptionOutputMime); outputMime == ioutils.MimeApplicationJSON {
			req.ResponseFormat = &openAIRespFormat{Type: "json_object"}
		}
	}

	return req
}

// convertMessages builds the OpenAI messages array from a genai.Message.
// It prepends system instructions if present in options.
func (o *OpenAIProvider) convertMessages(message *genai.Message, options *genai.Options) []openAIMessage {
	var messages []openAIMessage

	// Add system instructions if present
	if options != nil {
		if sysInstr := options.GetString(genai.OptionSystemInstructions); sysInstr != "" {
			messages = append(messages, openAIMessage{
				Role:    "system",
				Content: sysInstr,
			})
		}
	}

	oaiMsg := openAIMessage{
		Role: o.convertRole(message.Role),
	}

	var textParts []string
	var contentParts []openAIContentPart
	hasMultiModal := false

	for _, part := range message.Parts {
		switch {
		case part.Text != nil:
			textParts = append(textParts, part.Text.Text)
			contentParts = append(contentParts, openAIContentPart{
				Type: "text",
				Text: part.Text.Text,
			})

		case part.File != nil && ioutils.IsImageMime(part.MimeType):
			hasMultiModal = true
			contentParts = append(contentParts, openAIContentPart{
				Type: "image_url",
				ImageURL: &openAIImageURL{
					URL: part.File.URI,
				},
			})

		case part.Bin != nil && ioutils.IsImageMime(part.MimeType):
			hasMultiModal = true
			dataURI := fmt.Sprintf("data:%s;base64,%s", part.MimeType, base64.StdEncoding.EncodeToString(part.Bin.Data))
			contentParts = append(contentParts, openAIContentPart{
				Type: "image_url",
				ImageURL: &openAIImageURL{
					URL: dataURI,
				},
			})

		case part.FuncCall != nil:
			argsJSON, _ := json.Marshal(part.FuncCall.Arguments)
			oaiMsg.ToolCalls = append(oaiMsg.ToolCalls, openAIToolCall{
				ID:   part.FuncCall.Id,
				Type: "function",
				Function: openAIToolCallFunction{
					Name:      part.FuncCall.FunctionName,
					Arguments: string(argsJSON),
				},
			})

		case part.FuncResponse != nil:
			// Function responses are sent as separate tool messages
			toolMsg := openAIMessage{
				Role:       "tool",
				ToolCallID: part.Name,
			}
			if part.FuncResponse.Text != nil {
				toolMsg.Content = *part.FuncResponse.Text
			} else if part.FuncResponse.Data != nil {
				toolMsg.Content = string(part.FuncResponse.Data)
			}
			messages = append(messages, toolMsg)
		}
	}

	// Use multi-part content if we have images, otherwise plain string
	if hasMultiModal {
		oaiMsg.Content = contentParts
	} else if len(textParts) > 0 {
		oaiMsg.Content = strings.Join(textParts, "\n")
	}

	messages = append(messages, oaiMsg)

	return messages
}

// convertRole maps genai.Role to OpenAI role strings.
func (o *OpenAIProvider) convertRole(role genai.Role) string {
	switch role {
	case genai.RoleSystem:
		return "system"
	case genai.RoleUser:
		return "user"
	case genai.RoleAssistant:
		return "assistant"
	default:
		return "user"
	}
}

// convertFromOpenAIRole maps OpenAI role strings to genai.Role.
func (o *OpenAIProvider) convertFromOpenAIRole(role string) genai.Role {
	switch role {
	case "system":
		return genai.RoleSystem
	case "user":
		return genai.RoleUser
	case "assistant":
		return genai.RoleAssistant
	default:
		return genai.RoleUser
	}
}

// mapFinishReason maps an OpenAI finish_reason to genai.FinishReason.
func (o *OpenAIProvider) mapFinishReason(reason *string) genai.FinishReason {
	if reason == nil {
		return genai.FinishReasonInProgress
	}
	switch *reason {
	case "stop":
		return genai.FinishReasonStop
	case "length":
		return genai.FinishReasonLength
	case "tool_calls":
		return genai.FinishReasonToolCall
	case "content_filter":
		return genai.FinishReasonContentFilter
	default:
		return genai.FinishReasonUnknown
	}
}

// toGenResponse converts a full (non-streaming) OpenAI response to a genai.GenResponse.
func (o *OpenAIProvider) toGenResponse(resp *openAIChatResponse) *genai.GenResponse {
	genResp := &genai.GenResponse{
		Candidates: make([]genai.Candidate, 0, len(resp.Choices)),
	}

	for _, choice := range resp.Choices {
		msg := o.openAIMsgToGenMessage(&choice.Message)
		genResp.Candidates = append(genResp.Candidates, genai.Candidate{
			Index:        choice.Index,
			Message:      msg,
			FinishReason: o.mapFinishReason(choice.FinishReason),
		})
	}

	if resp.Usage != nil {
		genResp.Meta = genai.ResponseMeta{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		}
	}

	return genResp
}

// streamChunkToGenResponse converts a streaming chunk to a genai.GenResponse.
func (o *OpenAIProvider) streamChunkToGenResponse(resp *openAIChatResponse) *genai.GenResponse {
	genResp := &genai.GenResponse{
		Candidates: make([]genai.Candidate, 0, len(resp.Choices)),
	}

	for _, choice := range resp.Choices {
		msg := o.openAIMsgToGenMessage(&choice.Delta)
		genResp.Candidates = append(genResp.Candidates, genai.Candidate{
			Index:        choice.Index,
			Message:      msg,
			FinishReason: o.mapFinishReason(choice.FinishReason),
		})
	}

	if resp.Usage != nil {
		genResp.Meta = genai.ResponseMeta{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		}
	}

	return genResp
}

// openAIMsgToGenMessage converts an OpenAI message to a genai.Message.
func (o *OpenAIProvider) openAIMsgToGenMessage(msg *openAIMessage) *genai.Message {
	message := &genai.Message{
		Role:  o.convertFromOpenAIRole(msg.Role),
		Parts: []genai.Part{},
	}

	// Content can be a string or nil (e.g. when tool_calls are present)
	switch c := msg.Content.(type) {
	case string:
		if c != "" {
			message.Parts = append(message.Parts, genai.Part{
				Name:     "text",
				MimeType: ioutils.MimeTextPlain,
				Text: &genai.TextPart{
					Text: c,
				},
			})
		}
	}

	// Convert tool calls
	for _, tc := range msg.ToolCalls {
		args := make(map[string]interface{})
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

		message.Parts = append(message.Parts, genai.Part{
			Name: tc.Function.Name,
			FuncCall: &genai.FuncCallPart{
				Id:           tc.ID,
				FunctionName: tc.Function.Name,
				Arguments:    args,
			},
		})
	}

	return message
}
