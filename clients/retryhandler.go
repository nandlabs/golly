package clients

// RetryInfo represents the retry configuration for a client.
type RetryInfo struct {
	MaxRetries int // Maximum number of retries allowed.
	Wait       int // Wait time in milliseconds between retries.
}
