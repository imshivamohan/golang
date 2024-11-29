package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	customLogger "your_project/logger" // Replace with your actual project path
)

// Config represents the database configuration
type Config struct {
	Database struct {
		Driver    string `yaml:"driver"`
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		DBName    string `yaml:"dbname"`
		SSLMode   string `yaml:"sslmode"`  // PostgreSQL specific
		Filepath  string `yaml:"filepath"` // SQLite specific
		LogLevel  string `yaml:"log_level"`
		DBSchema  string `yaml:"dbschema"` // Database schema

		Pool struct {
			MaxOpenConns    int    `yaml:"max_open_conns"`
			MaxIdleConns    int    `yaml:"max_idle_conns"`
			ConnMaxLifetime string `yaml:"conn_max_lifetime"`
			ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
		} `yaml:"pool"`
	} `yaml:"database"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	return &config, nil
}

// InitDB initializes the database connection
func InitDB(configPath string) (*gorm.DB, *sql.DB, error) {
	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, nil, err
	}

	// Create logger
	zapLogger, zapGormLogger := customLogger.NewZapLogger()
	defer zapLogger.Sync() // Flush buffer, if any

	gormLogger := logger.New(
		zapGormLogger,
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.LogLevelValue(config.Database.LogLevel),
			Colorful:      false,
		},
	)

	// Choose the database driver dynamically
	var dialector gorm.Dialector
	switch config.Database.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
			config.Database.Host,
			config.Database.Port,
			config.Database.Username,
			config.Database.Password,
			config.Database.DBName,
			config.Database.SSLMode,
			config.Database.DBSchema,
		)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.DBName,
		)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(config.Database.Filepath)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver: %s", config.Database.Driver)
	}

	// Initialize GORM
	db, err := gorm.Open(dialector, &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	// Configure connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql.DB from GORM: %v", err)
	}

	// Parse connection pool durations
	connMaxLifetime, _ := time.ParseDuration(config.Database.Pool.ConnMaxLifetime)
	connMaxIdleTime, _ := time.ParseDuration(config.Database.Pool.ConnMaxIdleTime)

	sqlDB.SetMaxOpenConns(config.Database.Pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Database.Pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	return db, sqlDB, nil
}

// PingDB checks the database connection
func PingDB(sqlDB *sql.DB) error {
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}
	return nil
}

// GetDBStats retrieves database connection pool stats
func GetDBStats(sqlDB *sql.DB) sql.DBStats {
	return sqlDB.Stats()
}
