package api

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/decentralize-everything/indexer/store"
	"github.com/gin-gonic/gin"
)

func SetupRouter(db store.Database) *gin.Engine {
	r := gin.Default()

	r.GET("/api/v1/coins/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		ci, _ := db.GetCoinInfoById(id)
		c.JSON(http.StatusOK, gin.H{"result": ci != nil, "data": ci})
	})
	r.GET("/api/v1/coins", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("page_size", "10")
		sortedBy := c.DefaultQuery("sorted_by", "tx_count")
		dir := c.DefaultQuery("dir", "desc")
		listCoins(db, page, pageSize, sortedBy, dir, c)
	})
	r.GET("/api/v1/addresses/:address", func(c *gin.Context) {
		address := c.Params.ByName("address")
		coinBalances, _ := db.GetBalancesByAddress(address)
		c.JSON(http.StatusOK, gin.H{"result": coinBalances != nil, "data": coinBalances})
	})
	r.GET("/api/v1/addresses/:address/coins", func(c *gin.Context) {
		address := c.Params.ByName("address")
		coins, _ := db.GetCoinsByAddress(address)
		c.JSON(http.StatusOK, gin.H{"result": coins != nil, "data": coins})
	})

	return r
}

func listCoins(db store.Database, page string, pageSize string, sortedBy string, dir string, c *gin.Context) {
	p, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	size, err := strconv.Atoi(pageSize)
	if err != nil || size < 1 || size > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size, should be between 1 and 100"})
		return
	}

	if sortedBy != "tx_count" && sortedBy != "created_at" && sortedBy != "holder_count" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sorted_by, should be tx_count, created_at or holder_count"})
		return
	}

	if dir != "asc" && dir != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dir, should be asc or desc"})
		return
	}

	coins, _ := db.GetCoinInfos()
	start := (p - 1) * size
	end := start + size
	// Check if start and end are within bounds
	if start >= len(coins) {
		c.JSON(http.StatusOK, gin.H{"result": true, "data": nil})
		return
	}
	if end > len(coins) {
		end = len(coins)
	}

	var sortFunc func(i, j int) bool
	switch sortedBy {
	case "tx_count":
		sortFunc = func(i, j int) bool {
			if dir == "asc" {
				return coins[i].TxCount < coins[j].TxCount
			} else {
				return coins[i].TxCount > coins[j].TxCount
			}
		}
	case "created_at":
		sortFunc = func(i, j int) bool {
			if dir == "asc" {
				return coins[i].CreatedAt < coins[j].CreatedAt
			} else {
				return coins[i].CreatedAt > coins[j].CreatedAt
			}
		}
	case "holder_count":
		sortFunc = func(i, j int) bool {
			if dir == "asc" {
				return coins[i].HolderCount < coins[j].HolderCount
			} else {
				return coins[i].HolderCount > coins[j].HolderCount
			}
		}
	}

	sort.Slice(coins, sortFunc)
	c.JSON(http.StatusOK, gin.H{"result": true, "data": coins[start:end]})
}
