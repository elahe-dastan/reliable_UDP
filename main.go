package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/elahe-dastan/reliable_UDP/gbn/udp/client"
	"github.com/elahe-dastan/reliable_UDP/gbn/udp/server"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	s, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}

	s = strings.TrimSuffix(s, "\n")

	a, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println(err)
	}

	if a == 1 {
		c := client.New( "/home/raha/Downloads")
		addr := make(chan string, 1)
		name := make(chan string)
		go c.Connect(addr, name)
		addr <- "127.0.0.1:1995"
		name <- "thesis.pdf"
		select {}
	} else {
		s := server.New("127.0.0.1:1995", "/home/raha/Downloads")
		s.Up()
	}
}

