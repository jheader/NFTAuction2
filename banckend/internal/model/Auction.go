package model

import (
	"time"
)

// 拍卖表
type Auction struct {
	ID            uint64 `gorm:"primaryKey"`
	AuctionID     uint64 `gorm:"uniqueIndex;comment:链上拍卖ID"`
	NftContract   string `gorm:"size:66;comment:NFT合约地址"`
	TokenID       uint64 `gorm:"comment:NFT TokenID"`
	Seller        string `gorm:"size:42;comment:卖家地址"`
	StartPrice    string `gorm:"size:64;comment:起拍价(wei)"`
	StartTime     uint64 `gorm:"comment:开始时间"`
	Duration      uint64 `gorm:"comment:时长(秒)"`
	HighestBid    string `gorm:"size:64;comment:最高出价(wei)"`
	HighestBidder string `gorm:"size:42;comment:最高出价者"`
	Status        string `gorm:"size:20;comment:PENDING/ACTIVE/ENDED/CANCELLED"`
	CreateTime    uint64
	UpdateTime    uint64
}

func (m *DBModel) CreateAuction(auction *Auction) (uint64, error) {

	auction.CreateTime = uint64(time.Now().Unix())
	auction.UpdateTime = uint64(time.Now().Unix())
	result := m.db.Create(auction)
	return auction.ID, result.Error
}

func (m *DBModel) UpdateHighestBid(auctionID uint64, highestBid string, bidder string) error {
	return m.db.Model(&Auction{}).
		Where("auction_id = ?", auctionID).
		Updates(map[string]interface{}{
			"highest_bid":    highestBid,
			"highest_bidder": bidder,
			"status":         "ACTIVE",
			"update_time":    time.Now().Unix(),
		}).Error
}

func (m *DBModel) UpdateAuctionStatus(auctionID uint64, status string) error {
	return m.db.Model(&Auction{}).
		Where("auction_id = ?", auctionID).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now().Unix(),
		}).Error
}

// 获取拍卖列表
func (m *DBModel) GetAuctionList(status string, offset int, limit int) ([]Auction, int64, error) {
	var list []Auction
	query := m.db.Model(&Auction{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	err := query.
		Order("auction_id desc").
		Offset(offset).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

// 根据拍卖ID获取详情
func (m *DBModel) GetAuctionByAuctionID(auctionID uint64) (*Auction, error) {
	var item Auction
	err := m.db.Where("auction_id = ?", auctionID).First(&item).Error
	return &item, err
}
