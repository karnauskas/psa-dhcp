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

const (
	OptPadding            = 0
	OptSubnetMask         = 1
	OptRouter             = 3
	OptDNS                = 6
	OptDomainName         = 15
	OptBroadcastAddress   = 28
	OptIPAddressLeaseTime = 51
	OptMessageType        = 53
	OptServerIdentifier   = 54
	OptMessage            = 56
	OptRenewalTime        = 58
	OptRebindTime         = 59
	OptClientIdentifier   = 61
	OptEnd                = 255
)
