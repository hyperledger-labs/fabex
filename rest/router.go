package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabex/db"
)

func Run(db db.Manager, port string, withUI bool) {
	r := gin.Default()

	if withUI {
		r.Static("/", "./ui")
	}

	r.GET("/bytxid/:txid", bytxid(db))

	r.GET("/byblocknum/:blocknum", byblocknum(db))

	r.GET("/bykey/:payloadkey", bypayload(db))

	r.Run(":" + port)
}
