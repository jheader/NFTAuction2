package service

import (
	"strconv"

	"github.com/jheader/NFTAuction2/banckend/config"
	"github.com/jheader/NFTAuction2/banckend/internal/model"
)

type AuctionService struct {
	cfg *config.Config
	db  *model.DBModel
}

func NewAuctionService(cfg *config.Config, db *model.DBModel) *AuctionService {
	return &AuctionService{
		cfg: cfg,
		db:  db,
	}
}

// 获取拍卖列表
func (s *AuctionService) GetAuctionList(status string, pageStr string, sizeStr string) ([]model.Auction, int64, error) {
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size
	return s.db.GetAuctionList(status, offset, size)
}

// 获取拍卖详情
func (s *AuctionService) GetAuctionDetail(auctionIDStr string) (*model.Auction, error) {
	auctionID, err := strconv.ParseUint(auctionIDStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return s.db.GetAuctionByAuctionID(auctionID)
}

// 获取出价历史
func (s *AuctionService) GetBidHistory(auctionIDStr string) ([]model.Bid, error) {
	auctionID, err := strconv.ParseUint(auctionIDStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return s.db.GetBidListByAuctionID(auctionID)
}
