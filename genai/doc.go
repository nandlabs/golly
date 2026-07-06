// Package genai defines the provider abstraction for Generative AI services.
//
// It includes interfaces and types for interacting with large language models
// (LLMs) in a provider-agnostic way. Concrete implementations are available
// in the genai/impl sub-package.
//
// # Routing
//
// The package also provides a task-aware model router that sits above the
// unchanged Provider abstraction. Callers describe a Task (required
// capabilities, size estimates, cost budget, priority) and the router selects a
// (provider, model), reconciles Options with the chosen model's limits, and
// executes with error-classified, ctx-aware fallback composed with the
// clients.RetryPolicy and clients.CircuitBreaker primitives. Providers opt in to
// capability-based routing by implementing the CapabilityProvider interface;
// providers that do not remain fully usable and reachable by rule or Pin. See
// ModelRouter, ModelRegistry, RoutingStrategy, and EmbeddingRouter.
package genai
