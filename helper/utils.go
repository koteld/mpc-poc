package helper

import "os"

func GetEnv(key string, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		val = def
	}
	return val
}
