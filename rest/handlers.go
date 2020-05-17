package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/helpers"
	"net/http"
	"strconv"
)

func errorHandler(c *gin.Context, status int, err string) {
	c.JSON(status, gin.H{
		"error": err,
		"msg":   "",
	})
}

func bytxid(db db.Manager) func(c *gin.Context) {
	return func(c *gin.Context) {
		txid := c.Param("txid")
		QueryResults, err := db.GetByTxId(txid)
		if err != nil {
			errorHandler(c, http.StatusNotFound, err.Error())
		}

		blocks, err := helpers.PackTxsToBlocks(QueryResults)
		if err != nil {
			errorHandler(c, http.StatusInternalServerError, err.Error())
		}

		c.JSON(200, gin.H{
			"error": "",
			"msg":   blocks,
		})
	}
}

func byblocknum(db db.Manager) func(c *gin.Context) {
	return func(c *gin.Context) {
		blocknum := c.Param("blocknum")
		blocknumconverted, err := strconv.Atoi(blocknum)
		if err != nil {
			errorHandler(c, http.StatusInternalServerError, err.Error())
		}
		QueryResults, err := db.GetByBlocknum(uint64(blocknumconverted))
		if err != nil {
			errorHandler(c, http.StatusNotFound, err.Error())
		}

		blocks, err := helpers.PackTxsToBlocks(QueryResults)
		if err != nil {
			errorHandler(c, http.StatusInternalServerError, err.Error())
		}

		c.JSON(200, gin.H{
			"error": "",
			"msg":   blocks,
		})
	}
}

func bypayload(db db.Manager) func(c *gin.Context) {
	return func(c *gin.Context) {
		payloadkey := c.Param("payloadkey")
		QueryResults, err := db.GetBlockInfoByPayload(payloadkey)
		if err != nil {
			errorHandler(c, http.StatusNotFound, err.Error())
		}

		blocks, err := helpers.PackTxsToBlocks(QueryResults)
		if err != nil {
			errorHandler(c, http.StatusInternalServerError, err.Error())
		}

		c.JSON(200, gin.H{
			"error": "",
			"msg":   blocks,
		})
	}
}
