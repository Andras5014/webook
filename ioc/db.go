package ioc

import (
	"github.com/Andras5014/webook/config"
	"github.com/Andras5014/webook/internal/repository/dao"
	"github.com/Andras5014/webook/pkg/logx"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
)

func InitDB(cfg *config.Config, l logx.Logger) *gorm.DB {

	db, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			LogLevel:                  glogger.Info,
			IgnoreRecordNotFoundError: true,
			SlowThreshold:             time.Millisecond * 10,
		}),
	})
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 10,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
	}))

	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logx.Field)

func (g gormLoggerFunc) Printf(msg string, fields ...interface{}) {
	g(msg, logx.Field{Key: "fields", Value: fields})
}
