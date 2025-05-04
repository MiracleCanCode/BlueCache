package config

import (
	"flag"
	"fmt"
  "strings"
	"github.com/ilyakaznacheev/cleanenv"
)

type userData struct {
	Name     string `env:"USER_NAME"`
	Password string `env:"USER_PASSWORD"`
}

type Config struct {
	PathToStorageFile string
	Port              string
	Logging           bool
	MasterURL         string
	UserName          string
	UserPassword      string
}

const TXT_FORMAT = ".txt"

func New() (*Config, error) {
	var user userData
	_ = cleanenv.ReadEnv(&user)

	pathToStorage := flag.String("storage", "./isaRedis.txt", "Path to your JSON storage file")
	port := flag.String("port", "6066", "Custom port for BlueCache instance")
	logging := flag.Bool("logging", false, "Enable request logging")
	masterURL := flag.String("master_url", "", "Set the current server as a replica")
	userName := flag.String("user_name", "blueCache", "User name for BlueCache connection")
	userPassword := flag.String("user_password", "blueCache", "User password for BlueCache connection")

	flag.Parse()
  
  if(!strings.HasSuffix(*pathToStorage, TXT_FORMAT)) {
    return nil, fmt.Errorf("config: storage file is not txt file")
  }

	finalUserName := fallbackIfEmpty(*userName, user.Name)
	finalUserPassword := fallbackIfEmpty(*userPassword, user.Password)

	if finalUserName == "" {
		return nil, fmt.Errorf("config: user name is required but not provided")
	}
	if finalUserPassword == "" {
		return nil, fmt.Errorf("config: user password is required but not provided")
	}

	return &Config{
		PathToStorageFile: *pathToStorage,
		Port:              *port,
		Logging:           *logging,
		MasterURL:         *masterURL,
		UserName:          finalUserName,
		UserPassword:      finalUserPassword,
	}, nil
}

func fallbackIfEmpty(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}
