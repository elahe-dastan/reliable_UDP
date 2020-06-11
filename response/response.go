package response

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/elahe-dastan/reliable_UDP/message"
)

type Response interface {
	Marshal() string
}

type Size struct {
	Size int64
	Seq  int
}

type FileName struct {
	Name string
	Seq  int
}

type Segment struct {
	Part []byte
	Seq  int
}

func (s *Size) Marshal() string {
	fileSize := message.Size + "," + strconv.Itoa(s.Seq) + "," +
		strconv.FormatInt(s.Size, 10) + "\n"

	return fileSize
}

func (n *FileName) Marshal() string {
	fileName := message.FileName + "," + strconv.Itoa(n.Seq) + "," + n.Name + "\n"

	return fileName
}

func (s *Segment) Marshal() string {
	return fmt.Sprintf("%s,%d,%s\n", message.Segment, s.Seq, base64.StdEncoding.EncodeToString(s.Part))
}

func Unmarshal(s string) Response {
	s = strings.Split(s, "\n")[0]
	t := strings.Split(s, ",")

	switch t[0] {
	case message.Size:
		seq, _ := strconv.Atoi(t[1])
		size, _ := strconv.Atoi(t[2])
		size64 := int64(size)

		return &Size{
			Size: size64,
			Seq:  seq,
		}
	case message.FileName:
		seq, _ := strconv.Atoi(t[1])
		name := t[2]

		return &FileName{
			Name: name,
			Seq:  seq,
		}
	case message.Segment:
		seq, _ := strconv.Atoi(t[1])
		part, _ := base64.StdEncoding.DecodeString(t[2])

		return &Segment{
			Part: part,
			Seq:  seq,
		}
	}

	return nil
}
