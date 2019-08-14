package msgtmpl

import (
	"net"
)

func (rx *tmpl) RequestRebinding(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(requestedIP, serverIdentifier, nil, nil)
}
