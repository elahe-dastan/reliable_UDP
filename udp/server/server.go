package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"reliable_UDP/request"
	"reliable_UDP/response"
	"time"
)

// 1024 - 9
const BUFFERSIZE = 1015
const Periodic = 6

type Server struct {
	IP       string
	Port     int
	conn     *net.UDPConn
	folder   string
	ack      chan int
	seq      int
	periodic time.Duration
}

func New(ip string, port int, folder string) Server {
	return Server{
		IP:       ip,
		Port:     port,
		folder:   folder,
		ack:      make(chan int),
		seq:      0,
		periodic: Periodic * time.Second,
	}
}

func (s *Server) Up() {
	addr := net.UDPAddr{
		IP:   net.ParseIP(s.IP),
		Port: s.Port,
	}

	_, err := net.ResolveUDPAddr("udp", addr.String())
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
		_, remoteAddr, err := ser.ReadFromUDP(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		//r := strings.Split(string(m), "\n")[0]
		//
		//r = strings.TrimSuffix(r, "\n")

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
	case *request.Acknowledgment:
		s.ack <- t.Seq

		fmt.Println("Received ack and the seq is")
		fmt.Println(t.Seq)
	}
}

func (s *Server) send(name string, remoteAddr *net.UDPAddr) {
	fmt.Println("A stop and wait client has connected!")

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

	fileSize := (&response.Size{Size: fileInfo.Size(), Seq: s.seq}).Marshal()

	s.Write(fileSize, remoteAddr)

	fileName := (&response.FileName{Name: fileInfo.Name(), Seq: s.seq}).Marshal()

	s.Write(fileName, remoteAddr)

	sendBuffer := make([]byte, BUFFERSIZE)

	fmt.Println("Start sending file")

	for {
		read, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}

		sendBuff := sendBuffer[0:read]
		buffer := (&response.Segment{
			Part: sendBuff,
			Seq:  s.seq,
		}).Marshal()

		s.Write(buffer, remoteAddr)
	}

	fmt.Println("File has been sent, closing connection!")
}

func (s *Server) Write(message string, remoteAddr *net.UDPAddr) {
	_, err := s.conn.WriteToUDP([]byte(message), remoteAddr)
	if err != nil {
		fmt.Println(err)
	}

	for {
		ticker := time.NewTicker(s.periodic)
		b := false

		select {
		case <-ticker.C:
			_, err = s.conn.WriteToUDP([]byte(message), remoteAddr)
			if err != nil {
				fmt.Println(err)
			}
		case ack := <-s.ack:
			if ack == s.seq {
				s.seq++
				s.seq %= 2
				b = true

				break
			}
		}

		if b {
			break
		}
	}
}
