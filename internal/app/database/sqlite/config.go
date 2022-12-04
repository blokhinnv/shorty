package database

import "os"

type SQLiteConfig = struct {
	DBPath string
}

// Конструктор конфига SQLite на основе переменных окружения
func GetSQLiteConfig() SQLiteConfig {
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		panic("missing SQLITE_DB_PATH env variable")
	}
	return SQLiteConfig{dbPath}
}
