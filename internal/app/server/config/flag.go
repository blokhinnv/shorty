package config

// FlagConfig - структура для настроек, полученных из флагов.
type FlagConfig struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	SecretKey       string
	DatabaseDSN     string
}
