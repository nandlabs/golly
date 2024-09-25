package server

import (
	"io"
	"net/http"
	"strings"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/golly/turbo"
)

// Context is the struct that holds the request and response of the server.
type Context struct {
	request  *http.Request
	response http.ResponseWriter
}

// Options is the struct that holds the configuration for the Server.
func (c *Context) GetParam(name string, typ Paramtype) (string, error) {
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
func (c *Context) GetBody() (io.Reader, error) {
	return c.request.Body, nil
}

// GetHeader returns the header of the request.
func (c *Context) GetHeader(name string) string {
	return c.request.Header.Get(name)
}

// InHeaders returns the headers of the request.
func (c *Context) InHeaders() http.Header {
	// clone the headers
	headers := make(http.Header)
	for k, v := range c.request.Header {
		headers[k] = v
	}
	return headers
}

// GetMethod returns the method of the request.
func (c *Context) GetMethod() string {
	return c.request.Method
}

// GetURL returns the URL of the request.
func (c *Context) GetURL() string {
	return c.request.URL.String()
}

// GetRequest returns the request.
// for most Rest Use cases this would not be required
func (c *Context) GetRequest() *http.Request {
	return c.request
}

// Read reads the body of the request into the given object.
func (c *Context) Read(obj interface{}) error {
	contentType := c.request.Header.Get(rest.ContentTypeHeader)
	codec, err := codec.GetDefault(contentType)
	if err != nil {
		return err
	}
	err = codec.Read(c.request.Body, obj)
	return err
}

// Write writes the object to the response with the given content type and status code.
func (c *Context) Write(data interface{}, contentType string) error {
	codec, err := codec.GetDefault(contentType)
	if err != nil {
		return err
	}
	c.SetHeader(rest.ContentTypeHeader, contentType)
	return codec.Write(data, c.response)
}

// WriteData writes the data to the response.
func (c *Context) WriteData(data []byte) (int, error) {
	return c.response.Write(data)
}

// WriteString writes the string to the response.
func (c *Context) WriteString(data string) {

	io.Copy(c.response, strings.NewReader(data))
}

// SetHeader sets the header of the response.
func (c *Context) SetHeader(name, value string) {
	c.response.Header().Set(name, value)
}

// SetContentType sets the content type of the response.
func (c *Context) SetContentType(contentType string) {
	c.response.Header().Set(rest.ContentTypeHeader, contentType)
}

// SetStatusCode sets the status code of the response.
func (c *Context) SetStatusCode(statusCode int) {
	c.response.WriteHeader(statusCode)
}

// SetCookie sets the cookie of the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

// WriteFrom writes the data from the reader to the response.
func (c *Context) WriteFrom(data io.Reader) {
	io.Copy(c.response, data)
}

// HttpResWriter returns the http.ResponseWriter
func (c *Context) HttpResWriter() http.ResponseWriter {
	return c.response
}
