package duid

import (
	"fmt"
)

type Duid []byte

func (d Duid) String() string {
	if len(d) == 0 {
		return "<duid:nil>"
	}
	buf := make([]byte, 0, len(d)*3-1)
	for i, b := range d {
		if i > 0 {
			buf = append(buf, '-')
		}
		buf = append(buf, []byte(fmt.Sprintf("%02x", b))...)
	}

	return fmt.Sprintf("<duid:%s>", string(buf))
}
