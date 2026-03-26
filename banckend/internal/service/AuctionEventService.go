package service

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jheader/NFTAuction2/banckend/config"
	"github.com/jheader/NFTAuction2/banckend/internal/model"
)

type AuctionEventService struct {
	client     *ethclient.Client
	auctionABI abi.ABI
	proxyAddr  common.Address
	ctx        context.Context
	cancel     context.CancelFunc
	db         *model.DBModel
}

func NewAuctionEventService(cfg *config.Config, db *model.DBModel) (*AuctionEventService, error) {

	// 连接以太坊
	client, err := ethclient.Dial(cfg.EthRPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial eth rpc: %v", err)
	}

	//加载合约
	abiBytes, err := os.ReadFile(cfg.ABIFilePath)
	if err != nil {
		return nil, fmt.Errorf("read abi file: %v", err)
	}

	auctionABI, err := abi.JSON(bytes.NewReader(abiBytes))
	if err != nil {
		return nil, fmt.Errorf("parse abi: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &AuctionEventService{
		client:     client,
		auctionABI: auctionABI,
		proxyAddr:  common.HexToAddress(cfg.AuctionProxy),
		ctx:        ctx,
		cancel:     cancel,
		db:         db,
	}, nil

}

// start 启动事件监听
func (s *AuctionEventService) Start() error {

	//订阅拍卖事件
	query := ethereum.FilterQuery{
		Addresses: []common.Address{s.proxyAddr},
	}

	logs := make(chan types.Log)

	sub, err := s.client.SubscribeFilterLogs(s.ctx, query, logs)

	if err != nil {
		return fmt.Errorf("subscribe logs: %v", err)

	}

	go func() {
		for {
			select {
			case err := <-sub.Err():
				fmt.Printf("subscription error: %v\n", err)
				return
			case vLog := <-logs:
				s.parseLog(vLog)
			}
		}
	}()
	fmt.Println("auction event service started")
	return nil

}

func (s *AuctionEventService) parseLog(log types.Log) {

	if len(log.Topics) > 0 {

		eventID := log.Topics[0].Hex()
		switch eventID {
		case s.auctionABI.Events["AuctionCreated"].ID.Hex():
			s.handleAuctionCreated(log)

		case s.auctionABI.Events["BidPlaced"].ID.Hex():
			s.handleBidPlaced(log)

		case s.auctionABI.Events["AuctionEnded"].ID.Hex():
			s.handleAuctionEnded(log)

		case s.auctionABI.Events["AuctionCancelled"].ID.Hex():
			s.handleAuctionCancelled(log)
		}
	}

}

func (s *AuctionEventService) handleAuctionCreated(log types.Log) {

	var event struct {
		Seller     common.Address
		StartPrice *big.Int
		StartTime  *big.Int
		Duration   *big.Int
	}

	err := s.auctionABI.UnpackIntoInterface(&event, "AuctionCreated", log.Data)
	if err != nil {
		fmt.Printf(" 解析AuctionCreated失败: %v\n", err)
		return
	}

	// 从topic解析索引字段
	auctionId := new(big.Int).SetBytes(log.Topics[1].Bytes())
	nftContract := common.HexToAddress(log.Topics[2].Hex())
	tokenId := new(big.Int).SetBytes(log.Topics[3].Bytes())

	fmt.Printf(" AuctionCreated | ID:%v | NFT:%s/%v | 卖家:%s | 起拍价:%s ETH\n",
		auctionId,
		nftContract.String(),
		tokenId,
		event.Seller.String(),
		weiToEth(event.StartPrice),
	)

	// 写入数据库
	auction := &model.Auction{
		AuctionID:     auctionId.Uint64(),
		NftContract:   nftContract.String(),
		TokenID:       tokenId.Uint64(),
		Seller:        event.Seller.String(),
		StartPrice:    event.StartPrice.String(),
		StartTime:     event.StartTime.Uint64(),
		Duration:      event.Duration.Uint64(),
		HighestBid:    "0",
		HighestBidder: "",
		Status:        "PENDING",
	}

	_, err = s.db.CreateAuction(auction)
	if err != nil {
		fmt.Printf("保存拍卖失败: %v\n", err)
	}

}

// 2. 出价
func (s *AuctionEventService) handleBidPlaced(log types.Log) {
	var event struct {
		Amount *big.Int
	}
	_ = s.auctionABI.UnpackIntoInterface(&event, "BidPlaced", log.Data)

	auctionId := new(big.Int).SetBytes(log.Topics[1].Bytes())
	bidder := common.HexToAddress(log.Topics[2].Hex())

	// 保存出价
	_ = s.db.CreateBid(&model.Bid{
		AuctionID: auctionId.Uint64(),
		Bidder:    bidder.String(),
		Amount:    event.Amount.String(),
	})

	// 更新最高价
	_ = s.db.UpdateHighestBid(auctionId.Uint64(), event.Amount.String(), bidder.String())
	fmt.Printf("新出价: %s | 拍卖ID: %d\n", bidder, auctionId.Uint64())
}

// 3. 拍卖结束
func (s *AuctionEventService) handleAuctionEnded(log types.Log) {
	auctionId := new(big.Int).SetBytes(log.Topics[1].Bytes())
	_ = s.db.UpdateAuctionStatus(auctionId.Uint64(), "ENDED")
	fmt.Printf("🏁 拍卖已结束: %d\n", auctionId.Uint64())
}

// 4. 拍卖取消
func (s *AuctionEventService) handleAuctionCancelled(log types.Log) {
	auctionId := new(big.Int).SetBytes(log.Topics[1].Bytes())
	_ = s.db.UpdateAuctionStatus(auctionId.Uint64(), "CANCELLED")
	fmt.Printf("拍卖已取消: %d\n", auctionId.Uint64())
}

func (s *AuctionEventService) Stop() {
	s.cancel()
	s.client.Close()
	fmt.Println("事件服务已停止")
}

// 工具函数
func weiToEth(wei *big.Int) string {
	eth := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(big.NewInt(1e18)))
	return eth.Text('f', 4)
}
