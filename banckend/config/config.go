package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	//数据库配置
	MySQLDSN string
	RedisURL string

	//区块连配置
	EthRPCURL    string
	AuctionProxy string
	ABIFilePath  string
	// HTTP服务配置
	HTTPPort string
}

func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: No .env file found, using system environment variables: %v", err)
	}

	return &Config{
		MySQLDSN:     os.Getenv("MYSQL_DSN"),
		RedisURL:     os.Getenv("REDIS_URL"),
		EthRPCURL:    os.Getenv("ETH_RPC_URL"),
		AuctionProxy: os.Getenv("AUCTION_PROXY"),
		ABIFilePath:  os.Getenv("ABI_FILE_PATH"),
		HTTPPort:     os.Getenv("HTTP_PORT"),
	}

}
