package db

import (
	"fmt"
	"time"

	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *config.AppEnvVars) *gorm.DB {
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.PGHostIP, cfg.PGUser, cfg.PGPassword, cfg.PGDbName)
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger:          logger.Default.LogMode(logger.Info),
		CreateBatchSize: 1000,
	})
	if err != nil {
		fmt.Println(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Minute * 10)
	db.AutoMigrate(models.CompanyInfo{})

	return db
}
