package response

import (
	"fmt"
	"reliable_UDP/message"
	"strings"
)

type Response interface {
	Marshal() string
}

type StopWait struct {

}

func (s *StopWait) Marshal() string {
	return fmt.Sprintf("%s\n", message.OK)
}

func Unmarshal(s string) Response {
	s = strings.Split(s, "\n")[0]
	t := strings.Split(s, ",")

	switch t[0] {
	case message.OK:
		return &StopWait{}
	}
}

