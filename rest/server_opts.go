package rest

import (
	"net/http"

	"oss.nandlabs.io/golly/turbo/filters"
)

// SrvOptions is the configuration for the server
type SrvOptions struct {
	Id             string               `json:"id" yaml:"id" bson:"id" mapstructure:"id"`
	PathPrefix     string               `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty" bson:"path_prefix,omitempty" mapstructure:"path_prefix,omitempty"`
	ListenHost     string               `json:"listen_host" yaml:"listen_host" bson:"listen_host" mapstructure:"listen_host"`
	ListenPort     int16                `json:"listen_port" yaml:"listen_port" bson:"listen_port" mapstructure:"listen_port"`
	ReadTimeout    int64                `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty" bson:"read_timeout,omitempty" mapstructure:"read_timeout,omitempty"`
	WriteTimeout   int64                `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" bson:"write_timeout,omitempty" mapstructure:"write_timeout,omitempty"`
	EnableTLS      bool                 `json:"enable_tls" yaml:"enable_tls" bson:"enable_tls" mapstructure:"enable_tls"`
	PrivateKeyPath string               `json:"private_key_path,omitempty" yaml:"private_key_path,omitempty" bson:"private_key_path,omitempty" mapstructure:"private_key,omitempty"`
	CertPath       string               `json:"cert_path,omitempty" yaml:"cert_path,omitempty" bson:"cert_path,omitempty" mapstructure:"cert,omitempty"`
	Cors           *filters.CorsOptions `json:"cors,omitempty" yaml:"cors,omitempty" bson:"cors,omitempty" mapstructure:"cors,omitempty"`
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
		if o.PrivateKeyPath == "" {
			return ErrInvalidPrivateKeyPath
		}
		if o.CertPath == "" {
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
