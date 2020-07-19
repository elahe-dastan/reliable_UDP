package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/elahe-dastan/reliable_UDP/sw"
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
		c := sw.New( "127.0.0.1:1999")
		addr := make(chan string, 1)
		go c.Connect(addr)
		addr <- "127.0.0.1:1995"
		c.Send([]byte("Hi, my name is raha"))
		select {}
	} else {
		s := sw.New("127.0.0.1:1995")
		s.Up()
		select {}
	}
}

