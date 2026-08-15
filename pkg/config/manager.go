package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Manager struct {
	Path string
	Data *ConfigData
}

type ConfigData struct {
	DefaultProvider    string                    `json:"default_provider"`
	DefaultProviderCwd string                    `json:"default_provider_cwd"`
	Auths              map[string]json.RawMessage `json:"auths"`
	Providers          map[string]ProviderConfig  `json:"providers"`
}

type ProviderConfig struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(home, ".config", "storage-bridge", "config.json")
	return NewManagerWithPath(configPath)
}

func NewManagerWithPath(path string) (*Manager, error) {
	m := &Manager{
		Path: path,
		Data: &ConfigData{
			Auths:     make(map[string]json.RawMessage),
			Providers: make(map[string]ProviderConfig),
		},
	}
	if err := m.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return m, nil
}

func (m *Manager) Load() error {
	b, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, m.Data)
}

func (m *Manager) Save() error {
	dir := filepath.Dir(m.Path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.Path, b, 0600)
}
