package model

import "time"

// 出价表
type Bid struct {
	ID        uint64 `gorm:"primaryKey"`
	AuctionID uint64 `gorm:"comment:拍卖ID"`
	Bidder    string `gorm:"size:42;comment:出价者"`
	Amount    string `gorm:"size:64;comment:出价金额(wei)"`
	BidTime   uint64
}

// ==================== 出价记录 ====================
func (m *DBModel) CreateBid(bid *Bid) error {
	bid.BidTime = uint64(time.Now().Unix())
	return m.db.Create(bid).Error
}

// 获取出价记录
func (m *DBModel) GetBidListByAuctionID(auctionID uint64) ([]Bid, error) {
	var list []Bid
	err := m.db.Where("auction_id = ?", auctionID).
		Order("id desc").
		Find(&list).Error
	return list, err
}
