package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/helpers"
)

func bytxid(db db.Storage) func(c *gin.Context) {
	return func(c *gin.Context) {
		txid := c.Param("txid")
		ch := c.Param("channel")
		QueryResults, err := db.GetByTxId(ch, txid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"msg":   nil,
			})
			return
		}

		if len(QueryResults) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no such data",
				"msg":   nil,
			})
			return
		}

		blocks, err := helpers.PackTxsToBlocks(QueryResults)
		if err != nil {
			if len(QueryResults) == 0 {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
					"msg":   nil,
				})
				return
			}
		}

		c.JSON(200, gin.H{
			"error": "",
			"msg":   blocks[0],
		})
	}
}

func byblocknum(db db.Storage) func(c *gin.Context) {
	return func(c *gin.Context) {
		blocknum := c.Param("blocknum")
		ch := c.Param("channel")
		blocknumconverted, err := strconv.Atoi(blocknum)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"msg":   nil,
			})
			return
		}

		QueryResults, err := db.GetByBlocknum(ch, uint64(blocknumconverted))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"msg":   nil,
			})
			return
		}

		if len(QueryResults) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no such data",
				"msg":   nil,
			})
			return
		}

		blocks, err := helpers.PackTxsToBlocks(QueryResults)
		if err != nil {
			if len(QueryResults) == 0 {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
					"msg":   nil,
				})
				return
			}
		}

		c.JSON(200, gin.H{
			"error": "",
			"msg":   blocks[0],
		})
	}
}
