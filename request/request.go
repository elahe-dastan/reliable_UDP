package request

import (
	"fmt"
	"reliable_UDP/message"
	"strings"
)

type Request interface {
	Marshal() string
}

type Get struct {
	Name string
}

type Acknowledgment struct {
	Seq int
}

func (a *Acknowledgment) Marshal() string {
	return fmt.Sprintf("%s,%d\n", message.Ack, a.Seq)
}


func (g *Get) Marshal() string {
	return fmt.Sprintf("%s,%s\n", message.Get, g.Name)
}

func Unmarshal(req string) Request {
	req = strings.Split(req, "\n")[0]
	t := strings.Split(req, ",")

	switch t[0] {
	case message.Get:
		return &Get{Name:t[1]}
	}

	return nil
}