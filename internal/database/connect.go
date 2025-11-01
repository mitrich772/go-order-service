package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB подключается к базе данных PostgreSQL через GORM.
func ConnectDB(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=%s host=%s port=%s",
		cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.Host, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	log.Println("База данных подключена")
	return db
}

// Close закрывает соединение с базой данных
func Close(gormDB *gorm.DB) {
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Println("Ошибка получения *sql.DB:", err)
		return
	}
	sqlDB.Close()
}
