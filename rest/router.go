package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabex/db"
	"net/http"
)

func Run(db db.Storage, port string, withUI bool) {
	r := gin.Default()

	if withUI {
		r.Static("/ui", "./ui")
	}

	r.GET("/bytxid/:txid", bytxid(db))

	r.GET("/byblocknum/:blocknum", byblocknum(db))

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	r.Run(":" + port)
}
