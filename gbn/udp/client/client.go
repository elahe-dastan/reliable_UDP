package client

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/elahe-dastan/reliable_UDP/request"
	"github.com/elahe-dastan/reliable_UDP/response"
)

type Client struct {
	conn       net.Conn
	seq        int
	fileSize   int64
	fileName   string
	folder     string
	newFile    *os.File
	Fin        bool
	received   int
	windowSize int
}

func New(folder string) Client {
	return Client {
		seq:        0,
		folder:     folder,
		windowSize: 3,
	}
}

func (c *Client) Connect(addr chan string, name chan string) {
	for {
		cli, err := net.Dial("udp4", <-addr)
		if err != nil {
			fmt.Println(err)
			return
		}

		c.conn = cli

		_, err = cli.Write([]byte((&request.Get{Name: <-name}).Marshal()))
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("The UDP server is %s\n", cli.RemoteAddr().String())

		m := make([]byte, 2048)

		c.Fin = false
		c.received = 0

		for {
			if c.Fin {
				break
			}
			_, err := cli.Read(m)
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
}

func (c *Client) protocol(res response.Response) {
	switch t := res.(type) {
	case *response.Size:
		fmt.Println("received size the seq is")
		fmt.Println(t.Seq)

		if c.seqPlusOne(t.Seq) {
			c.fileSize = t.Size
		}

		go c.sendAck(c.seq)

	case *response.FileName:
		fmt.Println("received file name the seq is")
		fmt.Println(t.Seq)

		if c.seqPlusOne(t.Seq) {
			c.fileName = t.Name

			newFile, err := os.Create(filepath.Join(c.folder, filepath.Base("yep"+c.fileName)))
			if err != nil {
				fmt.Println(err)
			}

			c.newFile = newFile
		}

		go c.sendAck(c.seq)

	case *response.Segment:
		fmt.Println("received segment the seq is")
		fmt.Println(t.Seq)

		if c.seqPlusOne(t.Seq) {
			segment := t.Part

			received, err := c.newFile.Write(segment)
			if err != nil {
				fmt.Println(err)
			}

			c.received += received
			if int64(c.received) == c.fileSize {
				c.Fin = true
			}
		}

		go c.sendAck(t.Seq)
	}
}

func (c *Client) seqPlusOne(seq int) bool {
	if seq == c.seq {
		c.seq++
		c.seq %= c.windowSize

		return true
	}

	return false
}

func (c *Client) sendAck(seq int) {
	_, err := c.conn.Write([]byte((&request.Acknowledgment{Seq: seq}).Marshal()))
	if err != nil {
		fmt.Println(err)
	}
}

