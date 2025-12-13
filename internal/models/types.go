package models

// ServerConnection represents a Redis server connection configuration
type ServerConnection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password,omitempty"`
	Database int    `json:"database"`
	UseTLS   bool   `json:"use_tls"`
}

// RedisKey represents a key in Redis with its metadata
type RedisKey struct {
	Key  string
	Type string
	TTL  int64 // -1 for no expiry, -2 for key doesn't exist
}

// KeyValue represents a generic key-value pair
type KeyValue struct {
	Key   string
	Value string
}

// ScoredValue represents a value with score for sorted sets
type ScoredValue struct {
	Score  float64
	Member string
}

// ServerInfo holds Redis server information
type ServerInfo struct {
	Version          string
	Mode             string
	OS               string
	Uptime           int64
	ConnectedClients int64
	UsedMemory       int64
	UsedMemoryHuman  string
	UsedMemoryPeak   int64
	TotalKeys        int64
	ExpiredKeys      int64
	KeyspaceHits     int64
	KeyspaceMisses   int64
}

// ThemeName represents available theme options
type ThemeName string

const (
	ThemeDark      ThemeName = "dark"
	ThemeLight     ThemeName = "light"
	ThemeNord      ThemeName = "nord"
	ThemeDracula   ThemeName = "dracula"
	ThemeSolarized ThemeName = "solarized"
)

// AllThemes returns all available theme names
func AllThemes() []ThemeName {
	return []ThemeName{ThemeDark, ThemeLight, ThemeNord, ThemeDracula, ThemeSolarized}
}

// ThemeDisplayName returns a human-readable name for the theme
func (t ThemeName) DisplayName() string {
	switch t {
	case ThemeDark:
		return "Dark"
	case ThemeLight:
		return "Light"
	case ThemeNord:
		return "Nord"
	case ThemeDracula:
		return "Dracula"
	case ThemeSolarized:
		return "Solarized"
	default:
		return string(t)
	}
}
