package database

import (
	"fmt"
	"os"
	"time"

	"sample-tracker/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open() (*gorm.DB, error) {
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "tracker")
	pass := getenv("DB_PASSWORD", "tracker")
	name := getenv("DB_NAME", "tracker")
	sslmode := getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		host, port, user, pass, name, sslmode)

	var db *gorm.DB
	var err error
	// 等待数据库就绪（compose 场景）
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Location{},
		&model.DailySeq{},
		&model.Sample{},
		&model.Transfer{},
		&model.TestResult{},
		&model.Process{},
		&model.Machine{},
		&model.WaferLot{},
		&model.LotDailySeq{},
		&model.LotProcessRecord{},
		&model.MachineStatusLog{},
		&model.Wafer{},
		&model.WaferBinMap{},
	); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	seed(db)
	return db, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
