package rest

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"oss.nandlabs.io/golly/clients"
	"oss.nandlabs.io/golly/textutils"
)

const (
	defaultMaxIdleConnections    = 20
	defaultReqTimeout            = 60
	defaultIdleConnTimeout       = 90
	defaultTLSHandshakeTimeout   = 10
	defaultExpectContinueTimeout = 0
	proxyAuthHdr                 = "Proxy-Authorization"
)

type AuthHandlerFunc func(client *Client, req *http.Request) error

var basicAuthHandlerFunc = func(client *Client, req *http.Request) error {

	if client.options.Auth == nil || client.options.Auth.Type() != clients.AuthTypeBasic {
		return fmt.Errorf("invalid auth type ")
	}
	req.SetBasicAuth(client.options.Auth.User(), client.options.Auth.Pass())
	return nil
}

var bearerAuthHandlerFunc = func(client *Client, req *http.Request) error {
	if client.options.Auth == nil || client.options.Auth.Type() != clients.AuthTypeBasic {
		return fmt.Errorf("invalid auth type")
	}
	req.Header.Set("Authorization", "Bearer "+client.options.Auth.Token())

	return nil
}

// ClientOpts represents the options for the REST client.

type ClientOpts struct {
	*clients.ClientOptions
	proxyBasicAuth        string
	codecOptions          map[string]any
	maxIdlePerHost        int
	baseUrl               *url.URL
	errorOnMap            map[int]int
	tlsConfig             *tls.Config
	useCustomTLSConfig    bool
	jar                   http.CookieJar
	idleTimeout           time.Duration
	requestTimeout        time.Duration
	tlsHandShakeTimeout   time.Duration
	expectContinueTimeout time.Duration
	AuthHandlers          map[clients.AuthType]AuthHandlerFunc
}

type ClientOptsBuilder struct {
	*clients.OptionsBuilder
	opts *ClientOpts
}

func ClientOptBuilder() *ClientOptsBuilder {
	return &ClientOptsBuilder{
		OptionsBuilder: clients.NewOptionsBuilder(),
		opts: &ClientOpts{
			ClientOptions: clients.EmptyClientOptions,
			tlsConfig:     &tls.Config{},
		},
	}
}

func (co *ClientOptsBuilder) EnvProxy(proxyBasicAuth string) *ClientOptsBuilder {
	co.opts.proxyBasicAuth = proxyBasicAuth
	return co
}

func (co *ClientOptsBuilder) CodecOpts(options map[string]any) *ClientOptsBuilder {
	co.opts.codecOptions = options
	return co
}

func (co *ClientOptsBuilder) MaxIdlePerHost(maxIdleConnPerHost int) *ClientOptsBuilder {
	co.opts.maxIdlePerHost = maxIdleConnPerHost
	return co
}

func (co *ClientOptsBuilder) ProxyAuth(user, password, bypass string) *ClientOptsBuilder {
	co.opts.proxyBasicAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
	return co
}

func (co *ClientOptsBuilder) BaseUrl(baseurl string) (err error) {
	if baseurl == textutils.EmptyStr {
		return errors.New("invalid base url")
	}
	co.opts.baseUrl, err = url.Parse(baseurl)
	if err == nil && co.opts.baseUrl.Scheme == textutils.EmptyStr && co.opts.baseUrl.Host == textutils.EmptyStr {
		err = errors.New("invalid base url")
	} else {
		if !strings.HasSuffix(co.opts.baseUrl.Path, textutils.ForwardSlashStr) {
			co.opts.baseUrl.Path = co.opts.baseUrl.Path + textutils.ForwardSlashStr
		}
	}

	return
}

func (co *ClientOptsBuilder) ErrOnStatus(httpStatusCodes ...int) *ClientOptsBuilder {
	if co.opts.errorOnMap == nil {
		co.opts.errorOnMap = make(map[int]int)
	}
	for _, code := range httpStatusCodes {
		co.opts.errorOnMap[code] = code
	}
	return co
}

func (co *ClientOptsBuilder) SSLVerify(verify bool) *ClientOptsBuilder {
	if co.opts.tlsConfig == nil {
		co.opts.tlsConfig = &tls.Config{}
	}
	co.opts.tlsConfig.InsecureSkipVerify = verify
	co.opts.useCustomTLSConfig = true
	return co
}

func (co *ClientOptsBuilder) CaCerts(caFilePath ...string) *ClientOptsBuilder {
	co.opts.useCustomTLSConfig = true
	if co.opts.tlsConfig == nil {
		co.opts.tlsConfig = &tls.Config{}
	}
	if co.opts.tlsConfig.RootCAs == nil {
		co.opts.tlsConfig.RootCAs = x509.NewCertPool()
	}
	for _, v := range caFilePath {
		caCert, err := os.ReadFile(v)
		if err != nil {
			logger.Error("error reading ca cert file", err)
			continue
		}
		co.opts.tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
	}
	return co
}

func (co *ClientOptsBuilder) TlsCerts(certs ...tls.Certificate) *ClientOptsBuilder {
	if co.opts.tlsConfig == nil {
		co.opts.tlsConfig = &tls.Config{}
	}
	co.opts.useCustomTLSConfig = true
	co.opts.tlsConfig.Certificates = append(co.opts.tlsConfig.Certificates, certs...)
	return co
}

func (co *ClientOptsBuilder) IdleTimeoutMs(t int) *ClientOptsBuilder {
	co.opts.idleTimeout = time.Duration(t) * time.Millisecond
	return co
}

func (co *ClientOptsBuilder) RequestTimeoutMs(t int) *ClientOptsBuilder {
	co.opts.requestTimeout = time.Duration(t) * time.Millisecond
	return co
}
func (co *ClientOptsBuilder) TlsHandShakeTimeoutMs(t int) *ClientOptsBuilder {
	co.opts.tlsHandShakeTimeout = time.Duration(t) * time.Millisecond
	return co
}

func (co *ClientOptsBuilder) ExpectContinueTimeoutMs(t int) *ClientOptsBuilder {
	co.opts.expectContinueTimeout = time.Duration(t) * time.Millisecond
	return co
}
func (co *ClientOptsBuilder) CookieJar(jar http.CookieJar) *ClientOptsBuilder {
	co.opts.jar = jar
	return co
}

func (co *ClientOptsBuilder) Build() *ClientOpts {
	return co.opts
}

// Client represents a REST client.
type Client struct {
	// retryInfo      *clients.RetryInfo
	// circuitBreaker *clients.CircuitBreaker
	httpClient    http.Client
	httpTransport *http.Transport
	options       *ClientOpts
}

func NewClientWithOptions(options *ClientOpts) *Client {
	client := &Client{}
	if options.AuthHandlers == nil {
		options.AuthHandlers = map[clients.AuthType]AuthHandlerFunc{
			clients.AuthTypeBasic:  basicAuthHandlerFunc,
			clients.AuthTypeBearer: bearerAuthHandlerFunc,
		}
	}
	client.options = options
	client.httpTransport = &http.Transport{
		MaxIdleConnsPerHost:   options.maxIdlePerHost,
		IdleConnTimeout:       options.idleTimeout,
		TLSHandshakeTimeout:   defaultTLSHandshakeTimeout * time.Second,
		ExpectContinueTimeout: defaultExpectContinueTimeout * time.Second,
	}
	if options.useCustomTLSConfig {
		client.httpTransport.TLSClientConfig = options.tlsConfig
	}
	client.httpClient = http.Client{
		Transport: client.httpTransport,
		Timeout:   options.requestTimeout,
		Jar:       options.jar,
	}

	return client
}

// NewClient creates a new REST client with default values.
func NewClient() *Client {
	return NewClientWithOptions(&ClientOpts{
		ClientOptions: clients.EmptyClientOptions,
		tlsConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		maxIdlePerHost:        defaultMaxIdleConnections,
		useCustomTLSConfig:    false,
		tlsHandShakeTimeout:   defaultTLSHandshakeTimeout * time.Second,
		requestTimeout:        time.Duration(defaultReqTimeout) * time.Second,
		idleTimeout:           defaultIdleConnTimeout * time.Second,
		expectContinueTimeout: defaultExpectContinueTimeout * time.Second,
	})
}

// NewRequest creates a new request object for the client.
func (c *Client) NewRequest(reqUrl, method string) (req *Request, err error) {
	finalUrl := reqUrl
	u, err := url.Parse(reqUrl)
	if err != nil {
		return
	}
	if u.Scheme == textutils.EmptyStr && u.Host == textutils.EmptyStr && c.options.baseUrl != nil {
		finalUrl = c.options.baseUrl.String() + u.Path
	}

	req = &Request{
		url:    finalUrl,
		method: method,
		header: map[string][]string{},
		client: c,
	}
	return
}

// Execute sends the client request and returns the response object.
func (c *Client) Execute(req *Request) (res *Response, err error) {
	var httpReq *http.Request
	var httpRes *http.Response
	httpReq, err = req.toHttpRequest()
	if c.options.proxyBasicAuth != textutils.EmptyStr {
		httpReq.Header.Set(proxyAuthHdr, c.options.proxyBasicAuth)
	}
	if err == nil {
		if c.options.Auth != nil {
			if handlerFunc, ok := c.options.AuthHandlers[c.options.Auth.Type()]; ok {
				err = handlerFunc(c, httpReq)
				if err != nil {
					return
				}
			} else {
				err = fmt.Errorf("invalid auth type or no handlerfunc found for auth type %v", c.options.Auth.Type())
				return
			}
		}
		// Check if the circuit breaker is open
		if c.options.CircuitBreaker != nil {
			err = c.options.CircuitBreaker.CanExecute()
			// If the circuit breaker is open, return an error
			if err != nil {
				return
			}
		}
		// Execute the request
		httpRes, err = c.httpClient.Do(httpReq)
		// Check if the response is an error
		isErr := c.isError(err, httpRes)
		if c.options.CircuitBreaker != nil {
			c.options.CircuitBreaker.OnExecution(!isErr)
		}
		if isErr && c.options.RetryPolicy != nil {
			retryCount := 0
			// For each retry, sleep for the backoff interval and retry the request
			for isErr && retryCount < c.options.RetryPolicy.MaxRetries {
				time.Sleep(c.options.RetryPolicy.WaitTime(retryCount))
				retryCount++
				httpRes, err = c.httpClient.Do(httpReq)
				isErr = c.isError(err, httpRes)
				if c.options.CircuitBreaker != nil {
					c.options.CircuitBreaker.OnExecution(!isErr)
				}
			}
		}

		httpRes, err = c.httpClient.Do(httpReq)

		if err == nil {
			res = &Response{raw: httpRes, client: c}
		}
	}
	return
}

// isError checks if the response is an error response or an error has been received.
func (c *Client) isError(err error, httpRes *http.Response) (isErr bool) {
	isErr = err != nil
	if !isErr && c.options.errorOnMap != nil {
		_, isErr = c.options.errorOnMap[httpRes.StatusCode]
	}
	return
}

// Close closes all idle connections that are available.
func (c *Client) Close() (err error) {
	c.httpClient.CloseIdleConnections()
	return
}
