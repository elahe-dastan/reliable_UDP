package client

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reliable_UDP/request"
	"reliable_UDP/response"
	"strings"
)

type Client struct {
	ServerIP   string
	ServerPort int
	conn       *net.UDPConn
	FileName   string
	seq        int
	fileSize   int64
	fileName   string
	folder     string
	newFile    *os.File
}

func New(ip string, port int, name string, folder string) Client {
	return Client{
		ServerIP:   ip,
		ServerPort: port,
		FileName:   name,
		seq:        0,
		folder:     folder,
	}
}

func (c *Client) Connect() {
	addr := net.UDPAddr{
		IP:   net.ParseIP(c.ServerIP),
		Port: c.ServerPort,
	}

	cli, err := net.DialUDP("udp4", nil, &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.conn = cli

	_, err = cli.Write([]byte((&request.Get{Name: c.FileName}).Marshal()))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("The UDP server is %s\n", cli.RemoteAddr().String())

	m := make([]byte, 2048)

	for {
		_, _, err := cli.ReadFromUDP(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		r := strings.Split(string(m), "\n")[0]

		r = strings.TrimSuffix(r, "\n")

		fmt.Println(r)

		res := response.Unmarshal(r)

		c.protocol(res)
	}
}

func (c *Client) protocol(res response.Response) {
	switch t := res.(type) {
	case *response.Size:
		fmt.Println("received size the seq is")
		fmt.Println(t.Seq)

		if c.alternateSeq(t.Seq) {
			c.fileSize = t.Size
		}

		go c.sendAck(t.Seq)

	case *response.FileName:
		fmt.Println("received file name the seq is")
		fmt.Println(t.Seq)

		if c.alternateSeq(t.Seq) {
			c.fileName = t.Name

			newFile, err := os.Create(filepath.Join(c.folder, filepath.Base("yep"+c.fileName)))
			if err != nil {
				fmt.Println(err)
			}

			c.newFile = newFile
		}

		go c.sendAck(t.Seq)

	case *response.Segment:
		fmt.Println("received segment the seq is")
		fmt.Println(t.Seq)

		if c.alternateSeq(t.Seq) {
			segment := t.Part

			_, err := c.newFile.Write(segment)
			if err != nil {
				fmt.Println(err)
			}
		}

		go c.sendAck(t.Seq)
	}
}

func (c *Client) sendAck(seq int) {
	_, err := c.conn.Write([]byte((&request.Acknowledgment{Seq: seq}).Marshal()))
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Client) alternateSeq(seq int) bool {
	if seq == c.seq {
		c.seq++
		c.seq %= 2

		return true
	}

	return false
}
