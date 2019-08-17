package dhcpmsg

const (
	OpRequest = 1
	OpReply   = 2
)

const (
	HtypeETHER = 1
)

const (
	FlagBroadcast = 1 << 15
)

const (
	DHCPCookie = 0x63825363
)

const (
	MsgTypeDiscover = 1
	MsgTypeOffer    = 2
	MsgTypeRequest  = 3
	MsgTypeAck      = 5
	MsgTypeNack     = 6
)

const (
	OptPadding            = 0
	OptSubnetMask         = 1
	OptRouter             = 3
	OptDNS                = 6
	OptHostname           = 12
	OptDomainName         = 15
	OptBroadcastAddress   = 28
	OptRequestedIP        = 50
	OptIPAddressLeaseTime = 51
	OptMessageType        = 53
	OptServerIdentifier   = 54
	OptParametersList     = 55
	OptMessage            = 56
	OptMaxMessageSize     = 57
	OptRenewalTime        = 58
	OptRebindTime         = 59
	OptClientIdentifier   = 61
	OptEnd                = 255
)
