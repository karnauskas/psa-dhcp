package client

import (
	"context"

	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

func (mx *mclient) filterNetconfig(ctx context.Context, conf *libif.Ifconfig) {
	if conf == nil {
		return
	}

	if !mx.configureRoute {
		conf.Router = nil
	}
}
