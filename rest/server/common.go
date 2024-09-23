package server

import "errors"

var ErrInvalidListenHost = errors.New("empty listen host")

var ErrInvalidListenPort = errors.New("empty listen port")

var ErrInvalidPrivateKeyPath = errors.New("empty private key path")

var ErrInvalidCertPath = errors.New("empty cert path")

var ErrInvalidConfig = errors.New("empty config path")

var ErrInvalidID = errors.New("empty id")

var ErrNilOptions = errors.New("nil options")
