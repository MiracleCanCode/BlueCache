package config

import "flag"

type Config struct {
	PathToStorageFile string
	Port              string
}

func ParseCommandFlags() *Config {
	pathToStorage := flag.String("storage", "./isaRedis.log", "Path to your storage file")
	port := flag.String("port", "6066", "Your custom port for start this key-value storage")
	flag.Parse()
	return &Config{
		PathToStorageFile: *pathToStorage,
		Port:              *port,
	}
}
