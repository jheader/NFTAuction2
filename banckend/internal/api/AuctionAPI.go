package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jheader/NFTAuction2/banckend/internal/service"
)

type AuctionAPI struct {
	auctionService *service.AuctionService
}

func NewAuctionAPI(a *service.AuctionService) *AuctionAPI {

	return &AuctionAPI{
		auctionService: a,
	}
}

func (a *AuctionAPI) GetAuctionList(c *gin.Context) {

	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	list, total, err := a.auctionService.GetAuctionList(status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"msg":   "获取拍卖列表失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"data":  list,
		"total": total,
	})
}

// GetAuctionDetail 获取拍卖详情
func (a *AuctionAPI) GetAuctionDetail(c *gin.Context) {
	auctionID := c.Param("id")

	detail, err := a.auctionService.GetAuctionDetail(auctionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取拍卖详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": detail,
	})
}

// GetBidHistory 获取出价历史
func (a *AuctionAPI) GetBidHistory(c *gin.Context) {
	auctionID := c.Param("id")

	list, err := a.auctionService.GetBidHistory(auctionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取出价历史失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": list,
	})
}
