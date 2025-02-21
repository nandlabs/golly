package genai

import (
	"io"

	"oss.nandlabs.io/golly/errutils"
	"oss.nandlabs.io/golly/managers"
)

const (
	OptionMaxTokens         = "max_tokens"
	OptionMinTokens         = "min_tokens"
	OptionTemperature       = "temperature"
	OptionTopP              = "top_p"
	OptionCandidateCount    = "candidate_count"
	OptionStopWords         = "stop_words"
	OptionTopK              = "top_k"
	OptionSeed              = "seed"
	OptionMinLength         = "min_length"
	OptionMaxLength         = "max_length"
	OptionRepetitionPenalty = "repetition_penalty"
	OptionFrequencyPenalty  = "frequency_penalty"
	OptionPresencePenalty   = "presence_penalty"
	OptionBestOf            = "best_of"
	OptionLogProbs          = "log_probs"
	OptionEcho              = "echo"
	OptionOutputMimes       = "output_mimes"
)

var GetUnsupportedModelErr = errutils.NewCustomError("unsupported model %s")
var GetUnsupportedMimeErr = errutils.NewCustomError("unsupported mime type %s for model %s")
var GetUnsupportedConsumerErr = errutils.NewCustomError("unsupported consumer for model %s")
var GetUnsupportedProviderErr = errutils.NewCustomError("unsupported provider for model %s")

var Providers managers.ItemManager[Provider] = managers.NewItemManager[Provider]()

// Provider is the interface that represents a generative AI model
type Provider interface {
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
	Supports(model, mime string) (consumer bool, provider bool)
	// Accepts returns the supported input MIME types accepted by this model
	Accepts(model string) []string
	// Produces returns the supported output MIME types thats produced by this model
	Produces(model string) []string
	// Generate will invoke the model generation and fetch the result. This is a blocking call.
	Generate(model string, exchange Exchange, options *Options) error
	// GenerateStream will invoke the model generation and stream the result. This will be a non-blocking call.
	GenerateStream(model string, exchange Exchange, handler func(reader io.Reader), options Options) error
}

// Options represents the options for the Service
type Options struct {
	values map[string]any
}

// OptionsBuilder is a builder for the Options struct

func (o *Options) Get(name string) any {
	return o.values[name]
}

// Set sets the value of the option with the specified name.
// If the option already exists, it will be overwritten with the new value.
//
// Parameters:
//
//	name (string): The name of the option to set.
//	value (any): The value to set for the option.
func (o *Options) Set(name string, value any) {
	o.values[name] = value
}

// Remove removes the option with the specified name from the Options.
//
// Parameters:
//
//	name (string): The name of the option to remove.
func (o *Options) Remove(name string) {
	delete(o.values, name)
}

// GetString returns the string value of the option with the specified name.
// If the option does not exist or is not a string, an empty string is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	string: The string value of the option.

func (o *Options) GetString(name string) string {
	if v, ok := o.values[name].(string); ok {
		return v
	}
	return ""
}

// GetInt returns the integer value of the option with the specified name.
// If the option does not exist or is not an integer, 0 is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	int: The integer value of the option.

func (o *Options) GetInt(name string) int {
	if v, ok := o.values[name].(int); ok {
		return v
	}
	return 0
}

// GetBool returns the boolean value of the option with the specified name.
// If the option does not exist or is not a boolean, false is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	bool: The boolean value of the option.

func (o *Options) GetBool(name string) bool {
	if v, ok := o.values[name].(bool); ok {
		return v
	}
	return false
}

// GetStrings returns the string slice value of the option with the specified name.
// If the option does not exist or is not a string slice, nil is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	[]string: The string slice value of the option.
func (o *Options) GetStrings(name string) []string {
	if v, ok := o.values[name].([]string); ok {
		return v
	}
	return nil
}

// GetFloat32 returns the float32 value of the option with the specified name.
// If the option does not exist or is not a float32, 0 is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	float32: The float32 value of the option.
func (o *Options) GetFloat32(name string) float32 {
	if v, ok := o.values[name].(float32); ok {
		return v
	}
	return 0
}

// GetFloat64 returns the float64 value of the option with the specified name.
// If the option does not exist or is not a float64, 0 is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	float64: The float64 value of the option.
func (o *Options) GetFloat64(name string) float64 {
	if v, ok := o.values[name].(float64); ok {
		return v
	}
	return 0
}

// GetFloat returns the float64 value of the option with the specified name.
// If the option does not exist or is not a float64, 0 is returned.
//
// Parameters:
//
//	name (string): The name of the option to retrieve.
//
// Returns:
//
//	float64: The float64 value of the option.
func (o *Options) GetFloat(name string) float64 {
	if v, ok := o.values[name].(float64); ok {
		return v
	}
	return 0
}

// Has returns true if the option with the specified name exists in the Options.
//
// Parameters:
//
//	name (string): The name of the option to check.
//
// Returns:
//
//	bool: True if the option exists, false otherwise.
func (o *Options) Has(name string) bool {
	_, ok := o.values[name]
	return ok
}

// All returns a map of all the options in the Options.
//
// Returns:
//
//	map[string]any: A map of all the options.
func (o *Options) All() map[string]any {
	return o.values
}

// GetMaxTokens retrieves the "max_tokens" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetMaxTokens(defaultValue int) int {
	if o.Has(OptionMaxTokens) {
		return o.GetInt(OptionMaxTokens)
	}
	return defaultValue
}

// GetMinTokens retrieves the "min_tokens" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetMinTokens(defaultValue int) int {
	if o.Has(OptionMinTokens) {
		return o.GetInt(OptionMinTokens)
	}
	return defaultValue
}

// GetTemperature retrieves the "temperature" option from the Options.
// Returns the value as a float32, or the provided default value if the option does not exist.
func (o *Options) GetTemperature(defaultValue float32) float32 {
	if o.Has(OptionTemperature) {
		return o.GetFloat32(OptionTemperature)
	}
	return defaultValue
}

// GetTopP retrieves the "top_p" option from the Options.
// Returns the value as a float64, or the provided default value if the option does not exist.
func (o *Options) GetTopP(defaultValue float64) float64 {
	if o.Has(OptionTopP) {
		return o.GetFloat64(OptionTopP)
	}
	return defaultValue
}

// GetCandidateCount retrieves the "candidate_count" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetCandidateCount(defaultValue int) int {
	if o.Has(OptionCandidateCount) {
		return o.GetInt(OptionCandidateCount)
	}
	return defaultValue
}

// GetStopWords retrieves the "stop_words" option from the Options.
// Returns the value as a slice of strings, or the provided default value if the option does not exist.
func (o *Options) GetStopWords(defaultValue []string) []string {
	if o.Has(OptionStopWords) {
		return o.GetStrings(OptionStopWords)
	}
	return defaultValue
}

// GetTopK retrieves the "top_k" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetTopK(defaultValue int) int {
	if o.Has(OptionTopK) {
		return o.GetInt(OptionTopK)
	}
	return defaultValue
}

// GetSeed retrieves the "seed" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetSeed(defaultValue int) int {
	if o.Has(OptionSeed) {
		return o.GetInt(OptionSeed)
	}
	return defaultValue
}

// GetMinLength retrieves the "min_length" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetMinLength(defaultValue int) int {
	if o.Has(OptionMinLength) {
		return o.GetInt(OptionMinLength)
	}
	return defaultValue
}

// GetMaxLength retrieves the "max_length" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetMaxLength(defaultValue int) int {
	if o.Has(OptionMaxLength) {
		return o.GetInt(OptionMaxLength)
	}
	return defaultValue
}

// GetRepetitionPenalty retrieves the "repetition_penalty" option from the Options.
// Returns the value as a float64, or the provided default value if the option does not exist.
func (o *Options) GetRepetitionPenalty(defaultValue float64) float64 {
	if o.Has(OptionRepetitionPenalty) {
		return o.GetFloat64(OptionRepetitionPenalty)
	}
	return defaultValue
}

// GetFrequencyPenalty retrieves the "frequency_penalty" option from the Options.
// Returns the value as a float64, or the provided default value if the option does not exist.
func (o *Options) GetFrequencyPenalty(defaultValue float64) float64 {
	if o.Has(OptionFrequencyPenalty) {
		return o.GetFloat64(OptionFrequencyPenalty)
	}
	return defaultValue
}

// GetPresencePenalty retrieves the "presence_penalty" option from the Options.
// Returns the value as a float64, or the provided default value if the option does not exist.
func (o *Options) GetPresencePenalty(defaultValue float64) float64 {
	if o.Has(OptionPresencePenalty) {
		return o.GetFloat64(OptionPresencePenalty)
	}
	return defaultValue
}

// GetBestOf retrieves the "best_of" option from the Options.
// Returns the value as an integer, or the provided default value if the option does not exist.
func (o *Options) GetBestOf(defaultValue int) int {
	if o.Has(OptionBestOf) {
		return o.GetInt(OptionBestOf)
	}
	return defaultValue
}

// GetLogProbs retrieves the "log_probs" option from the Options.
// Returns the value as a boolean, or the provided default value if the option does not exist.
func (o *Options) GetLogProbs(defaultValue bool) bool {
	if o.Has(OptionLogProbs) {
		return o.GetBool(OptionLogProbs)
	}
	return defaultValue
}

// GetEcho retrieves the "echo" option from the Options.
// Returns the value as a boolean, or the provided default value if the option does not exist.
func (o *Options) GetEcho(defaultValue bool) bool {
	if o.Has(OptionEcho) {
		return o.GetBool(OptionEcho)
	}
	return defaultValue
}

// GetOutputMimes retrieves the "output_mimes" option from the Options.
// Returns the value as a slice of strings, or the provided default value if the option does not exist.
func (o *Options) GetOutputMimes(defaultValue []string) []string {
	if o.Has(OptionOutputMimes) {
		return o.GetStrings(OptionOutputMimes)
	}
	return defaultValue
}

// OptionsBuilder is a builder for the Options struct

type OptionsBuilder struct {
	options *Options
}

// NewOptionsBuilder creates a new OptionsBuilder instance.
//
// Returns:
//
//	*OptionsBuilder: A new OptionsBuilder instance.
func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{
		options: &Options{
			values: make(map[string]any),
		},
	}

}

// Build creates a new Options instance with the values set in the builder.
//
// Returns:
//
//	Options: The Options instance with the values set in the builder.
func (o *OptionsBuilder) Build() *Options {
	return o.options

}

// Add adds a new option with the specified name and value to the OptionsBuilder.
//
// Parameters:
//
//	name (string): The name of the option to add.
//	value (any): The value of the option to add.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) Add(name string, value any) *OptionsBuilder {
	o.options.values[name] = value
	return o
}

// SetMaxTokens sets the "max_tokens" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	maxTokens (int): The value to set the "max_tokens" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetMaxTokens(maxTokens int) *OptionsBuilder {
	o.options.values[OptionMaxTokens] = maxTokens
	return o
}

// SetMinTokens sets the "min_tokens" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	minTokens (int): The value to set the "min_tokens" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.

func (o *OptionsBuilder) SetMinTokens(minTokens int) *OptionsBuilder {
	o.options.values[OptionMinTokens] = minTokens
	return o
}

// SetTemperature sets the "temperature" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	temperature (float32): The value to set the "temperature" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.

func (o *OptionsBuilder) SetTemperature(temperature float32) *OptionsBuilder {
	o.options.values[OptionTemperature] = temperature
	return o
}

// SetTopP sets the "top_p" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	topP (float64): The value to set the "top_p" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.

func (o *OptionsBuilder) SetTopP(topP float64) *OptionsBuilder {
	o.options.values[OptionTopP] = topP
	return o
}

// SetCandidateCount sets the "candidate_count" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	candidateCount (int): The value to set the "candidate_count" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetCandidateCount(candidateCount int) *OptionsBuilder {
	o.options.values[OptionCandidateCount] = candidateCount
	return o
}

// SetStopWords sets the "stop_words" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	stopWords ([]string): The value to set the "stop_words" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetStopWords(stopWords ...string) *OptionsBuilder {
	o.options.values[OptionStopWords] = stopWords
	return o
}

// SetTopK sets the "top_k" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	topK (int): The value to set the "top_k" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.

func (o *OptionsBuilder) SetTopK(topK int) *OptionsBuilder {
	o.options.values[OptionTopK] = topK
	return o
}

// SetSeed sets the "seed" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	seed (int): The value to set the "seed" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetSeed(seed int) *OptionsBuilder {
	o.options.values[OptionSeed] = seed
	return o
}

// SetMaxLength sets the "max_length" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	maxLength (int): The value to set the "max_length" option to.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetMinLength(minLength int) *OptionsBuilder {
	o.options.values[OptionMinLength] = minLength
	return o
}

func (o *OptionsBuilder) SetMaxLength(maxLength int) *OptionsBuilder {
	o.options.values[OptionMaxLength] = maxLength
	return o
}

// SetRepetitionPenalty sets the repetition penalty value in the options.
// The repetition penalty is used to penalize the model for repeating the same token
// to encourage more diverse outputs.
//
// Parameters:
//
//	repetitionPenalty (float64): The penalty value to be set.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetRepetitionPenalty(repetitionPenalty float64) *OptionsBuilder {
	o.options.values[OptionRepetitionPenalty] = repetitionPenalty
	return o
}

// SetFrequencyPenalty sets the frequency penalty value in the options.
// The frequency penalty is a float64 value that adjusts the likelihood of
// repeating the same token in the generated output. A higher value decreases
// the likelihood of repetition.
//
// Parameters:
//
//	frequencyPenalty (float64): The frequency penalty value to set.
//
// Returns:
//
//	*OptionsBuilder: The OptionsBuilder instance with the updated frequency penalty.
func (o *OptionsBuilder) SetFrequencyPenalty(frequencyPenalty float64) *OptionsBuilder {
	o.options.values[OptionFrequencyPenalty] = frequencyPenalty
	return o
}

// SetPresencePenalty sets the presence penalty value in the options.
// The presence penalty is a float64 value that adjusts the likelihood of
// the model generating repetitive text. A higher value reduces the chance
// of repeating the same text.
//
// Parameters:
//
//	presencePenalty (float64): The presence penalty value to be set.
//
// Returns:
//
//	*OptionsBuilder: The OptionsBuilder instance with the updated presence penalty.
func (o *OptionsBuilder) SetPresencePenalty(presencePenalty float64) *OptionsBuilder {
	o.options.values[OptionPresencePenalty] = presencePenalty
	return o
}

// SetBestOf sets the "best_of" option to the specified value.
// This option determines the number of completions to generate server-side
// and return the best one. Higher values may increase latency but can improve
// the quality of the result.
//
// Parameters:
//
//	bestOf (int): The number of completions to generate and compare.
//
// Returns:
//
//	*OptionsBuilder: The updated OptionsBuilder instance.
func (o *OptionsBuilder) SetBestOf(bestOf int) *OptionsBuilder {
	o.options.values[OptionBestOf] = bestOf
	return o
}

// SetLogProbs sets the log_probs option in the OptionsBuilder.
// logProbs: A boolean value indicating whether to include log probabilities in the options.
// Returns the updated OptionsBuilder instance.
func (o *OptionsBuilder) SetLogProbs(logProbs bool) *OptionsBuilder {
	o.options.values[OptionLogProbs] = logProbs
	return o
}

// SetEcho sets the "echo" option in the OptionsBuilder to the specified value.
// It returns the updated OptionsBuilder instance.
//
// Parameters:
//
//	echo - a boolean value to set the "echo" option.
//
// Returns:
//
//	*OptionsBuilder - the updated OptionsBuilder instance.
func (o *OptionsBuilder) SetEcho(echo bool) *OptionsBuilder {
	o.options.values[OptionEcho] = echo
	return o
}

// SetOutputMimes sets the output MIME types for the OptionsBuilder.
// It accepts a variable number of string arguments representing the MIME types
// and returns a pointer to the updated OptionsBuilder.
//
// Example usage:
//
//	builder := &OptionsBuilder{}
//	builder.SetOutputMimes("application/json", "text/plain")
//
// Parameters:
//
//	outputMimes: A variable number of strings representing the desired output MIME types.
//
// Returns:
//
//	*OptionsBuilder: A pointer to the updated OptionsBuilder instance.
func (o *OptionsBuilder) SetOutputMimes(outputMimes ...string) *OptionsBuilder {
	o.options.values[OptionOutputMimes] = outputMimes
	return o
}

// AbstractModel represents a generic model with metadata information.
// It includes fields for the model's name, description, version, author, license,
// and supported input and output MIME types.
type AbstractModel struct {
	name        string
	description string
	version     string
	author      string
	license     string
	inputMime   []string
	outputMime  []string
}

// Name returns the name of the AbstractModel.
// It retrieves the value of the private field 'name' and returns it as a string.
//
// Returns:
//
//	string: The name of the model.
func (m *AbstractModel) Name() string {
	return m.name
}

// Description returns the description of the AbstractModel.
// It provides a brief summary or details about the model.
//
// Returns:
//
//	string: The description of the model.
func (m *AbstractModel) Description() string {
	return m.description
}

// Version returns the version of the AbstractModel.
// It retrieves the value of the private field 'version' and returns it as a string.
//
// Returns:
//
//	string: The version of the model.
func (m *AbstractModel) Version() string {
	return m.version
}

// Author returns the author of the AbstractModel.
// It retrieves the value of the private field 'author' and returns it as a string.
//
// Returns:
//
//	string: The author of the model.
func (m *AbstractModel) Author() string {
	return m.author
}

// License returns the license of the AbstractModel.
// It retrieves the value of the private field 'license' and returns it as a string.
//
// Returns:
//
//	string: The license of the model.
func (m *AbstractModel) License() string {
	return m.license
}

// InputMimeTypes returns the supported input MIME types for the AbstractModel.
// It retrieves the value of the private field 'inputMime' and returns it as a slice of strings.
//
// Returns:
//
//	[]string: A slice of strings representing the supported input MIME types.
func (m *AbstractModel) InputMimeTypes() []string {
	return m.inputMime
}

// OutputMimeTypes returns the supported output MIME types for the AbstractModel.
// It retrieves the value of the private field 'outputMime' and returns it as a slice of strings.
//
// Returns:
//
//	[]string: A slice of strings representing the supported output MIME types.
func (m *AbstractModel) OutputMimeTypes() []string {
	return m.outputMime
}
