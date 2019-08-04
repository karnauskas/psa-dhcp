package dhcpmsg

const (
	OpRequest = 1
	OpReply   = 2
)

const (
	HtypeETHER   = 1
	HtypeIEEE802 = 6
)

const (
	FlagBroadcast = 1 << 15
)

const (
	DHCPCookie = 0x63825363
)
