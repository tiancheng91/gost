package gost

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/go-log/log"
)

var localIpv6 []*net.TCPAddr

func InitLocalAddress(interfaceName string) {
	interfaces, err := net.Interfaces()
	if err != nil || len(interfaces) == 0 {
		return
	}

	for _, iface := range interfaces {
		if iface.Name != interfaceName {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Failed to retrieve addresses for interface", iface.Name, ":", err)
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip.IsLoopback() || ip.To4() != nil || strings.Index(ip.String(), "fe80") == 0 {
				continue
			}

			log.Logf("load local ipv6 %s", ip.String())
			ipAddr, _ := net.ResolveTCPAddr("tcp", ip.String())
			localIpv6 = append(localIpv6, ipAddr)
		}
	}
}

func GetLocalAddress() *net.TCPAddr {
	rand.Seed(time.Now().UnixMilli())
	index := rand.Intn(len(localIpv6))
	return localIpv6[index]
}
