// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proxy_register.proto

package process

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type ProxyRegisterMsg struct {
	CommitteePublicKey   string   `protobuf:"bytes,1,opt,name=CommitteePublicKey,proto3" json:"CommitteePublicKey,omitempty"`
	WantedMessages       []string `protobuf:"bytes,2,rep,name=WantedMessages,proto3" json:"WantedMessages,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProxyRegisterMsg) Reset()         { *m = ProxyRegisterMsg{} }
func (m *ProxyRegisterMsg) String() string { return proto.CompactTextString(m) }
func (*ProxyRegisterMsg) ProtoMessage()    {}
func (*ProxyRegisterMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_de8d6ae9acf7f7ea, []int{0}
}

func (m *ProxyRegisterMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProxyRegisterMsg.Unmarshal(m, b)
}
func (m *ProxyRegisterMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProxyRegisterMsg.Marshal(b, m, deterministic)
}
func (m *ProxyRegisterMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProxyRegisterMsg.Merge(m, src)
}
func (m *ProxyRegisterMsg) XXX_Size() int {
	return xxx_messageInfo_ProxyRegisterMsg.Size(m)
}
func (m *ProxyRegisterMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_ProxyRegisterMsg.DiscardUnknown(m)
}

var xxx_messageInfo_ProxyRegisterMsg proto.InternalMessageInfo

func (m *ProxyRegisterMsg) GetCommitteePublicKey() string {
	if m != nil {
		return m.CommitteePublicKey
	}
	return ""
}

func (m *ProxyRegisterMsg) GetWantedMessages() []string {
	if m != nil {
		return m.WantedMessages
	}
	return nil
}

type ProxyRegisterResponse struct {
	Pair                 []*MessageTopicPair `protobuf:"bytes,1,rep,name=Pair,proto3" json:"Pair,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *ProxyRegisterResponse) Reset()         { *m = ProxyRegisterResponse{} }
func (m *ProxyRegisterResponse) String() string { return proto.CompactTextString(m) }
func (*ProxyRegisterResponse) ProtoMessage()    {}
func (*ProxyRegisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_de8d6ae9acf7f7ea, []int{1}
}

func (m *ProxyRegisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProxyRegisterResponse.Unmarshal(m, b)
}
func (m *ProxyRegisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProxyRegisterResponse.Marshal(b, m, deterministic)
}
func (m *ProxyRegisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProxyRegisterResponse.Merge(m, src)
}
func (m *ProxyRegisterResponse) XXX_Size() int {
	return xxx_messageInfo_ProxyRegisterResponse.Size(m)
}
func (m *ProxyRegisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ProxyRegisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ProxyRegisterResponse proto.InternalMessageInfo

func (m *ProxyRegisterResponse) GetPair() []*MessageTopicPair {
	if m != nil {
		return m.Pair
	}
	return nil
}

type MessageTopicPair struct {
	Message              string   `protobuf:"bytes,1,opt,name=Message,proto3" json:"Message,omitempty"`
	Topic                string   `protobuf:"bytes,2,opt,name=Topic,proto3" json:"Topic,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MessageTopicPair) Reset()         { *m = MessageTopicPair{} }
func (m *MessageTopicPair) String() string { return proto.CompactTextString(m) }
func (*MessageTopicPair) ProtoMessage()    {}
func (*MessageTopicPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_de8d6ae9acf7f7ea, []int{2}
}

func (m *MessageTopicPair) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MessageTopicPair.Unmarshal(m, b)
}
func (m *MessageTopicPair) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MessageTopicPair.Marshal(b, m, deterministic)
}
func (m *MessageTopicPair) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageTopicPair.Merge(m, src)
}
func (m *MessageTopicPair) XXX_Size() int {
	return xxx_messageInfo_MessageTopicPair.Size(m)
}
func (m *MessageTopicPair) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageTopicPair.DiscardUnknown(m)
}

var xxx_messageInfo_MessageTopicPair proto.InternalMessageInfo

func (m *MessageTopicPair) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *MessageTopicPair) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func init() {
	proto.RegisterType((*ProxyRegisterMsg)(nil), "peerv2.ProxyRegisterMsg")
	proto.RegisterType((*ProxyRegisterResponse)(nil), "peerv2.ProxyRegisterResponse")
	proto.RegisterType((*MessageTopicPair)(nil), "peerv2.MessageTopicPair")
}

func init() { proto.RegisterFile("proxy_register.proto", fileDescriptor_de8d6ae9acf7f7ea) }

var fileDescriptor_de8d6ae9acf7f7ea = []byte{
	// 242 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0xc7, 0xed, 0x36, 0x2b, 0x7b, 0xa2, 0x8c, 0x47, 0x85, 0x20, 0x08, 0xa5, 0x07, 0xe9, 0x41,
	0x7a, 0xa8, 0xdf, 0x60, 0xe2, 0x45, 0x19, 0x94, 0x28, 0x08, 0x5e, 0x24, 0xad, 0x8f, 0x12, 0xb1,
	0x4d, 0x48, 0xe2, 0x70, 0xdf, 0x5e, 0x96, 0x65, 0x42, 0x4b, 0x8f, 0xef, 0xff, 0x7b, 0x21, 0xbf,
	0xf7, 0x87, 0x44, 0x1b, 0xf5, 0xbb, 0xfb, 0x30, 0xd4, 0x4a, 0xeb, 0xc8, 0x14, 0xda, 0x28, 0xa7,
	0x30, 0xd6, 0x44, 0x66, 0x5b, 0x66, 0x5f, 0xb0, 0xaa, 0xf6, 0x9c, 0x07, 0xbc, 0xb1, 0x2d, 0x16,
	0x80, 0x0f, 0xaa, 0xeb, 0xa4, 0x73, 0x44, 0xd5, 0x4f, 0xfd, 0x2d, 0x9b, 0x67, 0xda, 0xb1, 0x28,
	0x8d, 0xf2, 0x25, 0x9f, 0x20, 0x78, 0x0b, 0x97, 0x6f, 0xa2, 0x77, 0xf4, 0xb9, 0x21, 0x6b, 0x45,
	0x4b, 0x96, 0xcd, 0xd2, 0x79, 0xbe, 0xe4, 0xa3, 0x34, 0x7b, 0x84, 0xab, 0xc1, 0x5f, 0x9c, 0xac,
	0x56, 0xbd, 0x25, 0xbc, 0x83, 0x45, 0x25, 0xa4, 0x61, 0x51, 0x3a, 0xcf, 0xcf, 0x4b, 0x56, 0x1c,
	0xdc, 0x8a, 0xf0, 0xf0, 0x55, 0x69, 0xd9, 0xec, 0x39, 0xf7, 0x5b, 0xd9, 0x1a, 0x56, 0x63, 0x82,
	0x0c, 0xce, 0x42, 0x16, 0x3c, 0x8f, 0x23, 0x26, 0x70, 0xea, 0xd7, 0xd8, 0xcc, 0xe7, 0x87, 0xa1,
	0xac, 0x21, 0x19, 0xa8, 0xbc, 0x90, 0xd9, 0xca, 0x86, 0xf0, 0x09, 0x2e, 0x06, 0x39, 0xfe, 0xcb,
	0x8c, 0x5b, 0xba, 0xbe, 0x99, 0x24, 0xc7, 0x9b, 0xb2, 0x93, 0x75, 0xfc, 0xbe, 0xe8, 0x84, 0xec,
	0xeb, 0xd8, 0x37, 0x7e, 0xff, 0x17, 0x00, 0x00, 0xff, 0xff, 0xea, 0xc3, 0x56, 0x09, 0x89, 0x01,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ProxyRegisterServiceClient is the client API for ProxyRegisterService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ProxyRegisterServiceClient interface {
	//Unary
	ProxyRegister(ctx context.Context, in *ProxyRegisterMsg, opts ...grpc.CallOption) (*ProxyRegisterResponse, error)
}

type proxyRegisterServiceClient struct {
	cc *grpc.ClientConn
}

func NewProxyRegisterServiceClient(cc *grpc.ClientConn) ProxyRegisterServiceClient {
	return &proxyRegisterServiceClient{cc}
}

func (c *proxyRegisterServiceClient) ProxyRegister(ctx context.Context, in *ProxyRegisterMsg, opts ...grpc.CallOption) (*ProxyRegisterResponse, error) {
	out := new(ProxyRegisterResponse)
	err := c.cc.Invoke(ctx, "/peerv2.ProxyRegisterService/ProxyRegister", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyRegisterServiceServer is the server API for ProxyRegisterService service.
type ProxyRegisterServiceServer interface {
	//Unary
	ProxyRegister(context.Context, *ProxyRegisterMsg) (*ProxyRegisterResponse, error)
}

// UnimplementedProxyRegisterServiceServer can be embedded to have forward compatible implementations.
type UnimplementedProxyRegisterServiceServer struct {
}

func (*UnimplementedProxyRegisterServiceServer) ProxyRegister(ctx context.Context, req *ProxyRegisterMsg) (*ProxyRegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProxyRegister not implemented")
}

func RegisterProxyRegisterServiceServer(s *grpc.Server, srv ProxyRegisterServiceServer) {
	s.RegisterService(&_ProxyRegisterService_serviceDesc, srv)
}

func _ProxyRegisterService_ProxyRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProxyRegisterMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProxyRegisterServiceServer).ProxyRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peerv2.ProxyRegisterService/ProxyRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProxyRegisterServiceServer).ProxyRegister(ctx, req.(*ProxyRegisterMsg))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProxyRegisterService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "peerv2.ProxyRegisterService",
	HandlerType: (*ProxyRegisterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProxyRegister",
			Handler:    _ProxyRegisterService_ProxyRegister_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proxy_register.proto",
}
