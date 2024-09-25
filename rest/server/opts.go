package server

// Options is the configuration for the server
type Options struct {
	Id             string `json:"id" yaml:"id" bson:"id" mapstructure:"id"`
	PathPrefix     string `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty" bson:"path_prefix,omitempty" mapstructure:"path_prefix,omitempty"`
	ListenHost     string `json:"listen_host" yaml:"listen_host" bson:"listen_host" mapstructure:"listen_host"`
	ListenPort     int16  `json:"listen_port" yaml:"listen_port" bson:"listen_port" mapstructure:"listen_port"`
	ReadTimeout    int64  `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty" bson:"read_timeout,omitempty" mapstructure:"read_timeout,omitempty"`
	WriteTimeout   int64  `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" bson:"write_timeout,omitempty" mapstructure:"write_timeout,omitempty"`
	EnableTLS      bool   `json:"enable_tls" yaml:"enable_tls" bson:"enable_tls" mapstructure:"enable_tls"`
	PrivateKeyPath string `json:"private_key_path,omitempty" yaml:"private_key_path,omitempty" bson:"private_key_path,omitempty" mapstructure:"private_key,omitempty"`
	CertPath       string `json:"cert_path,omitempty" yaml:"cert_path,omitempty" bson:"cert_path,omitempty" mapstructure:"cert,omitempty"`
}

// Validate validates the server options
func (o Options) Validate() error {
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
func (o Options) GetListenHost() string {
	return o.ListenHost
}

// GetListenPort returns the listen port
func (o Options) GetListenPort() int16 {
	return o.ListenPort
}

// GetEnableTLS returns the enable TLS value
func (o Options) GetEnableTLS() bool {
	return o.EnableTLS
}

// GetPrivateKeyPath returns the private key path
func (o Options) GetPrivateKeyPath() string {
	return o.PrivateKeyPath
}

// GetCertPath returns the cert path
func (o Options) GetCertPath() string {
	return o.CertPath
}

// SetListenHost sets the listen host
func (o Options) SetListenHost(host string) Options {
	o.ListenHost = host
	return o
}

// SetListenPort sets the listen port
func (o Options) SetListenPort(port int16) Options {

	o.ListenPort = port
	return o
}

// SetEnableTLS sets the enable TLS value
func (o Options) SetEnableTLS(enableTLS bool) Options {
	o.EnableTLS = enableTLS
	return o
}

// SetPrivateKeyPath sets the private key path
func (o Options) SetPrivateKeyPath(privateKeyPath string) Options {
	o.PrivateKeyPath = privateKeyPath
	return o
}

// SetCertPath sets the cert path
func (o Options) SetCertPath(certPath string) Options {
	o.CertPath = certPath
	return o
}

// NewOptions returns a new server options
func NewOptions() Options {
	return Options{}
}

// NewOptionsWithDefaults returns a new server options with default values
func NewOptionsWithDefaults() Options {
	return Options{
		ListenHost: "localhost",
		ListenPort: 8080,
	}
}

// DefaultOptions returns the default options for the server
func DefaultOptions() *Options {
	return &Options{
		PathPrefix:   "/",
		Id:           "default-http-server",
		ListenHost:   "localhost",
		ListenPort:   8080,
		ReadTimeout:  20000,
		WriteTimeout: 20000,
	}
}
