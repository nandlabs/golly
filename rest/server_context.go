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

	io.Copy(c.response, strings.NewReader(data))
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
	io.Copy(c.response, data)
}

// HttpResWriter returns the http.ResponseWriter
func (c *ServerContext) HttpResWriter() http.ResponseWriter {
	return c.response
}
