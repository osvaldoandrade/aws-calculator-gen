package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load returns a viper instance reading config from XDG directory.
func Load() (*viper.Viper, error) {
	v := viper.New()
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "seidor-aws-cli")
	os.MkdirAll(dir, 0o755)
	v.AddConfigPath(dir)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		// ignore missing file
	}
	return v, nil
}
