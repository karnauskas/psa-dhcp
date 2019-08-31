package duid

import (
	"fmt"
)

type Duid []byte

func (d Duid) String() string {
	return fmt.Sprintf("duid(%x)", []byte(d))
}
