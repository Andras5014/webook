package main

import (
	"context"
	"fmt"
	"github.com/Andras5014/gohub/ioc"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	_ "gorm.io/driver/mysql"
	"net/http"
	"time"
)

func main() {
	initViper()

	app := InitApp()
	closeFunc := ioc.InitOtel()
	initPrometheus()
	for _, consumer := range app.Consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}

	}

	app.Cron.Start()
	defer func() {
		// 等待定时任务推出
		<-app.Cron.Stop().Done()
	}()

	server := app.Server
	server.Run(":8080")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// 实时监听配置变更
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
		fmt.Println("config file changed")
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
func initViperV1() {
	cflie := pflag.String("config", "config/config.yaml", "config file path")
	pflag.Parse()
	viper.SetConfigFile(*cflie)

	// 实时监听配置变更
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config file changed")
	})
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initViperRemote() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/gohub")
	if err != nil {
		return
	}
	err = viper.WatchRemoteConfig()
	if err != nil {
		return
	}
	if err := viper.ReadRemoteConfig(); err != nil {
		panic(err)
	}
}

func initLogger() {
	// 1. 配置日志
	// 2. 配置日志格式
	// 3. 配置日志输出
	// 4. 配置日志级别
	// 5. 配置日志文件
	// 6. 配置日志轮转
	// 7. 配置日志输出到控制台
	// 8. 配置日志输出到文件
	// 9. 配置日志输出到多个位置
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
}
