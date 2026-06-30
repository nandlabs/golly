package rest

import (
	"context"
	"io"
	"net/http"
	"strings"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/golly/turbo"
)

// jsonCodec is the default codec for json
var jsonCodec = codec.JsonCodec()

// xmlCodec is the default codec for xml
var xmlCodec = codec.XmlCodec()

// yamlCodec is the default codec for yaml
var yamlCodec = codec.YamlCodec()

// ServerContext is the struct that holds the request and response of the server.
type ServerContext struct {
	request  *http.Request
	response http.ResponseWriter
}

// Context returns the request's context. This is the context associated with
// the incoming HTTP request and can be used for cancellation, deadlines,
// and passing request-scoped values to downstream operations.
func (c *ServerContext) Context() context.Context {
	return c.request.Context()
}

// Options is the struct that holds the configuration for the Server.
func (c *ServerContext) GetParam(name string, typ Paramtype) (string, error) {
	switch typ {
	case QueryParam:
		return turbo.GetQueryParam(name, c.request)
	case PathParam:
		return turbo.GetPathParam(name, c.request)
	default:
		return textutils.EmptyStr, ErrInvalidParamType
	}
}

// GetBody returns the body of the request.
func (c *ServerContext) GetBody() (io.Reader, error) {
	return c.request.Body, nil
}

// GetHeader returns the header of the request.
func (c *ServerContext) GetHeader(name string) string {
	return c.request.Header.Get(name)
}

// InHeaders returns the headers of the request.
func (c *ServerContext) InHeaders() http.Header {
	// clone the headers
	headers := make(http.Header)
	for k, v := range c.request.Header {
		headers[k] = v
	}
	return headers
}

// GetMethod returns the method of the request.
func (c *ServerContext) GetMethod() string {
	return c.request.Method
}

// GetURL returns the URL of the request.
func (c *ServerContext) GetURL() string {
	return c.request.URL.String()
}

// GetRequest returns the request.
// for most Rest Use cases this would not be required
func (c *ServerContext) GetRequest() *http.Request {
	return c.request
}

// Read reads the body of the request into the given object.
func (c *ServerContext) Read(obj interface{}) error {
	contentType := c.request.Header.Get(ContentTypeHeader)
	codec, err := codec.GetDefault(contentType)
	if err != nil {
		return err
	}
	err = codec.Read(c.request.Body, obj)
	return err
}

// Bind decodes the request body into obj based on the request's
// Content-Type header (JSON, XML, YAML — anything codec.GetDefault
// understands). It is the read-side counterpart to Respond.
//
// Bind is an alias for Read with a name that pairs naturally with
// Respond in handler code. Prefer it in new code.
func (c *ServerContext) Bind(obj interface{}) error {
	return c.Read(obj)
}

// Respond writes payload to the response with status code, picking the
// serialization format by negotiating the client's Accept header
// against what the codec package can produce.
//
// Negotiation rules (kept simple — server picks the first match):
//   - if Accept lists application/json (or */*, or is empty)  → JSON
//   - else if Accept lists application/xml or text/xml         → XML
//   - else if Accept lists text/yaml or application/yaml       → YAML
//   - otherwise → 406 Not Acceptable, no body, returns nil
//
// Quality values (q=...) are ignored; this is a deliberately small
// negotiator. Callers needing full RFC 9110 §12.5.1 behavior should
// inspect Accept themselves and call WriteJSON / WriteXML / WriteYAML
// directly.
func (c *ServerContext) Respond(status int, payload interface{}) error {
	accept := c.request.Header.Get(AcceptHeader)
	switch chooseAcceptable(accept) {
	case ioutils.MimeApplicationJSON:
		c.SetHeader(ContentTypeHeader, ioutils.MimeApplicationJSON)
		c.SetStatusCode(status)
		return jsonCodec.Write(payload, c.response)
	case ioutils.MimeApplicationXML:
		c.SetHeader(ContentTypeHeader, ioutils.MimeApplicationXML)
		c.SetStatusCode(status)
		return xmlCodec.Write(payload, c.response)
	case ioutils.MimeTextYAML:
		c.SetHeader(ContentTypeHeader, ioutils.MimeTextYAML)
		c.SetStatusCode(status)
		return yamlCodec.Write(payload, c.response)
	default:
		c.SetStatusCode(http.StatusNotAcceptable)
		return nil
	}
}

// chooseAcceptable returns the canonical MIME type to render given the
// raw Accept header value. Empty / "*/*" defaults to JSON. The match
// is a simple substring check — q-values are ignored.
func chooseAcceptable(accept string) string {
	a := strings.ToLower(strings.TrimSpace(accept))
	if a == "" || strings.Contains(a, "*/*") {
		return ioutils.MimeApplicationJSON
	}
	if strings.Contains(a, ioutils.MimeApplicationJSON) {
		return ioutils.MimeApplicationJSON
	}
	if strings.Contains(a, ioutils.MimeApplicationXML) || strings.Contains(a, "text/xml") {
		return ioutils.MimeApplicationXML
	}
	if strings.Contains(a, ioutils.MimeTextYAML) || strings.Contains(a, "application/yaml") {
		return ioutils.MimeTextYAML
	}
	return ""
}

// WriteJSON writes the object to the response in JSON format.
func (c *ServerContext) WriteJSON(data interface{}) error {
	c.SetHeader(ContentTypeHeader, ioutils.MimeApplicationJSON)
	return jsonCodec.Write(data, c.response)
}

// WriteXML writes the object to the response in XML format.
func (c *ServerContext) WriteXML(data interface{}) error {
	c.SetHeader(ContentTypeHeader, ioutils.MimeApplicationXML)
	return xmlCodec.Write(data, c.response)
}

// WriteYAML writes the object to the response in YAML format.
func (c *ServerContext) WriteYAML(data interface{}) error {
	c.SetHeader(ContentTypeHeader, ioutils.MimeTextYAML)
	return yamlCodec.Write(data, c.response)
}

// Write writes the object to the response with the given content type and status code.
func (c *ServerContext) Write(data interface{}, contentType string) error {
	codec, err := codec.GetDefault(contentType)
	if err != nil {
		return err
	}
	c.SetHeader(ContentTypeHeader, contentType)
	return codec.Write(data, c.response)
}

// WriteData writes the data to the response.
func (c *ServerContext) WriteData(data []byte) (int, error) {
	return c.response.Write(data)
}

// WriteString writes the string to the response.
func (c *ServerContext) WriteString(data string) {

	_, _ = io.Copy(c.response, strings.NewReader(data))
}

// SetHeader sets the header of the response.
func (c *ServerContext) SetHeader(name, value string) {
	c.response.Header().Set(name, value)
}

// SetContentType sets the content type of the response.
func (c *ServerContext) SetContentType(contentType string) {
	c.response.Header().Set(ContentTypeHeader, contentType)
}

// SetStatusCode sets the status code of the response.
func (c *ServerContext) SetStatusCode(statusCode int) {
	c.response.WriteHeader(statusCode)
}

// SetCookie sets the cookie of the response.
func (c *ServerContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

// WriteFrom writes the data from the reader to the response.
func (c *ServerContext) WriteFrom(data io.Reader) {
	_, _ = io.Copy(c.response, data)
}

// HttpResWriter returns the http.ResponseWriter
func (c *ServerContext) HttpResWriter() http.ResponseWriter {
	return c.response
}
