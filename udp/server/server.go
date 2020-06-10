package server

import (
	"fmt"
	"net"
	"reliable_UDP/message"
	"reliable_UDP/request"
	"reliable_UDP/response"
	"strings"
)

type Server struct {
	IP 		string
	Port	int
	conn 	*net.UDPConn
}

func New(ip string, port int) Server {
	return Server{
		IP:   ip,
		Port: port,
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

		r := strings.Split(string(m), "\n")[0]

		r = strings.TrimSuffix(r, "\n")

		fmt.Println(r)

		req := response.Unmarshal(r)

		s.protocol(req, remoteAddr)
	}
}

func (s *Server) protocol(req request.Request, remoteAddr *net.UDPAddr) {
	switch t := req.(type) {
	}
}