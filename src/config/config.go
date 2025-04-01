package config

import "flag"

type Config struct {
	PathToStorageFile string
	Port              string
	LogRequest        bool
}

func ParseCommandFlags() *Config {
	pathToStorage := flag.String("storage",
		"./isaRedis.txt", "Path to your json storage file")
	port := flag.String("port",
		"6066", "Your custom port for start this key-value storage")

	logRequest := flag.Bool("logging", false, "Logging all requests to the repository")
	flag.Parse()
	return &Config{
		PathToStorageFile: *pathToStorage,
		Port:              *port,
		LogRequest:        *logRequest,
	}
}
