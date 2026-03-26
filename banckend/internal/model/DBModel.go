package model

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBModel 数据库模型操作类
type DBModel struct {
	db *gorm.DB // 持有 GORM 数据库连接
}

// NewDBModel 构造函数：初始化 DBModel
func NewDBModel(dsn string) (*DBModel, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	// 自动建表
	err = db.AutoMigrate(&Auction{}, &Bid{})
	if err != nil {
		return nil, err
	}
	return &DBModel{db: db}, nil
}
