package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env"         validate:"required,oneof=dev debug production" env-default:"production"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Storage    Storage    `yaml:"storage"`
}

type HTTPServer struct {
	Address         string        `yaml:"address"          env-default:"localhost:8080"`
	Timeout         time.Duration `yaml:"timeout"          env-default:"4s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"     env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"1s"`
}

type Storage struct {
	Postgres PostgresStorage `yaml:"postgres"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fullConfigPath, _ := filepath.Abs(configPath)
		log.Fatalf("config file does not exists by path \"%s\"", fullConfigPath)
	}

	cfg := &Config{}
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		log.Fatalf("can not read config: %s", err)
	}

	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		if validationError, ok := errors.AsType[validator.ValidationErrors](err); ok {
			log.Fatalf("config validation failed: %s", validationErrorMessages(validationError))
		}
		log.Fatalf("config validation error: %s", err)
	}

	return cfg
}

func validationErrorMessages(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(
				errMsgs,
				fmt.Sprintf("field %q is required", err.StructNamespace()),
			)
		case "oneof":
			errMsgs = append(
				errMsgs,
				fmt.Sprintf("field %q must be one of %q", err.StructNamespace(), err.Param()),
			)
		case "gte":
			errMsgs = append(
				errMsgs,
				fmt.Sprintf(
					"field %q must be greater or equal to %s",
					err.StructNamespace(),
					err.Param(),
				),
			)
		default:
			errMsgs = append(
				errMsgs,
				fmt.Sprintf("field %q is not valid", err.StructNamespace()),
			)
		}
	}

	return strings.Join(errMsgs, ", ")
}
