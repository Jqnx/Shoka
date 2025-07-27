package util

import (
	"net"
	"net/netip"

	"github.com/gin-gonic/gin"
)

func GetIPFromGinContext(c *gin.Context) netip.Addr {
	addr16 := netip.AddrFrom16([16]byte(net.ParseIP(c.ClientIP())))
	addr4 := netip.AddrFrom4(addr16.As4())
	return addr4
}
