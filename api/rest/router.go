package rest

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabex/db"
)

func Run(db db.Storage, host, port string, withUI bool) error {
	r := gin.Default()

	if withUI {
		r.Static("/ui", "./ui")
	}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	r.GET("/api/:channel/bytxid/:txid", bytxid(db))

	r.GET("/api/:channel/byblocknum/:blocknum", byblocknum(db))

	return r.Run(net.JoinHostPort(host, port))
}
