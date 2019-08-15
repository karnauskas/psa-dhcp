package msgtmpl

import (
	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func (rx *tmpl) Discover() []byte {
	return rx.request(dhcpmsg.MsgTypeDiscover, ipNone, ipBcast, nil, nil)
}
