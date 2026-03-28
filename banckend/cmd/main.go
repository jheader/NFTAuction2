package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jheader/NFTAuction2/banckend/config"
	"github.com/jheader/NFTAuction2/banckend/internal/api"
	"github.com/jheader/NFTAuction2/banckend/internal/model"
	"github.com/jheader/NFTAuction2/banckend/internal/service"
)

func main() {

	//加载配置
	cfg := config.LoadConfig()
	// 初始化数据库

	log.Println(cfg)
	db, err := model.NewDBModel(cfg.MySQLDSN)

	if err != nil {
		panic(fmt.Sprintf("init db error: %v", err))
	}

	// 3. 启动链上事件监听

	eventService, err := service.NewAuctionEventService(cfg, db)
	if err != nil {
		panic(fmt.Sprintf("init event service error: %v", err))
	}
	_ = eventService.Start()

	// 4. 初始化 API 服务
	auctionService := service.NewAuctionService(cfg, db)

	auctionAPI := api.NewAuctionAPI(auctionService)

	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/auctions", auctionAPI.GetAuctionList)
		v1.GET("/auctions/:id", auctionAPI.GetAuctionDetail)
	}
	// 6. 启动服务
	fmt.Printf("服务启动: 0.0.0.0:%s\n", cfg.HTTPPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		fmt.Println("正在关闭服务...")
		eventService.Stop()
		os.Exit(0)
	}()
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		eventService.Stop()
		log.Fatalf("Failed to start server: %v", err)
	}
}
