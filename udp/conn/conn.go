package conn

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elahe-dastan/reliable_UDP/request"
	"github.com/elahe-dastan/reliable_UDP/response"
)

type Conn struct {
	Host       string
	conn       net.Conn
	seq        int
	//fileSize   int64
	//fileName   string
	//folder     string
	//newFile    *os.File
	//Fin        bool
	//received   int
	sndBuff []byte
	rcvBuff []byte
	sndLock sync.Mutex
	rcvLock sync.Mutex
	ack     chan int
}

func New(host string) Conn {
	return Conn {
		Host:host,
		seq:        0,
		//folder:     folder,
		sndBuff:    make([]byte, 0),
		rcvBuff:    make([]byte, 0),
		ack:        make(chan int),
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

	//m := make([]byte, 2048)
	//
	//for {
	//	_, remoteAddr, err := ser.ReadFromUDP(m)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//
	//	r := string(m)
	//
	//	fmt.Println(r)
	//
	//	req := request.Unmarshal(r)
	//
	//	s.protocol(req, remoteAddr)
	//}
}

func (c *Conn) Connect(addr chan string) {
	for {
		cli, err := net.Dial("udp4", <-addr)
		if err != nil {
			fmt.Println(err)
			return
		}

		c.conn = cli

		//_, err = cli.Write([]byte((&request.Get{Name: <-name}).Marshal()))
		//if err != nil {
		//	fmt.Println(err)
		//}

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

			d := response.Data {
				Data: s,
				Seq:  c.seq,
			}

			_, err := c.conn.Write([]byte(d.Marshal()))
			if err != nil {
				fmt.Println(err)
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
		_, err := c.conn.Read(m)
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

func (c *Conn) protocol(res response.Response) {
	switch t := res.(type) {
	case *response.Data:
		c.rcvLock.Lock()

		c.rcvBuff = append(c.rcvBuff, t.Data...)

		c.rcvLock.Unlock()

		go c.sendAck(t.Seq)
	case *response.Ack:
		c.alternateSeq(t.Seq)
	//case *response.Size:
	//	fmt.Println("received size the seq is")
	//	fmt.Println(t.Seq)
	//
	//	if c.alternateSeq(t.Seq) {
	//		c.fileSize = t.Size
	//	}
	//
	//	go c.sendAck(t.Seq)
	//
	//case *response.FileName:
	//	fmt.Println("received file name the seq is")
	//	fmt.Println(t.Seq)
	//
	//	if c.alternateSeq(t.Seq) {
	//		c.fileName = t.Name
	//
	//		newFile, err := os.Create(filepath.Join(c.folder, filepath.Base("yep"+c.fileName)))
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//
	//		c.newFile = newFile
	//	}
	//
	//	go c.sendAck(t.Seq)
	//
	//case *response.Segment:
	//	fmt.Println("received segment the seq is")
	//	fmt.Println(t.Seq)
	//
	//	if c.alternateSeq(t.Seq) {
	//		segment := t.Part
	//
	//		received, err := c.newFile.Write(segment)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//
	//		c.received += received
	//		if int64(c.received) == c.fileSize {
	//			c.Fin = true
	//		}
	//	}
	//
	//	go c.sendAck(t.Seq)
	}
}

func (c *Conn) sendAck(seq int) {
	_, err := c.conn.Write([]byte((&request.Acknowledgment{Seq: seq}).Marshal()))
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Conn) alternateSeq(seq int) {
	if seq == c.seq {
		c.seq++
		c.seq %= 2
	}
}
