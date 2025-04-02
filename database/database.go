package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "root:root@tcp(127.0.0.1:3307)/chat_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	log.Println("Database connected successfully!")

	db.AutoMigrate(&User{}, &Group{}, &GroupMember{})
}

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
	Online   bool
}

type Group struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique"`
}

type GroupMember struct {
	ID      uint `gorm:"primaryKey"`
	UserID  uint
	GroupID uint
}
