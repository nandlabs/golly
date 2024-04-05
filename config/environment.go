package config

//This program contains utility functions related to environment variables
import (
	"os"
	"strconv"
)

//GetEnvAsString function will fetch the val from environment variable.
//If the value is absent then it will return defaultVal supplied.
func GetEnvAsString(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal

}

//GetEnvAsInt function will fetch the val from environment variable and convert that to an integer.
//If the value is absent then it will return defaultVal supplied.
func GetEnvAsInt(key string, defaultVal int) (int, error) {
	if val, ok := os.LookupEnv(key); ok {
		return strconv.Atoi(val)
	}
	return defaultVal, nil
}

//GetEnvAsInt64 function will fetch the val from environment variable and convert that to an integer of 64 bit.
//If the value is absent then it will return defaultVal supplied.
func GetEnvAsInt64(key string, defaultVal int64) (int64, error) {
	if val, ok := os.LookupEnv(key); ok {
		return strconv.ParseInt(val, 10, 64)
	}
	return defaultVal, nil
}

//GetEnvAsDecimal function will fetch the val from environment variable and convert that to an float64.
//If the value is absent then it will return defaultVal supplied.
func GetEnvAsDecimal(key string, defaultVal float64) (float64, error) {
	if val, ok := os.LookupEnv(key); ok {
		return strconv.ParseFloat(val, 64)
	}
	return defaultVal, nil
}

//GetEnvAsBool function will fetch the val from environment variable and convert that to an GetEnvAsBool.
//If the value is absent then it will return defaultVal supplied.
// Valid boolean vals are  1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
func GetEnvAsBool(key string, defaultVal bool) (bool, error) {
	if val, ok := os.LookupEnv(key); ok {
		return strconv.ParseBool(val)
	}
	return defaultVal, nil
}
