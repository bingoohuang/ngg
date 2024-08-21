package unit

import "os"

// GetEnvBytes 获得环境变量 name 的值所表示的大小，例如. 30MiB
func GetEnvBytes(name string, defaultValue uint64) (uint64, error) {
	env := os.Getenv(name)
	if env == "" {
		return defaultValue, nil
	}

	return ParseBytes(env)
}
