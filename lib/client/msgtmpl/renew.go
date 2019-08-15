package msgtmpl

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func (rx *tmpl) RequestRenewing(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(dhcpmsg.MsgTypeRequest, requestedIP, serverIdentifier, nil, nil)
}
