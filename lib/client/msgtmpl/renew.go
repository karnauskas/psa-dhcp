package msgtmpl

import (
	"net"
)

func (rx *tmpl) RequestRenewing(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(requestedIP, serverIdentifier, nil, nil)
}
