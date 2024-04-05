package config

import "io"

// Configuration is an interface that wraps the  methods for a standard configuration.

type Configuration interface {

	//Load a reader from Reader
	Load(r io.Reader) error
	//Save to a writer
	Save(w io.Writer) error
	//Get returns configuration value as string identified by the key
	//If the value is absent then it will return defaultVal supplied.
	Get(k, defaultVal string) string
	//GetAsInt returns the config value as int64 identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non Int value is present for the key
	GetAsInt(k string, defaultVal int) (int, error)
	//GetAsInt64 returns the config value as int64 identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non int64 value is present for the key
	GetAsInt64(k string, defaultVal int64) (int64, error)
	//GetAsBool returns the config value as bool identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non bool value is present for the key
	GetAsBool(k string, defaultVal bool) (bool, error)
	//GetAsDecimal returns the config value as decimal float64 identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non float64 value is present for the key
	GetAsDecimal(k string, defaultVal float64) (float64, error)
	//Put returns configuration value as string identified by the key
	//If the value is absent then it will return defaultVal supplied.
	Put(k, v string) string
	//PutInt returns the config value as int64 identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non Int value is present for the key
	PutInt(k string, v int) (int, error)
	//PutInt64 returns the config value as int64 identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non int64 value is present for the key
	PutInt64(k string, v int64) (int64, error)
	//PutBool returns the config value as bool identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non bool value is present for the key
	PutBool(k string, v bool) (bool, error)
	//PutDecimal returns the config value as decimal float64 identified by the key
	//If the value is absent then it will return defaultVal supplied.
	//This may throw an error if a non float64 value is present for the key
	PutDecimal(k string, v float64) (float64, error)
}
