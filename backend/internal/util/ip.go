package util

import (
	"net"
	"net/netip"

	"github.com/gin-gonic/gin"
)

func GetIPFromGinContext(c *gin.Context) *netip.Addr {
	addr16 := netip.AddrFrom16([16]byte(net.ParseIP(c.ClientIP())))
	if addr16.Is4() {

		addr4 := netip.AddrFrom4(addr16.As4())
		return &addr4
	} else {
		return nil
	}
}
