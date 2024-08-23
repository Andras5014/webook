package ioc

import (
	"github.com/Andras5014/webook/config"
	"github.com/Andras5014/webook/internal/repository/dao"
	"github.com/Andras5014/webook/pkg/logger"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config, l logger.Logger) *gorm.DB {

	db, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			LogLevel:                  glogger.Info,
			IgnoreRecordNotFoundError: true,
			SlowThreshold:             time.Millisecond * 10,
		}),
	})
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, fields ...interface{}) {
	g(msg, logger.Field{Key: "fields", Value: fields})
}
