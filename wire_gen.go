// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/Andras5014/gohub/interactive/events"
	repository2 "github.com/Andras5014/gohub/interactive/repository"
	cache2 "github.com/Andras5014/gohub/interactive/repository/cache"
	dao2 "github.com/Andras5014/gohub/interactive/repository/dao"
	service2 "github.com/Andras5014/gohub/interactive/service"
	article3 "github.com/Andras5014/gohub/internal/events/article"
	"github.com/Andras5014/gohub/internal/repository"
	article2 "github.com/Andras5014/gohub/internal/repository/article"
	"github.com/Andras5014/gohub/internal/repository/cache"
	"github.com/Andras5014/gohub/internal/repository/dao"
	"github.com/Andras5014/gohub/internal/repository/dao/article"
	"github.com/Andras5014/gohub/internal/service"
	article4 "github.com/Andras5014/gohub/internal/web/handler/article"
	"github.com/Andras5014/gohub/internal/web/handler/oauth2"
	"github.com/Andras5014/gohub/internal/web/handler/user"
	"github.com/Andras5014/gohub/internal/web/jwt"
	"github.com/Andras5014/gohub/ioc"
	"github.com/google/wire"
)

import (
	_ "github.com/spf13/viper/remote"
	_ "gorm.io/driver/mysql"
)

// Injectors from wire.go:

func InitApp() *App {
	config := ioc.InitConfig()
	cmdable := ioc.InitRedis(config)
	limiter := ioc.InitLimiter(cmdable)
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitMiddlewares(limiter, handler, logger)
	db := ioc.InitDB(config, logger)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, logger)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := user.NewUserHandler(userService, codeService, handler, logger)
	oauth2Service := ioc.InitOAuth2WeChatService(logger)
	weChatHandler := oauth2.NewOAuth2WeChatHandler(oauth2Service, userService, handler)
	articleDAO := article.NewArticleDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(articleDAO, articleCache, logger)
	client := ioc.InitKafka(config)
	syncProducer := ioc.InitSyncProducer(client)
	producer := article3.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, logger)
	interactiveDAO := dao2.NewInteractiveDAO(db)
	interactiveCache := cache2.NewInteractiveCache(cmdable)
	interactiveRepository := repository2.NewInteractiveRepository(interactiveDAO, interactiveCache, logger)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	interactiveServiceClient := ioc.InitInteractiveGrpcClient(interactiveService, config)
	articleHandler := article4.NewArticleHandler(articleService, interactiveServiceClient, logger)
	engine := ioc.InitWebServer(v, userHandler, weChatHandler, articleHandler)
	interactiveReadEventBatchConsumer := events.NewInteractiveReadEventBatchConsumer(client, interactiveRepository, logger)
	v2 := ioc.InitConsumers(interactiveReadEventBatchConsumer)
	rankingService := service.NewRankingService(articleService, interactiveServiceClient)
	universalClient := ioc.InitRedisUniversalClient(config)
	redsync := ioc.InitRedSync(universalClient)
	rankingJob := ioc.InitRankingJob(rankingService, redsync, logger)
	cron := ioc.InitJobs(rankingJob, logger)
	app := &App{
		Server:    engine,
		Consumers: v2,
		Cron:      cron,
	}
	return app
}

// wire.go:

var rankingSvcSet = wire.NewSet(cache.NewRedisRankingCache, repository.NewRankingRepository, service.NewRankingService)

var interactiveSvcSet = wire.NewSet(service2.NewInteractiveService, repository2.NewInteractiveRepository, cache2.NewInteractiveCache, dao2.NewInteractiveDAO)

var userSvcSet = wire.NewSet(service.NewUserService, repository.NewUserRepository, cache.NewUserCache, dao.NewUserDAO)

var articleSvcSet = wire.NewSet(service.NewArticleService, article2.NewArticleRepository, article.NewArticleDAO, cache.NewRedisArticleCache)

var codeSvcProvider = wire.NewSet(cache.NewCodeCache, repository.NewCodeRepository, service.NewCodeService)

var thirdPartySet = wire.NewSet(ioc.InitConfig, ioc.InitLogger, ioc.InitDB, ioc.InitRedis, ioc.InitRedisUniversalClient, ioc.InitRedSync, ioc.InitSmsService, ioc.InitKafka, ioc.InitSyncProducer, ioc.InitConsumers)
