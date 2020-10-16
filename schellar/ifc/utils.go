package ifc

import "os"

func GetEnvOrDefault(key string, defaultValue string) string {
	found, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return found
}
