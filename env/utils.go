package env

import "os"

func GetEnv(name string, defaults ...string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	if len(defaults) > 0 {
		return defaults[0]
	}

	return ""
}

func SetEnv(name, value string) error {
	return os.Setenv(name, value)
}
