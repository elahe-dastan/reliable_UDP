package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elahe-dastan/reliable_UDP/request"
	"github.com/elahe-dastan/reliable_UDP/response"
)

// 1024 - 9.
const BUFFERSIZE = 1015
const Periodic = 10

type Server struct {
	Host       string
	conn       *net.UDPConn
	folder     string
	base       int
	nextSeq    int
	windowSize int
	periodic   time.Duration
	window     map[int]string
	ack        chan int
	fin        bool
}

func New(host string, folder string) Server {
	return Server{
		Host:       host,
		folder:     folder,
		base:       -1,
		nextSeq:    0,
		windowSize: 3,
		periodic:   Periodic * time.Second,
		window:     make(map[int]string),
		ack:        make(chan int),
		fin:        false,
	}
}

func (s *Server) Up() {
	host := strings.Split(s.Host, ":")
	ip := host[0]
	port, err := strconv.Atoi(host[1])
	if err != nil {
		fmt.Println(err)
	}

	addr := net.UDPAddr {
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

	s.conn = ser

	m := make([]byte, 2048)

	for {
		fmt.Println("reading")
		_, remoteAddr, err := ser.ReadFromUDP(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		r := string(m)

		fmt.Println(r)

		req := request.Unmarshal(r)

		s.protocol(req, remoteAddr)
	}
}

func (s *Server) protocol(req request.Request, remoteAddr *net.UDPAddr) {
	switch t := req.(type) {
	case *request.Get:
		go s.send(t.Name, remoteAddr)
		go s.acknowledgment(remoteAddr)
	case *request.Acknowledgment:
		s.ack <- t.Seq

		fmt.Println("Received ack and the seq is")
		fmt.Println(t.Seq)
	}
}

func (s *Server) send(name string, remoteAddr *net.UDPAddr) {
	fmt.Println("A gbn client has connected!")

	file, err := os.Open(s.folder + "/" + name)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Sending filename and filesize!")

	fileSize := (&response.Size{Size: fileInfo.Size(), Seq: s.nextSeq}).Marshal()

	s.Write(fileSize, remoteAddr)

	fileName := (&response.FileName{Name: fileInfo.Name(), Seq: s.nextSeq}).Marshal()

	s.Write(fileName, remoteAddr)

	sendBuffer := make([]byte, BUFFERSIZE)

	fmt.Println("Start sending file")

	for {
		read, err := file.Read(sendBuffer)
		if err == io.EOF {
			fmt.Println("break")
			break
		}

		sendBuff := sendBuffer[0:read]
		buffer := (&response.Segment{
			Part: sendBuff,
			Seq:  s.nextSeq,
		}).Marshal()

		s.Write(buffer, remoteAddr)
	}

	s.fin = true
	fmt.Println("File has been sent, closing connection!")
}

func (s *Server) Write(message string, remoteAddr *net.UDPAddr) {
	for {
		if (s.nextSeq+1)%s.windowSize != s.base {
			s.window[s.nextSeq] = message
			_, err := s.conn.WriteToUDP([]byte(message), remoteAddr)
			if err != nil {
				fmt.Println(err)
			}

			if s.base == -1 {
				s.base = s.nextSeq
			}
			s.nextSeq++
			s.nextSeq %= s.windowSize

			break
		}
	}
}

func (s *Server) acknowledgment(remoteAddr *net.UDPAddr) {
	for {
		ticker := time.NewTicker(s.periodic)

		select {
		case <-ticker.C:
			index := s.base
			for i := 0;i < s.windowSize; i++ {
				index += i
				index %= s.windowSize
				if index == s.nextSeq {
					break
				}
				_, err := s.conn.WriteToUDP([]byte(s.window[index]), remoteAddr)
				if err != nil {
					fmt.Println(err)
				}
			}
		case ack := <-s.ack:
			s.base = ack
			s.base++
			s.base %= s.windowSize
			if s.base == s.nextSeq {
				s.base = -1
			}

			break
		}

		//if s.fin && s.base == -1 {
		//	fmt.Println("ack break")
		//	break
		//}
	}
}
