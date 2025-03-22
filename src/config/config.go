package config

import "flag"

type Config struct {
	PathToStorageFile string
	Port              string
}

func ParseCommandFlags() *Config {
	pathToStorage := flag.String("storage",
		"./isaRedis.json", "Path to your json storage file")
	port := flag.String("port",
		"6066", "Your custom port for start this key-value storage")
	flag.Parse()
	return &Config{
		PathToStorageFile: *pathToStorage,
		Port:              *port,
	}
}
