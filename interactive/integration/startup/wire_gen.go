// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/Andras5014/webook/interactive/grpc"
	"github.com/Andras5014/webook/interactive/repository"
	"github.com/Andras5014/webook/interactive/repository/cache"
	"github.com/Andras5014/webook/interactive/repository/dao"
	"github.com/Andras5014/webook/interactive/service"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitInteractiveSvc() service.InteractiveService {
	config := InitConfig()
	logger := InitLogger()
	db := InitDB(config, logger)
	interactiveDAO := dao.NewInteractiveDAO(db)
	cmdable := InitRedis(config)
	interactiveCache := cache.NewInteractiveCache(cmdable)
	interactiveRepository := repository.NewInteractiveRepository(interactiveDAO, interactiveCache, logger)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	return interactiveService
}

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	config := InitConfig()
	logger := InitLogger()
	db := InitDB(config, logger)
	interactiveDAO := dao.NewInteractiveDAO(db)
	cmdable := InitRedis(config)
	interactiveCache := cache.NewInteractiveCache(cmdable)
	interactiveRepository := repository.NewInteractiveRepository(interactiveDAO, interactiveCache, logger)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	return interactiveServiceServer
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitRedis,
	InitDB,
	InitConfig,
	InitLogger,
)

var interactiveSvcSet = wire.NewSet(service.NewInteractiveService, repository.NewInteractiveRepository, cache.NewInteractiveCache, dao.NewInteractiveDAO)
