package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"redis-explorer/internal/models"
)

// Config holds all application settings
type Config struct {
	Theme             models.ThemeName          `json:"theme"`
	Connections       []models.ServerConnection `json:"connections"`
	LastConnectionID  string                    `json:"last_connection_id,omitempty"`
	KeyScanCount      int                       `json:"key_scan_count"`
	AutoRefreshSecs   int                       `json:"auto_refresh_secs"`
	WindowWidth       float32                   `json:"window_width"`
	WindowHeight      float32                   `json:"window_height"`
}

var (
	instance *Config
	once     sync.Once
	mu       sync.RWMutex
	configPath string
)

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Theme: models.ThemeDark,
		Connections: []models.ServerConnection{
			{
				ID:       "default",
				Name:     "Local Redis",
				Host:     "localhost",
				Port:     6379,
				Database: 0,
				UseTLS:   false,
			},
		},
		LastConnectionID: "default",
		KeyScanCount:     100,
		AutoRefreshSecs:  0,
		WindowWidth:      1200,
		WindowHeight:     800,
	}
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "redis-explorer")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

// Load loads config from file or creates default
func Load() (*Config, error) {
	var loadErr error
	once.Do(func() {
		path, err := getConfigPath()
		if err != nil {
			loadErr = err
			return
		}
		configPath = path

		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				instance = DefaultConfig()
				loadErr = Save()
				return
			}
			loadErr = err
			return
		}

		instance = &Config{}
		if err := json.Unmarshal(data, instance); err != nil {
			instance = DefaultConfig()
			loadErr = Save()
			return
		}

		// Ensure defaults for missing fields
		if instance.KeyScanCount == 0 {
			instance.KeyScanCount = 100
		}
		if instance.WindowWidth == 0 {
			instance.WindowWidth = 1200
		}
		if instance.WindowHeight == 0 {
			instance.WindowHeight = 800
		}
		if len(instance.Connections) == 0 {
			instance.Connections = DefaultConfig().Connections
		}
	})
	return instance, loadErr
}

// Get returns the current config instance
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// Save saves the current config to file
func Save() error {
	mu.Lock()
	defer mu.Unlock()
	return saveWithoutLock()
}

// saveWithoutLock saves config (caller must hold lock)
func saveWithoutLock() error {
	if configPath == "" {
		path, err := getConfigPath()
		if err != nil {
			return err
		}
		configPath = path
	}

	data, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// SetTheme updates the theme setting
func SetTheme(theme models.ThemeName) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Theme = theme
	return saveWithoutLock()
}

// AddConnection adds a new server connection
func AddConnection(conn models.ServerConnection) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Connections = append(instance.Connections, conn)
	return saveWithoutLock()
}

// UpdateConnection updates an existing connection
func UpdateConnection(conn models.ServerConnection) error {
	mu.Lock()
	defer mu.Unlock()
	for i, c := range instance.Connections {
		if c.ID == conn.ID {
			instance.Connections[i] = conn
			break
		}
	}
	return saveWithoutLock()
}

// RemoveConnection removes a connection by ID
func RemoveConnection(id string) error {
	mu.Lock()
	defer mu.Unlock()
	for i, c := range instance.Connections {
		if c.ID == id {
			instance.Connections = append(instance.Connections[:i], instance.Connections[i+1:]...)
			break
		}
	}
	return saveWithoutLock()
}

// GetConnection returns a connection by ID
func GetConnection(id string) *models.ServerConnection {
	mu.RLock()
	defer mu.RUnlock()
	for _, c := range instance.Connections {
		if c.ID == id {
			conn := c // Create copy to avoid returning pointer to loop variable
			return &conn
		}
	}
	return nil
}

// SetLastConnection sets the last used connection ID
func SetLastConnection(id string) error {
	mu.Lock()
	defer mu.Unlock()
	instance.LastConnectionID = id
	return saveWithoutLock()
}

// SetWindowSize updates the window dimensions
func SetWindowSize(width, height float32) error {
	mu.Lock()
	defer mu.Unlock()
	instance.WindowWidth = width
	instance.WindowHeight = height
	return saveWithoutLock()
}
