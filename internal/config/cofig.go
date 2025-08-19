package config

import (
	"flag"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"address"`
}
type Config struct {
	HTTPServer `yaml:"http_server"`
}

func MustLoad(logger *slog.Logger) *Config {
	var configPath string
	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config", "", "path of configuration file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			logger.Error("config path is not set")
			os.Exit(1)
		}

	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Error("config file does not exist", "CONFIG_PATH", configPath)
		os.Exit(1)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		logger.Error("failed to read config file", "error", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded successfully", "config_path", configPath)
	return &cfg

}
