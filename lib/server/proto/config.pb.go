// Code generated by protoc-gen-go. DO NOT EDIT.
// source: lib/server/proto/config.proto

package config

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ServerConfig struct {
	// IP&cidr we are responsible for.
	Network string `protobuf:"bytes,1,opt,name=network,proto3" json:"network,omitempty"`
	// If set, restricts the range we use for dynamic IP assignment; must be within network.
	DynamicRange string `protobuf:"bytes,2,opt,name=dynamic_range,json=dynamicRange,proto3" json:"dynamic_range,omitempty"`
	// Validity of leases.
	LeaseDuration string `protobuf:"bytes,3,opt,name=lease_duration,json=leaseDuration,proto3" json:"lease_duration,omitempty"`
	// Domain name to announce.
	Domain string `protobuf:"bytes,4,opt,name=domain,proto3" json:"domain,omitempty"`
	// Router to announce.
	Router string `protobuf:"bytes,5,opt,name=router,proto3" json:"router,omitempty"`
	// List of DNS to announce.
	Dns []string `protobuf:"bytes,6,rep,name=dns,proto3" json:"dns,omitempty"`
	// List of NTP servers to announce.
	Ntp []string `protobuf:"bytes,7,rep,name=ntp,proto3" json:"ntp,omitempty"`
	// Disable dynamic configuration, only hand out IPs to staticly configured hosts.
	StaticOnly bool `protobuf:"varint,8,opt,name=static_only,json=staticOnly,proto3" json:"static_only,omitempty"`
	// Static hwaddr -> config mapping.
	Overrides            map[string]*ClientConfig `protobuf:"bytes,9,rep,name=overrides,proto3" json:"overrides,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ServerConfig) Reset()         { *m = ServerConfig{} }
func (m *ServerConfig) String() string { return proto.CompactTextString(m) }
func (*ServerConfig) ProtoMessage()    {}
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_495b121871ab1746, []int{0}
}

func (m *ServerConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServerConfig.Unmarshal(m, b)
}
func (m *ServerConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServerConfig.Marshal(b, m, deterministic)
}
func (m *ServerConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServerConfig.Merge(m, src)
}
func (m *ServerConfig) XXX_Size() int {
	return xxx_messageInfo_ServerConfig.Size(m)
}
func (m *ServerConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ServerConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ServerConfig proto.InternalMessageInfo

func (m *ServerConfig) GetNetwork() string {
	if m != nil {
		return m.Network
	}
	return ""
}

func (m *ServerConfig) GetDynamicRange() string {
	if m != nil {
		return m.DynamicRange
	}
	return ""
}

func (m *ServerConfig) GetLeaseDuration() string {
	if m != nil {
		return m.LeaseDuration
	}
	return ""
}

func (m *ServerConfig) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *ServerConfig) GetRouter() string {
	if m != nil {
		return m.Router
	}
	return ""
}

func (m *ServerConfig) GetDns() []string {
	if m != nil {
		return m.Dns
	}
	return nil
}

func (m *ServerConfig) GetNtp() []string {
	if m != nil {
		return m.Ntp
	}
	return nil
}

func (m *ServerConfig) GetStaticOnly() bool {
	if m != nil {
		return m.StaticOnly
	}
	return false
}

func (m *ServerConfig) GetOverrides() map[string]*ClientConfig {
	if m != nil {
		return m.Overrides
	}
	return nil
}

type ClientConfig struct {
	// Requests from this host will be ignored.
	Ignored bool `protobuf:"varint,1,opt,name=ignored,proto3" json:"ignored,omitempty"`
	// IP we will try to assign to this host.
	Ip string `protobuf:"bytes,2,opt,name=ip,proto3" json:"ip,omitempty"`
	// Netmask used during assignment.
	Netmask string `protobuf:"bytes,3,opt,name=netmask,proto3" json:"netmask,omitempty"`
	// Router to announce.
	Router string `protobuf:"bytes,4,opt,name=router,proto3" json:"router,omitempty"`
	// Hostname to give to client.
	Hostname string `protobuf:"bytes,5,opt,name=hostname,proto3" json:"hostname,omitempty"`
	// DNS to announce.
	Dns []string `protobuf:"bytes,6,rep,name=dns,proto3" json:"dns,omitempty"`
	// NTP servers to announce.
	Ntp                  []string `protobuf:"bytes,7,rep,name=ntp,proto3" json:"ntp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClientConfig) Reset()         { *m = ClientConfig{} }
func (m *ClientConfig) String() string { return proto.CompactTextString(m) }
func (*ClientConfig) ProtoMessage()    {}
func (*ClientConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_495b121871ab1746, []int{1}
}

func (m *ClientConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClientConfig.Unmarshal(m, b)
}
func (m *ClientConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClientConfig.Marshal(b, m, deterministic)
}
func (m *ClientConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClientConfig.Merge(m, src)
}
func (m *ClientConfig) XXX_Size() int {
	return xxx_messageInfo_ClientConfig.Size(m)
}
func (m *ClientConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ClientConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ClientConfig proto.InternalMessageInfo

func (m *ClientConfig) GetIgnored() bool {
	if m != nil {
		return m.Ignored
	}
	return false
}

func (m *ClientConfig) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *ClientConfig) GetNetmask() string {
	if m != nil {
		return m.Netmask
	}
	return ""
}

func (m *ClientConfig) GetRouter() string {
	if m != nil {
		return m.Router
	}
	return ""
}

func (m *ClientConfig) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *ClientConfig) GetDns() []string {
	if m != nil {
		return m.Dns
	}
	return nil
}

func (m *ClientConfig) GetNtp() []string {
	if m != nil {
		return m.Ntp
	}
	return nil
}

func init() {
	proto.RegisterType((*ServerConfig)(nil), "config.ServerConfig")
	proto.RegisterMapType((map[string]*ClientConfig)(nil), "config.ServerConfig.OverridesEntry")
	proto.RegisterType((*ClientConfig)(nil), "config.ClientConfig")
}

func init() { proto.RegisterFile("lib/server/proto/config.proto", fileDescriptor_495b121871ab1746) }

var fileDescriptor_495b121871ab1746 = []byte{
	// 355 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xcd, 0x8a, 0xdb, 0x30,
	0x14, 0x85, 0xb1, 0x9d, 0x38, 0xf6, 0xcd, 0x0f, 0x45, 0x94, 0x22, 0x02, 0xa5, 0x26, 0xa1, 0x60,
	0xba, 0x48, 0x20, 0xdd, 0x94, 0xee, 0x4a, 0xda, 0x75, 0x40, 0x7d, 0x80, 0xa0, 0xc4, 0x9a, 0x8c,
	0x88, 0x7d, 0x65, 0x64, 0x25, 0x83, 0x77, 0xf3, 0x3c, 0xf3, 0x94, 0x83, 0x25, 0x79, 0x26, 0xd9,
	0xcd, 0x4e, 0xe7, 0xbb, 0xc7, 0x96, 0xf9, 0xae, 0xe1, 0x6b, 0x29, 0x0f, 0xeb, 0x46, 0xe8, 0xab,
	0xd0, 0xeb, 0x5a, 0x2b, 0xa3, 0xd6, 0x47, 0x85, 0x0f, 0xf2, 0xb4, 0xb2, 0x81, 0xc4, 0x2e, 0x2d,
	0x9e, 0x23, 0x98, 0xfc, 0xb7, 0xad, 0xad, 0x05, 0x84, 0xc2, 0x08, 0x85, 0x79, 0x52, 0xfa, 0x4c,
	0x83, 0x2c, 0xc8, 0x53, 0xd6, 0x47, 0xb2, 0x84, 0x69, 0xd1, 0x22, 0xaf, 0xe4, 0x71, 0xaf, 0x39,
	0x9e, 0x04, 0x0d, 0xed, 0x7c, 0xe2, 0x21, 0xeb, 0x18, 0xf9, 0x0e, 0xb3, 0x52, 0xf0, 0x46, 0xec,
	0x8b, 0x8b, 0xe6, 0x46, 0x2a, 0xa4, 0x91, 0x6d, 0x4d, 0x2d, 0xfd, 0xeb, 0x21, 0xf9, 0x02, 0x71,
	0xa1, 0x2a, 0x2e, 0x91, 0x0e, 0xec, 0xd8, 0xa7, 0x8e, 0x6b, 0x75, 0x31, 0x42, 0xd3, 0xa1, 0xe3,
	0x2e, 0x91, 0x4f, 0x10, 0x15, 0xd8, 0xd0, 0x38, 0x8b, 0xf2, 0x94, 0x75, 0xc7, 0x8e, 0xa0, 0xa9,
	0xe9, 0xc8, 0x11, 0x34, 0x35, 0xf9, 0x06, 0xe3, 0xc6, 0x70, 0x23, 0x8f, 0x7b, 0x85, 0x65, 0x4b,
	0x93, 0x2c, 0xc8, 0x13, 0x06, 0x0e, 0xed, 0xb0, 0x6c, 0xc9, 0x1f, 0x48, 0xd5, 0x55, 0x68, 0x2d,
	0x0b, 0xd1, 0xd0, 0x34, 0x8b, 0xf2, 0xf1, 0x66, 0xb9, 0xf2, 0x56, 0x6e, 0x1d, 0xac, 0x76, 0x7d,
	0xeb, 0x1f, 0x1a, 0xdd, 0xb2, 0xf7, 0xa7, 0xe6, 0x0c, 0x66, 0xf7, 0xc3, 0xee, 0x3b, 0xce, 0xa2,
	0xf5, 0xae, 0xba, 0x23, 0xf9, 0x01, 0xc3, 0x2b, 0x2f, 0x2f, 0xce, 0xcf, 0x78, 0xf3, 0xb9, 0xbf,
	0x62, 0x5b, 0x4a, 0x81, 0xc6, 0x5d, 0xc1, 0x5c, 0xe5, 0x77, 0xf8, 0x2b, 0x58, 0xbc, 0x04, 0x30,
	0xb9, 0x9d, 0x75, 0x2b, 0x90, 0x27, 0x54, 0x5a, 0x14, 0xf6, 0xb5, 0x09, 0xeb, 0x23, 0x99, 0x41,
	0x28, 0x6b, 0xef, 0x3d, 0x94, 0xb5, 0x5f, 0x56, 0xc5, 0x9b, 0xb3, 0xd7, 0xdc, 0xc7, 0x1b, 0x91,
	0x83, 0x3b, 0x91, 0x73, 0x48, 0x1e, 0x55, 0x63, 0x90, 0x57, 0xc2, 0x2b, 0x7e, 0xcb, 0x1f, 0x91,
	0x7c, 0x88, 0xed, 0xef, 0xf3, 0xf3, 0x35, 0x00, 0x00, 0xff, 0xff, 0x41, 0xea, 0x25, 0x4a, 0x5f,
	0x02, 0x00, 0x00,
}