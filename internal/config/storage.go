package config

import (
	"fmt"
)

type PostgresStorage struct {
	Username string `yaml:"username" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Host     string `yaml:"host"     env-default:"localhost"`
	Port     int    `yaml:"port"     env-default:"5432"                validate:"omitempty,gte=0"`
	Database string `yaml:"database" env-default:"smart_pc_pc_service"`
	SSL      bool   `yaml:"ssl"      env-default:"false"`
}

func (s *PostgresStorage) ConnectionString() string {
	const provider = "postgres"

	sslMode := "disable"
	if s.SSL {
		sslMode = "require"
	}
	return fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?sslmode=%s",
		provider,
		s.Username,
		s.Password,
		s.Host,
		s.Port,
		s.Database,
		sslMode,
	)
}
