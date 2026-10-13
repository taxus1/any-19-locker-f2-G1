package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"somepro/internal/handler"
	"somepro/internal/middleware"
	"somepro/internal/repository"
	"somepro/internal/service"
)

// New 组装路由。验收测试契约无关：运行时用 gin 的 Routes() 扫路由，不写死类名/方法名。
func New(db *gorm.DB, rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.OperatorAuth())

	// Item 模块
	itemSvc := service.NewItem(repository.NewItem(db), rdb)
	itemH := handler.NewItem(itemSvc)

	// Locker 模块
	lockerCellRepo := repository.NewLockerCellRepository(db)
	lockerParcelRepo := repository.NewLockerParcelRepository(db)
	lockerSvc := service.NewLockerService(lockerCellRepo, lockerParcelRepo, rdb)
	lockerH := handler.NewLockerHandler(lockerSvc)

	// Demo API
	api := r.Group("/api/demo")
	{
		api.GET("/ping", itemH.Ping)
		api.POST("/item", itemH.CreateItem)
		api.GET("/list", itemH.List)
		api.GET("/cache", itemH.Cache)
	}

	// Locker API
	locker := r.Group("/api/locker")
	{
		locker.POST("/cell", lockerH.RegisterCell)
		locker.POST("/cell/disable", lockerH.DisableCell)
		locker.POST("/parcel", lockerH.StoreParcel)
		locker.POST("/parcel/pickup", lockerH.PickUpParcel)
		locker.GET("/parcel/list", lockerH.ListParcelsByPhone)
	}

	return r
}
