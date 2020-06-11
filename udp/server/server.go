package server

import (
	"P2P/message"
	"fmt"
	"io"
	"net"
	"os"
	"reliable_UDP/request"
	"reliable_UDP/response"
	"time"
)

type Server struct {
	IP     string
	Port   int
	conn   *net.UDPConn
	folder string
	ack    chan int
}

func New(ip string, port int, folder string) Server {
	return Server{
		IP:     ip,
		Port:   port,
		folder: folder,
		ack:make(chan int),
	}
}

func (s *Server) Up() {
	addr := net.UDPAddr {
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
		fmt.Println("Recieved ack and the seq is")
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

	seq := 0

	fileSize := (&response.Size{Size: fileInfo.Size(), Seq: seq}).Marshal()

	_, err = s.conn.WriteToUDP([]byte(fileSize), remoteAddr)
	if err != nil {
		fmt.Println(err)
	}

	for {
		ticker := time.NewTicker(6 * time.Second)
		b := false

		select {
		case <-ticker.C:
			_, err = s.conn.WriteToUDP([]byte(fileSize), remoteAddr)
			if err != nil {
				fmt.Println(err)
			}
		case ack := <-s.ack:
			if ack == seq {
				seq += 1
				seq %= 2
				b = true
				break
			}
		}

		if b {
			break
		}
	}

	fileName := (&message.FileName{Name: fileInfo.Name(), Seq: seq}).Marshal()

	_, err = s.conn.WriteToUDP([]byte(fileName), s.SWAddr)
	if err != nil {
		fmt.Println(err)
	}

	for {
		ticker := time.NewTicker(6 * time.Second)
		b := false

		select {
		case <-ticker.C:
			_, err = s.conn.WriteToUDP([]byte(fileName), s.SWAddr)
			if err != nil {
				fmt.Println(err)
			}
		case ack := <-s.SWAck:
			if ack == seq {
				seq += 1
				seq %= 2
				b = true
				break
			}
		}

		if b {
			break
		}
	}

	sendBuffer := make([]byte, BUFFERSIZE-9)

	fmt.Println("Start sending file")

	for {
		read, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}

		sendBuff := sendBuffer[0:read]
		buffer := (&message.Segment{
			Part: sendBuff,
			Seq:  seq,
		}).Marshal()

		send := []byte(buffer)

		_, err = s.conn.WriteToUDP(send, s.SWAddr)
		if err != nil {
			fmt.Println(err)
		}

		for {
			ticker := time.NewTicker(6 * time.Second)
			b := false

			select {
			case <-ticker.C:
				_, err = s.conn.WriteToUDP(send, s.SWAddr)
				if err != nil {
					fmt.Println(err)
				}
			case ack := <-s.SWAck:
				if ack == seq {
					seq += 1
					seq %= 2
					b = true
					break
				}
			}

			if b {
				break
			}
		}
	}

	fmt.Println("File has been sent, closing connection!")
}