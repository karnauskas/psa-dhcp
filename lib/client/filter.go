package client

import (
	"context"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/libif"
)

// filterNetconfig modifies the netconfig to align with the server configuration.
func (mx *mclient) filterNetconfig(ctx context.Context, conf *libif.Ifconfig) {
	if conf == nil {
		return
	}

	if !mx.configureRoute {
		conf.Router = nil
	}
}
