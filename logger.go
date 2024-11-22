package logger

import (
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

// ZapLogger wraps Zap's SugaredLogger for use with GORM
type ZapLogger struct {
	sugarLogger *zap.SugaredLogger
}

// NewZapLogger creates a new Zap logger for structured logging
func NewZapLogger() (*zap.Logger, *ZapLogger) {
	zapLogger, _ := zap.NewProduction() // Use zap.NewDevelopment() for local development
	sugarLogger := zapLogger.Sugar()
	return zapLogger, &ZapLogger{sugarLogger: sugarLogger}
}

// LogMode implements the GORM logger interface
func (zl *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return zl
}

// Info logs info-level messages
func (zl *ZapLogger) Info(ctx context.Context, s string, args ...interface{}) {
	zl.sugarLogger.Infof(s, args...)
}

// Warn logs warn-level messages
func (zl *ZapLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	zl.sugarLogger.Warnf(s, args...)
}

// Error logs error-level messages
func (zl *ZapLogger) Error(ctx context.Context, s string, args ...interface{}) {
	zl.sugarLogger.Errorf(s, args...)
}

// Trace logs SQL queries with their execution time
func (zl *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		zl.sugarLogger.Errorf("SQL Error: %s | Duration: %s | Rows: %d | Error: %v", sql, elapsed, rows, err)
	} else {
		zl.sugarLogger.Infof("SQL Query: %s | Duration: %s | Rows: %d", sql, elapsed, rows)
	}
}
