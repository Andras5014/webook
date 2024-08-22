package ioc

import (
	"github.com/Andras5014/webook/config"
	"github.com/Andras5014/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) *gorm.DB {

	db, err := gorm.Open(mysql.Open(cfg.DB.DSN))
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
