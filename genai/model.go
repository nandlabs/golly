package genai

import (
	"io"
)

// Model is the interface that represents a generative AI model
type Model interface {
	// Name returns the name of the model
	Name() string
	// Description returns a description of the model
	Description() string
	// Version returns the version of the model
	Version() string
	// Author returns the author of the model
	Author() string
	// License returns the license of the model
	License() string
	// Supports returns true if the model supports the given MIME type as a consumer or provider
	Supports(mime string) (consumer bool, provider bool)
	// Accepts returns the supported input MIME types accepted by this model
	Accepts() []string
	// Produces returns the supported output MIME types thats produced by this model
	Produces() []string
	// Generate will invoke the model generation and fetch the result
	Generate(exchange Exchange) error
	// GenerateStream will invoke the model generation and stream the result
	GenerateStream(exchnage Exchange) error
}

// Options is a set of options for the model
type Options struct {

	//MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens" yaml:"max_tokens"`
	//MinTokens is the minimum number of tokens to generate.
	MinTokens int `json:"min_tokens"  yaml:"min_tokens"`
	//Temperature is the model temperature, parameter that regulates the randomness, or creativity, of the AI's responses.
	Temperature float32 `json:"temperature"  yaml:"temperature"`
	//CandidateCount is the number of response candidates to generate.
	CandidateCount int `json:"candidate_count" yaml:"candidate_count"`
	//StopWords is a list of words to stop generation on.
	StopWords []string `json:"stop_words" yaml:"stop_words"`
	//TopK is the number of tokens to consider for the top-k sampling selection.
	TopK int `json:"top_k" yaml:"top_k"`
	//TopP is the cumulative probability of the top-p sampling selection.
	TopP float64 `json:"top_p" yaml:"top_p"`
	//Seed is the seed to use for random generation.
	Seed int `json:"seed" yaml:"seed"`
	//MinLength is the minimum length of the generated text.
	MinLength int `json:"min_length" yaml:"min_length"`
	//MaxLength is the maximum length of the generated text.
	MaxLength int `json:"max_length" yaml:"max_length"`
	//RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty float64 `json:"repetition_penalty" yaml:"repetition_penalty"`
	//FrequencyPenalty is the frequency penalty for sampling.
	FrequencyPenalty float64 `json:"frequency_penalty" yaml:"frequency_penalty"`
	//PresencePenalty is the presence penalty for sampling.
	PresencePenalty float64 `json:"presence_penalty" yaml:"presence_penalty"`
	//StreamHandler is the handler for streaming responses
	StreamHandler func(reader io.Reader) error
}

// SetMaxTokens sets the maximum number of tokens to generate.
func (o *Options) SetMaxTokens(maxTokens int) *Options {
	o.MaxTokens = maxTokens
	return o
}

// SetMinTokens sets the minimum number of tokens to generate.
func (o *Options) SetMinTokens(minTokens int) *Options {
	o.MinTokens = minTokens
	return o
}

// SetTemperature sets the model temperature, a hyperparameter that
// regulates the randomness, or creativity, of the AI's responses.
func (o *Options) SetTemperature(temperature float32) *Options {
	o.Temperature = temperature
	return o
}

// SetCandidateCount sets the number of response candidates to generate.
func (o *Options) SetCandidateCount(c int) *Options {

	o.CandidateCount = c
	return o
}

// SetStopWords sets a list of words to stop generation on.
func (o *Options) SetStopWords(stopWords ...string) *Options {
	o.StopWords = stopWords
	return o
}

// SetTopK sets the number of tokens to consider for the top-k sampling selection.
func (o *Options) SetTopK(topK int) *Options {
	o.TopK = topK
	return o
}

// SetTopP sets the cumulative probability of the top-p sampling selection.
func (o *Options) SetTopP(topP float64) *Options {
	o.TopP = topP
	return o
}

// SetSeed sets the seed to use for random generation.
func (o *Options) SetSeed(seed int) *Options {
	o.Seed = seed
	return o
}

// SetMinLength sets the minimum length of the generated text.
func (o *Options) SetMinLength(minLength int) *Options {
	o.MinLength = minLength
	return o
}

// SetMaxLength sets the maximum length of the generated text.
func (o *Options) SetMaxLength(maxLength int) *Options {
	o.MaxLength = maxLength
	return o
}

// SetRepetitionPenalty sets the repetition penalty for sampling.
func (o *Options) SetRepetitionPenalty(repetitionPenalty float64) *Options {
	o.RepetitionPenalty = repetitionPenalty
	return o
}

// SetFrequencyPenalty sets the frequency penalty for sampling.
func (o *Options) SetFrequencyPenalty(frequencyPenalty float64) *Options {
	o.FrequencyPenalty = frequencyPenalty
	return o
}

// SetPresencePenalty sets the presence penalty for sampling.
func (o *Options) SetPresencePenalty(presencePenalty float64) *Options {
	o.PresencePenalty = presencePenalty
	return o
}

// SetStreamHandler sets the streaming function to use.
func (o *Options) SetStreamHandler(streamHandler func(reader io.Reader) error) *Options {
	o.StreamHandler = streamHandler
	return o
}

// AbstractModel is a simple implementation of the Model interface
type AbstractModel struct {
	name        string
	description string
	version     string
	author      string
	license     string
	inputMime   []string
	outputMime  []string
	options     *Options
}

// Name returns the name of the model
func (m *AbstractModel) Name() string {
	return m.name
}

// Description returns a description of the model
func (m *AbstractModel) Description() string {
	return m.description
}

// Version returns the version of the model
func (m *AbstractModel) Version() string {
	return m.version
}

// Author returns the author of the model
func (m *AbstractModel) Author() string {
	return m.author
}

// License returns the license of the model
func (m *AbstractModel) License() string {
	return m.license
}

// InputMimeTypes returns the supported input MIME types
func (m *AbstractModel) InputMimeTypes() []string {
	return m.inputMime
}

// OutputMimeTypes returns the supported output MIME types
func (m *AbstractModel) OutputMimeTypes() []string {
	return m.outputMime
}

// Options returns the options for the model
func (m *AbstractModel) Options() *Options {
	return m.options
}
