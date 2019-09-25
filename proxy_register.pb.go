// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proxy_register.proto

package main

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
	Pubkey               string   `protobuf:"bytes,1,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
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

func (m *ProxyRegisterMsg) GetPubkey() string {
	if m != nil {
		return m.Pubkey
	}
	return ""
}

type ProxyRegisterResponse struct {
	Result               string   `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
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

func (m *ProxyRegisterResponse) GetResult() string {
	if m != nil {
		return m.Result
	}
	return ""
}

func init() {
	proto.RegisterType((*ProxyRegisterMsg)(nil), "peerv2.ProxyRegisterMsg")
	proto.RegisterType((*ProxyRegisterResponse)(nil), "peerv2.ProxyRegisterResponse")
}

func init() { proto.RegisterFile("proxy_register.proto", fileDescriptor_de8d6ae9acf7f7ea) }

var fileDescriptor_de8d6ae9acf7f7ea = []byte{
	// 161 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x29, 0x28, 0xca, 0xaf,
	0xa8, 0x8c, 0x2f, 0x4a, 0x4d, 0xcf, 0x2c, 0x2e, 0x49, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0x62, 0x2b, 0x48, 0x4d, 0x2d, 0x2a, 0x33, 0x52, 0xd2, 0xe2, 0x12, 0x08, 0x00, 0xc9, 0x07,
	0x41, 0xa5, 0x7d, 0x8b, 0xd3, 0x85, 0xc4, 0xb8, 0xd8, 0x0a, 0x4a, 0x93, 0xb2, 0x53, 0x2b, 0x25,
	0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xa0, 0x3c, 0x25, 0x7d, 0x2e, 0x51, 0x14, 0xb5, 0x41, 0xa9,
	0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x20, 0x0d, 0x45, 0xa9, 0xc5, 0xa5, 0x39, 0x25, 0x30, 0x0d,
	0x10, 0x9e, 0x51, 0x12, 0x97, 0x08, 0x8a, 0x86, 0xe0, 0xd4, 0xa2, 0xb2, 0xcc, 0xe4, 0x54, 0x21,
	0x2f, 0x2e, 0x5e, 0x14, 0x71, 0x21, 0x09, 0x3d, 0x88, 0x73, 0xf4, 0xd0, 0xdd, 0x22, 0x25, 0x8b,
	0x55, 0x06, 0x66, 0xb3, 0x12, 0x83, 0x13, 0x5b, 0x14, 0x4b, 0x6e, 0x62, 0x66, 0x5e, 0x12, 0x1b,
	0xd8, 0x5f, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x20, 0xe3, 0xd9, 0x66, 0xef, 0x00, 0x00,
	0x00,
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
