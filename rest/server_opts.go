package rest

import (
	"crypto/tls"
	"net/http"

	"oss.nandlabs.io/golly/turbo/filters"
)

// SrvOptions is the configuration for the server.
//
// TLS notes:
//   - The simple path: set EnableTLS = true + CertPath + PrivateKeyPath.
//     The server reads the cert from disk and serves TLS with sensible
//     defaults (min version TLS 1.2).
//   - For finer control (custom cipher suites, mTLS / client-cert
//     verification, cert reloading via GetCertificate, ALPN protos)
//     set TLSConfig to a fully-populated *tls.Config. When TLSConfig
//     is non-nil it takes precedence over CertPath / PrivateKeyPath.
//   - TLSMinVersion overrides the negotiated minimum on either path;
//     default is TLS 1.2 (tls.VersionTLS12). Set to tls.VersionTLS13
//     to require 1.3.
//
// StrictSlash forwards to turbo.Router.StrictSlash — when false (the
// default) the router treats "/users" and "/users/" as equivalent;
// when true they are distinct routes that must be registered
// separately.
type SrvOptions struct {
	Id             string `json:"id" yaml:"id" bson:"id" mapstructure:"id"`
	PathPrefix     string `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty" bson:"path_prefix,omitempty" mapstructure:"path_prefix,omitempty"`
	ListenHost     string `json:"listen_host" yaml:"listen_host" bson:"listen_host" mapstructure:"listen_host"`
	ListenPort     int16  `json:"listen_port" yaml:"listen_port" bson:"listen_port" mapstructure:"listen_port"`
	ReadTimeout    int64  `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty" bson:"read_timeout,omitempty" mapstructure:"read_timeout,omitempty"`
	WriteTimeout   int64  `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" bson:"write_timeout,omitempty" mapstructure:"write_timeout,omitempty"`
	EnableTLS      bool   `json:"enable_tls" yaml:"enable_tls" bson:"enable_tls" mapstructure:"enable_tls"`
	PrivateKeyPath string `json:"private_key_path,omitempty" yaml:"private_key_path,omitempty" bson:"private_key_path,omitempty" mapstructure:"private_key,omitempty"`
	CertPath       string `json:"cert_path,omitempty" yaml:"cert_path,omitempty" bson:"cert_path,omitempty" mapstructure:"cert,omitempty"`
	// TLSMinVersion is the minimum negotiated TLS version (one of
	// tls.VersionTLS12, tls.VersionTLS13, etc). When 0, defaults to
	// TLS 1.2.
	TLSMinVersion uint16 `json:"tls_min_version,omitempty" yaml:"tls_min_version,omitempty" bson:"tls_min_version,omitempty" mapstructure:"tls_min_version,omitempty"`
	// TLSConfig, when non-nil, is used verbatim for TLS — overriding
	// CertPath / PrivateKeyPath / TLSMinVersion. Use for mTLS, custom
	// cipher suites, GetCertificate-based reloading, ALPN protos.
	// Not serializable; populate programmatically.
	TLSConfig *tls.Config `json:"-" yaml:"-" bson:"-" mapstructure:"-"`
	// StrictSlash forwards to turbo.Router.StrictSlash. Default false
	// (lenient — "/users" and "/users/" match the same route).
	StrictSlash bool                 `json:"strict_slash,omitempty" yaml:"strict_slash,omitempty" bson:"strict_slash,omitempty" mapstructure:"strict_slash,omitempty"`
	Cors        *filters.CorsOptions `json:"cors,omitempty" yaml:"cors,omitempty" bson:"cors,omitempty" mapstructure:"cors,omitempty"`
}

// Validate validates the server options
func (o *SrvOptions) Validate() error {
	if o.Id == "" {
		return ErrInvalidID
	}
	if o.ListenHost == "" {
		return ErrInvalidListenHost
	}
	if o.ListenPort <= 0 {
		return ErrInvalidListenPort
	}
	if o.EnableTLS {
		// EnableTLS requires either disk-based cert/key paths OR a
		// caller-supplied TLSConfig that already carries certificates
		// (via Certificates / GetCertificate). The latter is how mTLS
		// and cert reloading are configured.
		hasPaths := o.CertPath != "" && o.PrivateKeyPath != ""
		hasConfig := o.TLSConfig != nil &&
			(len(o.TLSConfig.Certificates) > 0 || o.TLSConfig.GetCertificate != nil)
		if !hasPaths && !hasConfig {
			if o.PrivateKeyPath == "" {
				return ErrInvalidPrivateKeyPath
			}
			return ErrInvalidCertPath
		}
	}
	return nil
}

// GetListenHost returns the listen host
func (o *SrvOptions) GetListenHost() string {
	return o.ListenHost
}

// GetListenPort returns the listen port
func (o *SrvOptions) GetListenPort() int16 {
	return o.ListenPort
}

// GetEnableTLS returns the enable TLS value
func (o *SrvOptions) GetEnableTLS() bool {
	return o.EnableTLS
}

// GetPrivateKeyPath returns the private key path
func (o *SrvOptions) GetPrivateKeyPath() string {
	return o.PrivateKeyPath
}

// GetCertPath returns the cert path
func (o *SrvOptions) GetCertPath() string {
	return o.CertPath
}

// SetListenHost sets the listen host
func (o *SrvOptions) SetListenHost(host string) *SrvOptions {
	o.ListenHost = host
	return o
}

// SetListenPort sets the listen port
func (o *SrvOptions) SetListenPort(port int16) *SrvOptions {

	o.ListenPort = port
	return o
}

// SetEnableTLS sets the enable TLS value
func (o *SrvOptions) SetEnableTLS(enableTLS bool) *SrvOptions {
	o.EnableTLS = enableTLS
	return o
}

// SetPrivateKeyPath sets the private key path
func (o *SrvOptions) SetPrivateKeyPath(privateKeyPath string) *SrvOptions {
	o.PrivateKeyPath = privateKeyPath
	return o
}

// SetCertPath sets the cert path
func (o *SrvOptions) SetCertPath(certPath string) *SrvOptions {
	o.CertPath = certPath
	return o
}

// EmptySrvOptions returns a new server options
func EmptySrvOptions() *SrvOptions {
	return &SrvOptions{}
}

// DefaultSrvOptions returns the default options for the server
// The default options are:
//   - PathPrefix: "/"
//   - Id: "default-http-server"
//   - ListenHost: "localhost"
//   - ListenPort: 8080
//   - ReadTimeout: 20000
//   - WriteTimeout: 20000
//   - Cors: &filters.CorsOptions{
//     MaxAge:         0,
//     AllowedOrigins: []string{"*"},
//     AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
//     ResponseStatus: http.StatusNoContent,
//     }
func DefaultSrvOptions() *SrvOptions {
	return &SrvOptions{
		PathPrefix:   "/",
		Id:           "default-http-server",
		ListenHost:   "localhost",
		ListenPort:   8080,
		ReadTimeout:  20000,
		WriteTimeout: 20000,
		Cors: &filters.CorsOptions{
			MaxAge:         filters.DefaultAccessControlMaxAge,
			AllowedOrigins: []string{filters.AccessControlAllowAllOrigins},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
			ResponseStatus: http.StatusNoContent,
		},
	}
}
