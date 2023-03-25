package config

// FlagConfig - structure for settings obtained from flags.
type FlagConfig struct {
	ServerAddress   string
	BaseURL         string
	EnableHTTPS     bool
	FileStoragePath string
	SecretKey       string
	DatabaseDSN     string
}
