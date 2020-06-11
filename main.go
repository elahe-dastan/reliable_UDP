package main

import (
	"bufio"
	"fmt"
	"os"
	"reliable_UDP/udp/client"
	"reliable_UDP/udp/server"
	"strconv"
	"strings"
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
		c := client.New("127.0.0.1", 1995, "thesis.pdf", "/home/raha/Downloads")
		c.Connect()
	} else {
		s := server.New("127.0.0.1", 1995, "/home/raha/Downloads")
		s.Up()
	}
}
