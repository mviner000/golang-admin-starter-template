package project_name

import (
	"os"
	"strings"

	"github.com/mviner000/eyymi/config"
	"github.com/mviner000/eyymi/eyygo/shared"
)

var AppSettings SettingsStruct

type WebSocketConfig struct {
	Port string
}

type SettingsStruct struct {
	Database         shared.DatabaseConfig // Using shared.DatabaseConfig
	Debug            bool
	TimeZone         string
	WebSocket        WebSocketConfig
	CertFile         string
	KeyFile          string
	AllowedOrigins   []string
	TemplateBasePath string
	SecretKey        string
	LogFile          string
	InstalledApps    []string
	Environment      string
	IsDevelopment    bool
}

// LoadSettings initializes application settings
func LoadSettings() {
	dbConfig := shared.DatabaseConfig{ // Using shared.DatabaseConfig
		Engine:   os.Getenv("DB_ENGINE"),
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	debug := os.Getenv("DEBUG") == "true"

	// Initialize the shared config
	shared.SetSecretKey(os.Getenv("SECRET_KEY"))
	shared.SetDatabaseConfig(dbConfig) // This will work
	shared.SetDebug(debug)

	// Initialize the settings struct
	AppSettings = SettingsStruct{
		TemplateBasePath: os.Getenv("TEMPLATE_BASE_PATH"),
		WebSocket:        WebSocketConfig{Port: os.Getenv("WS_PORT")},
		AllowedOrigins:   strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		CertFile:         os.Getenv("CERT_FILE"),
		KeyFile:          os.Getenv("KEY_FILE"),
		LogFile:          os.Getenv("LOG_FILE"),
		Debug:            debug,
		TimeZone:         os.Getenv("TIME_ZONE"),
		InstalledApps:    strings.Split(os.Getenv("INSTALLED_APPS"), ","),
		Database:         dbConfig, // Using the same shared.DatabaseConfig
		Environment:      os.Getenv("ENVIRONMENT"),
		IsDevelopment:    os.Getenv("ENVIRONMENT") == "development",
	}

	// Log loaded settings
	config.LogStruct("Loaded settings", AppSettings)
}

func (s *SettingsStruct) GetDatabaseConfig() shared.DatabaseConfig {
	return s.Database
}

func (s *SettingsStruct) SetDatabaseConfig(dbConfig shared.DatabaseConfig) {
	s.Database = dbConfig
}

func (s *SettingsStruct) IsDebug() bool {
	return s.Debug
}

func (s *SettingsStruct) SetDebug(debug bool) {
	s.Debug = debug
}
