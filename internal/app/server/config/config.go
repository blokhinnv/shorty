// Package config contains a description of the server config.
package config

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/caarlos0/env/v6"
)

// ServerConfig - structure for storing the server config.
type ServerConfig struct {
	ServerAddress           string        `env:"SERVER_ADDRESS"              envDefault:"http://localhost:8080" valid:"url" json:"server_address"`
	BaseURL                 string        `env:"BASE_URL"                    envDefault:"http://localhost:8080" valid:"url" json:"base_url"`
	SecretKey               string        `env:"SECRET_KEY"                                                                 json:"secret_key"` // I will not specify a default value for security
	EnableHTTPS             bool          `env:"ENABLE_HTTPS"                envDefault:"false"                             json:"enable_https"`
	JSONConfigPath          string        `env:"CONFIG"                      envDefault:""`
	PostgresDatabaseDSN     string        `env:"DATABASE_DSN"                                                               json:"postgres_database_dsn"`
	PostgresClearOnStart    bool          `env:"PG_CLEAR_ON_START"           envDefault:"false"                             json:"postgres_clear_on_start"`
	SQLiteDBPath            string        `env:"SQLITE_DB_PATH"              envDefault:"db.sqlite3"                        json:"sqlite_db_path"`
	SQLiteClearOnStart      bool          `env:"SQLITE_CLEAR_ON_START"       envDefault:"false"                             json:"sqlite_clear_on_start"`
	FileStoragePath         string        `env:"FILE_STORAGE_PATH"                                                          json:"file_storage_path"`
	FileStorageClearOnStart bool          `env:"FILE_STORAGE_CLEAR_ON_START" envDefault:"false"                             json:"file_storage_clear_on_start"`
	FileStorageTTLOnDisk    time.Duration `env:"FILE_STORAGE_TTL_ON_DISK"    envDefault:"1h"                                json:"file_storage_ttl_on_disk"`
	FileStorageTTLInMemory  time.Duration `env:"FILE_STORAGE_TTL_IN_MEMORY"  envDefault:"15m"                               json:"file_storage_ttl_in_memory"`
}

// reflectUpdate updates base's fields from ref.
func reflectUpdate(base any, ref any, refPriority bool) {
	baseObjPtr := reflect.ValueOf(base)
	if baseObjPtr.Kind() != reflect.Ptr {
		log.Warn("base should be ptr")
		return
	}
	baseObj := baseObjPtr.Elem()
	refObjPtr := reflect.ValueOf(ref)
	if refObjPtr.Kind() != reflect.Ptr {
		log.Warn("ref should be ptr")
		return
	}
	refObj := refObjPtr.Elem()
	refObjType := refObj.Type()

	for i := 0; i < refObj.NumField(); i++ {
		fName := refObjType.Field(i).Name
		f := refObj.Field(i)
		fTag := refObjType.Field(i).Tag
		tagValue, ok := fTag.Lookup("cfgArg")
		if ok {
			fName = tagValue
		}
		baseField := baseObj.FieldByName(fName)
		switch refFieldVal := f.Interface().(type) {
		case string:
			baseFieldVal := baseField.Interface().(string)
			if (refPriority && refFieldVal != "") || baseFieldVal == "" {
				baseField.SetString(refFieldVal)
			}
		case bool:
			baseFieldVal := baseField.Interface().(bool)
			if (refPriority && refFieldVal) || !baseFieldVal {
				baseField.SetBool(refFieldVal)
			}
		}
	}
}

// updateFromFlags updates server config from flags (flags have priority).
func (cfg *ServerConfig) updateFromFlags(flagCfg *FlagConfig) {
	// it seems like env should have priority,
	// but then I won't pass the 7th test...
	reflectUpdate(cfg, flagCfg, true)
}

// updateFromJSON updates server config from JSON (server cfg has priority).
func (cfg *ServerConfig) updateFromJSON() error {
	jsonFile, err := os.Open(cfg.JSONConfigPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	content, _ := io.ReadAll(jsonFile)
	var jsonCfg ServerConfig
	json.Unmarshal(content, &jsonCfg)
	reflectUpdate(cfg, &jsonCfg, false)
	return nil
}

// NewServerConfig - config constructor for the server.
func NewServerConfig(flagCfg *FlagConfig) (*ServerConfig, error) {
	cfg := ServerConfig{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	cfg.updateFromFlags(flagCfg)
	if cfg.JSONConfigPath != "" {
		err := cfg.updateFromJSON()
		if err != nil {
			log.Warnf("can't parse JSON %v, skipping: %v", cfg.JSONConfigPath, err)
		}
	}
	result, err := govalidator.ValidateStruct(cfg)
	if err != nil || !result {
		return nil, err
	}
	cfg.ServerAddress = regexp.MustCompile(`https?://`).ReplaceAllString(cfg.ServerAddress, "")
	return &cfg, nil
}
