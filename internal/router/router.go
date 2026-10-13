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

	// 初始化快递柜相关组件
	lockerRepo := repository.NewLockerRepository(db)
	lockerSvc := service.NewLockerService(lockerRepo)
	lockerH := handler.NewLockerHandler(lockerSvc)

	// 初始化 demo 相关组件
	svc := service.NewItem(repository.NewItem(db), rdb)
	h := handler.NewItem(svc)

	api := r.Group("/api")
	{
		// demo 接口
		demo := api.Group("/demo")
		{
			demo.GET("/ping", h.Ping)
			demo.POST("/item", h.CreateItem)
			demo.GET("/list", h.List)
			demo.GET("/cache", h.Cache)
		}

		// 快递柜接口
		locker := api.Group("/locker")
		{
			// 格口管理
			locker.POST("/cell", lockerH.RegisterCell)
			locker.POST("/cell/disable", lockerH.DisableCell)

			// 包裹存取
			locker.POST("/parcel", lockerH.StoreParcel)
			locker.POST("/parcel/pickup", lockerH.PickupParcel)

			// 查询在柜包裹
			locker.GET("/parcel/list", lockerH.ListParcelsByPhone)
		}
	}
	return r
}
