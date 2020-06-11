package client

import (
	"P2P/message"
	"fmt"
	"net"
	"reliable_UDP/request"
	"reliable_UDP/response"
	"strings"
)

type Client struct {
	ServerIP   string
	ServerPort int
	conn    *net.UDPConn
	FileName   string
	seq        int
	fileSize  int64
}

func New(ip string, port int, name string) Client {
	return Client{
		ServerIP:   ip,
		ServerPort: port,
		FileName: name,
		seq:0,
	}
}

func (c *Client) Up() {
	addr := net.UDPAddr {
		IP:   net.ParseIP(c.ServerIP),
		Port: c.ServerPort,
	}

	cli, err := net.DialUDP("udp4", nil, &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.conn = cli

	_, err = cli.Write([]byte((&request.Get{Name:c.FileName}).Marshal()))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("The UDP server is %s\n", cli.RemoteAddr().String())

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

		res := response.Unmarshal(r)

		c.protocol(res, remoteAddr)
	}
}

func (c *Client) protocol(res response.Response, remoteAddr *net.UDPAddr) {
	switch t := res.(type) {
	case *response.Size:
		fmt.Println("recieved size the seq is")
		fmt.Println(t.Seq)
		if t.Seq == c.seq {
			c.seq += 1
			c.seq %= 2

			c.fileSize = t.Size
		}

		go c.sendAck(t.Seq)
	}
}

func (c *Client) sendAck(seq int) {
	_, err := c.conn.Write([]byte((&request.Acknowledgment{Seq:seq}).Marshal()))
	if err != nil {
		fmt.Println(err)
	}
}