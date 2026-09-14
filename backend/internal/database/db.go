// Package database 负责数据库连接、版本化迁移与演示种子数据。
// 三者职责分离：Open 只建立连接；Migrate 演进表结构；SeedDemo 仅在显式开启时写入演示数据。
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"sample-tracker/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 建立数据库连接并配置连接池。
// 启动阶段按配置重试，等待数据库就绪（compose 场景）；不再负责建表与种子数据。
func Open(ctx context.Context, cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	var db *gorm.DB
	var err error
	attempts := cfg.ConnectRetries
	if attempts < 1 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			break
		}
		slog.Warn("等待数据库就绪", "attempt", i+1, "max", attempts, "error", err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("等待数据库就绪时被取消: %w", ctx.Err())
		case <-time.After(cfg.ConnectRetryWait):
		}
	}
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败（已重试 %d 次）: %w", attempts, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层连接失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMin) * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("数据库 Ping 失败: %w", err)
	}
	return db, nil
}
