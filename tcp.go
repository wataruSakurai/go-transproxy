package transproxy

import (
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/proxy"
)

type TCPProxy struct {
	TCPProxyConfig
}

type TCPProxyConfig struct {
	ListenAddress string
	NoProxy       NoProxy
}

func NewTCPProxy(c TCPProxyConfig) *TCPProxy {
	return &TCPProxy{
		TCPProxyConfig: c,
	}
}

func (s TCPProxy) Start() error {
	//pdialer := proxy.FromEnvironment()

	dialer := &net.Dialer{
		KeepAlive: 3 * time.Minute,
		DualStack: true,
	}
	u, err := url.Parse(os.Getenv("http_proxy"))
	if err != nil {
		return err
	}

	pdialer, err := proxy.FromURL(u, dialer)
	if err != nil {
		return err
	}

	npdialer := proxy.Direct

	log.Infof("TCP-Proxy: Start listener on %s", s.ListenAddress)

	go func() {
		ListenTCP(s.ListenAddress, func(tc *TCPConn) {
			var destConn net.Conn
			// TODO Convert OrigAddr to domain and check useProxy with domain too?
			if useProxy(s.NoProxy, strings.Split(tc.OrigAddr, ":")[0]) {

				destConn, err = pdialer.Dial("tcp", tc.OrigAddr)
			} else {
				destConn, err = npdialer.Dial("tcp", tc.OrigAddr)
			}

			if err != nil {
				log.Errorf("TCP-Proxy: Failed to connect to destination - %s", err.Error())
				return
			}

			Pipe(tc, destConn)
		})
	}()

	return nil
}
