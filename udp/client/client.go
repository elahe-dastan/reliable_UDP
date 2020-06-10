package client

import (
	"fmt"
	"net"
	"reliable_UDP/message"
	"reliable_UDP/response"
	"strings"
)

type Client struct {
	IP		string
	Port 	int
	conn 	*net.UDPConn
}

func New(ip string, port int) Client {
	return Client{
		IP:   ip,
		Port: port,
	}
}

func (c *Client) Up() {
	addr := net.UDPAddr {
		IP:   net.ParseIP(c.IP),
		Port: c.Port,
	}

	_, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		fmt.Println(err)
	}

	cli, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.conn = cli

	m := make([]byte, 2048)

	for {
		_, remoteAddr, err := cli.ReadFromUDP(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		r := strings.Split(string(m), "\n")[0]

		r = strings.TrimSuffix(r, "\n")

		fmt.Println(r)

		res := message.Unmarshal(r)

		c.protocol(res, remoteAddr)
	}
}

func (c *Client) protocol(res response.Response, remoteAddr *net.UDPAddr) {
	switch t := res.(type) {
	case *response.StopWait:
		if s.waiting {
			// Add to prior list
			exists := false

			for _, ip := range s.prior {
				if ip == remoteAddr.String() {
					exists = true
					break
				}
			}

			if !exists {
				s.prior = append(s.prior, remoteAddr.String())
			}

			s.SWAddr = remoteAddr
			s.waiting = false

			s.seq = 0
			s.AskFile()
		}
	}
}