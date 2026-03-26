package main

import (
	"github.com/jheader/NFTAuction2/banckend/config"
	"github.com/jheader/NFTAuction2/banckend/model"
)

func main() {

	//加载配置
	cfg := config.LoadConfig()
	// 初始化数据库

	db, err := model.NewDBModel(cfg.MySQLDSN)

}
