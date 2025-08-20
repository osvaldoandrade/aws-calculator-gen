package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents configuration on disk.
type Config struct {
	Locale        string             `mapstructure:"locale"`
	Profiles      map[string]Profile `mapstructure:"profiles"`
	ActiveProfile string             `mapstructure:"active_profile"`
}

// Profile is a named profile.
type Profile struct {
	Name       string `mapstructure:"name"`
	AWSProfile string `mapstructure:"aws_profile"`
}

// Manager manages config.
type Manager struct {
	v    *viper.Viper
	path string
}

// NewManager loads config from XDG path.
func NewManager() (*Manager, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "seidor-aws-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.SetDefault("locale", "pt-BR")
	v.SetDefault("active_profile", "default")
	v.SetDefault("profiles", map[string]Profile{"default": {Name: "default"}})
	path := filepath.Join(dir, "config.yaml")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := v.WriteConfigAs(path); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &Manager{v: v, path: path}, nil
}

// Get returns Config.
func (m *Manager) Get() (Config, error) {
	var c Config
	if err := m.v.Unmarshal(&c); err != nil {
		return c, err
	}
	return c, nil
}

// Save writes config.
func (m *Manager) Save(cfg Config) error {
	m.v.Set("locale", cfg.Locale)
	m.v.Set("active_profile", cfg.ActiveProfile)
	m.v.Set("profiles", cfg.Profiles)
	return m.v.WriteConfigAs(m.path)
}

// ActiveProfile returns profile.
func (m *Manager) ActiveProfile() (Profile, error) {
	cfg, err := m.Get()
	if err != nil {
		return Profile{}, err
	}
	p, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok {
		return Profile{}, fmt.Errorf("profile not found")
	}
	return p, nil
}
