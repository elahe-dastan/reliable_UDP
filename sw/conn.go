package sw

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elahe-dastan/reliable_UDP/request"
	"github.com/elahe-dastan/reliable_UDP/response"
)

type Conn struct {
	Host    string
	conn    net.Conn
	seq     int
	sndBuff []byte
	rcvBuff []byte
	sndLock sync.Mutex
	rcvLock sync.Mutex
	ack     chan int
}

func New(host string) Conn {
	return Conn{
		Host:    host,
		seq:     0,
		sndBuff: make([]byte, 0),
		rcvBuff: make([]byte, 0),
		ack:     make(chan int),
	}
}

func (c *Conn) Up() {
	host := strings.Split(c.Host, ":")
	ip := host[0]
	port, err := strconv.Atoi(host[1])
	if err != nil {
		fmt.Println(err)
	}

	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}

	_, err = net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		fmt.Println(err)
	}

	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.conn = ser

	go c.write()
	go c.read()
}

func (c *Conn) Connect(addr chan string) {
	for {
		cli, err := net.Dial("udp4", <-addr)
		if err != nil {
			fmt.Println(err)
			return
		}

		c.conn = cli

		fmt.Printf("The UDP server is %s\n", cli.RemoteAddr().String())

		go c.write()
		go c.read()
	}
}

func (c *Conn) Send(message []byte) {
	c.sndLock.Lock()

	c.sndBuff = append(c.sndBuff, message...)

	c.sndLock.Unlock()
}

func (c *Conn) write() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		s := make([]byte, 0)
		select {
		case <-ticker.C:
			c.sndLock.Lock()

			s = c.sndBuff
			if len(s) > 2048 {
				s = s[:2048]
			}

			c.sndLock.Unlock()

			d := response.Data{
				Data: s,
				Seq:  c.seq,
			}

			if len(s) == 0 {
				continue
			}

			_, err := c.conn.Write([]byte(d.Marshal()))
			if err != nil {
				log.Fatal(err)
			}
		case <-c.ack:
			c.sndBuff = c.sndBuff[len(s):]
		}
	}
}

func (c *Conn) Read(p []byte) (n int, err error) {
	min := len(p)
	if len(c.rcvBuff) < min {
		min = len(c.rcvBuff)
	}

	c.rcvLock.Lock()

	d := c.rcvBuff
	d = d[:min]

	c.rcvBuff = c.rcvBuff[min:]

	c.rcvLock.Unlock()

	return min, nil
}

func (c *Conn) read() {
	m := make([]byte, 4096)

	for {
		n, err := c.conn.Read(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		if n == 0 {
			continue
		}

		r := strings.Split(string(m), "\n")[0]

		r = strings.TrimSuffix(r, "\n")

		fmt.Println(r)

		res := response.Unmarshal(r)

		c.protocol(res)
	}
}

func (c *Conn) protocol(res response.Response) {
	switch t := res.(type) {
	case *response.Data:
		c.rcvLock.Lock()

		c.rcvBuff = append(c.rcvBuff, t.Data...)

		c.rcvLock.Unlock()

		go c.sendAck(t.Seq)
	case *response.Ack:
		c.alternateSeq(t.Seq)
	}
}

func (c *Conn) sendAck(seq int) {
	_, err := c.conn.Write([]byte((&request.Acknowledgment{Seq: seq}).Marshal()))
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Conn) alternateSeq(seq int) {
	if seq == c.seq {
		c.seq++
		c.seq %= 2
	}
}
