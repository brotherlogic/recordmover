// Code generated by protoc-gen-go. DO NOT EDIT.
// source: recordmover.proto

package recordprocessor

import (
	fmt "fmt"
	proto1 "github.com/brotherlogic/recordcollection/proto"
	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Context struct {
	Before               *proto1.Record `protobuf:"bytes,1,opt,name=before,proto3" json:"before,omitempty"`
	Location             string         `protobuf:"bytes,2,opt,name=location,proto3" json:"location,omitempty"`
	After                *proto1.Record `protobuf:"bytes,3,opt,name=after,proto3" json:"after,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Context) Reset()         { *m = Context{} }
func (m *Context) String() string { return proto.CompactTextString(m) }
func (*Context) ProtoMessage()    {}
func (*Context) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{0}
}

func (m *Context) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Context.Unmarshal(m, b)
}
func (m *Context) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Context.Marshal(b, m, deterministic)
}
func (m *Context) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Context.Merge(m, src)
}
func (m *Context) XXX_Size() int {
	return xxx_messageInfo_Context.Size(m)
}
func (m *Context) XXX_DiscardUnknown() {
	xxx_messageInfo_Context.DiscardUnknown(m)
}

var xxx_messageInfo_Context proto.InternalMessageInfo

func (m *Context) GetBefore() *proto1.Record {
	if m != nil {
		return m.Before
	}
	return nil
}

func (m *Context) GetLocation() string {
	if m != nil {
		return m.Location
	}
	return ""
}

func (m *Context) GetAfter() *proto1.Record {
	if m != nil {
		return m.After
	}
	return nil
}

type RecordMove struct {
	InstanceId           int32          `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	FromFolder           int32          `protobuf:"varint,2,opt,name=from_folder,json=fromFolder,proto3" json:"from_folder,omitempty"`
	ToFolder             int32          `protobuf:"varint,3,opt,name=to_folder,json=toFolder,proto3" json:"to_folder,omitempty"`
	MoveDate             int64          `protobuf:"varint,4,opt,name=move_date,json=moveDate,proto3" json:"move_date,omitempty"`
	Record               *proto1.Record `protobuf:"bytes,5,opt,name=record,proto3" json:"record,omitempty"`
	BeforeContext        *Context       `protobuf:"bytes,6,opt,name=before_context,json=beforeContext,proto3" json:"before_context,omitempty"`
	AfterContext         *Context       `protobuf:"bytes,7,opt,name=after_context,json=afterContext,proto3" json:"after_context,omitempty"`
	LastUpdate           int64          `protobuf:"varint,8,opt,name=last_update,json=lastUpdate,proto3" json:"last_update,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *RecordMove) Reset()         { *m = RecordMove{} }
func (m *RecordMove) String() string { return proto.CompactTextString(m) }
func (*RecordMove) ProtoMessage()    {}
func (*RecordMove) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{1}
}

func (m *RecordMove) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RecordMove.Unmarshal(m, b)
}
func (m *RecordMove) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RecordMove.Marshal(b, m, deterministic)
}
func (m *RecordMove) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RecordMove.Merge(m, src)
}
func (m *RecordMove) XXX_Size() int {
	return xxx_messageInfo_RecordMove.Size(m)
}
func (m *RecordMove) XXX_DiscardUnknown() {
	xxx_messageInfo_RecordMove.DiscardUnknown(m)
}

var xxx_messageInfo_RecordMove proto.InternalMessageInfo

func (m *RecordMove) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

func (m *RecordMove) GetFromFolder() int32 {
	if m != nil {
		return m.FromFolder
	}
	return 0
}

func (m *RecordMove) GetToFolder() int32 {
	if m != nil {
		return m.ToFolder
	}
	return 0
}

func (m *RecordMove) GetMoveDate() int64 {
	if m != nil {
		return m.MoveDate
	}
	return 0
}

func (m *RecordMove) GetRecord() *proto1.Record {
	if m != nil {
		return m.Record
	}
	return nil
}

func (m *RecordMove) GetBeforeContext() *Context {
	if m != nil {
		return m.BeforeContext
	}
	return nil
}

func (m *RecordMove) GetAfterContext() *Context {
	if m != nil {
		return m.AfterContext
	}
	return nil
}

func (m *RecordMove) GetLastUpdate() int64 {
	if m != nil {
		return m.LastUpdate
	}
	return 0
}

type Moves struct {
	Moves                []*RecordMove `protobuf:"bytes,1,rep,name=moves,proto3" json:"moves,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Moves) Reset()         { *m = Moves{} }
func (m *Moves) String() string { return proto.CompactTextString(m) }
func (*Moves) ProtoMessage()    {}
func (*Moves) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{2}
}

func (m *Moves) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Moves.Unmarshal(m, b)
}
func (m *Moves) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Moves.Marshal(b, m, deterministic)
}
func (m *Moves) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Moves.Merge(m, src)
}
func (m *Moves) XXX_Size() int {
	return xxx_messageInfo_Moves.Size(m)
}
func (m *Moves) XXX_DiscardUnknown() {
	xxx_messageInfo_Moves.DiscardUnknown(m)
}

var xxx_messageInfo_Moves proto.InternalMessageInfo

func (m *Moves) GetMoves() []*RecordMove {
	if m != nil {
		return m.Moves
	}
	return nil
}

type MoveRequest struct {
	Move                 *RecordMove `protobuf:"bytes,1,opt,name=move,proto3" json:"move,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *MoveRequest) Reset()         { *m = MoveRequest{} }
func (m *MoveRequest) String() string { return proto.CompactTextString(m) }
func (*MoveRequest) ProtoMessage()    {}
func (*MoveRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{3}
}

func (m *MoveRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MoveRequest.Unmarshal(m, b)
}
func (m *MoveRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MoveRequest.Marshal(b, m, deterministic)
}
func (m *MoveRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MoveRequest.Merge(m, src)
}
func (m *MoveRequest) XXX_Size() int {
	return xxx_messageInfo_MoveRequest.Size(m)
}
func (m *MoveRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MoveRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MoveRequest proto.InternalMessageInfo

func (m *MoveRequest) GetMove() *RecordMove {
	if m != nil {
		return m.Move
	}
	return nil
}

type MoveResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MoveResponse) Reset()         { *m = MoveResponse{} }
func (m *MoveResponse) String() string { return proto.CompactTextString(m) }
func (*MoveResponse) ProtoMessage()    {}
func (*MoveResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{4}
}

func (m *MoveResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MoveResponse.Unmarshal(m, b)
}
func (m *MoveResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MoveResponse.Marshal(b, m, deterministic)
}
func (m *MoveResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MoveResponse.Merge(m, src)
}
func (m *MoveResponse) XXX_Size() int {
	return xxx_messageInfo_MoveResponse.Size(m)
}
func (m *MoveResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MoveResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MoveResponse proto.InternalMessageInfo

type ListRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListRequest) Reset()         { *m = ListRequest{} }
func (m *ListRequest) String() string { return proto.CompactTextString(m) }
func (*ListRequest) ProtoMessage()    {}
func (*ListRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{5}
}

func (m *ListRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListRequest.Unmarshal(m, b)
}
func (m *ListRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListRequest.Marshal(b, m, deterministic)
}
func (m *ListRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListRequest.Merge(m, src)
}
func (m *ListRequest) XXX_Size() int {
	return xxx_messageInfo_ListRequest.Size(m)
}
func (m *ListRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListRequest proto.InternalMessageInfo

type ListResponse struct {
	Moves                []*RecordMove `protobuf:"bytes,1,rep,name=moves,proto3" json:"moves,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *ListResponse) Reset()         { *m = ListResponse{} }
func (m *ListResponse) String() string { return proto.CompactTextString(m) }
func (*ListResponse) ProtoMessage()    {}
func (*ListResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{6}
}

func (m *ListResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListResponse.Unmarshal(m, b)
}
func (m *ListResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListResponse.Marshal(b, m, deterministic)
}
func (m *ListResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListResponse.Merge(m, src)
}
func (m *ListResponse) XXX_Size() int {
	return xxx_messageInfo_ListResponse.Size(m)
}
func (m *ListResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListResponse proto.InternalMessageInfo

func (m *ListResponse) GetMoves() []*RecordMove {
	if m != nil {
		return m.Moves
	}
	return nil
}

type ClearRequest struct {
	InstanceId           int32    `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClearRequest) Reset()         { *m = ClearRequest{} }
func (m *ClearRequest) String() string { return proto.CompactTextString(m) }
func (*ClearRequest) ProtoMessage()    {}
func (*ClearRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{7}
}

func (m *ClearRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClearRequest.Unmarshal(m, b)
}
func (m *ClearRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClearRequest.Marshal(b, m, deterministic)
}
func (m *ClearRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClearRequest.Merge(m, src)
}
func (m *ClearRequest) XXX_Size() int {
	return xxx_messageInfo_ClearRequest.Size(m)
}
func (m *ClearRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ClearRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ClearRequest proto.InternalMessageInfo

func (m *ClearRequest) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

type ClearResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClearResponse) Reset()         { *m = ClearResponse{} }
func (m *ClearResponse) String() string { return proto.CompactTextString(m) }
func (*ClearResponse) ProtoMessage()    {}
func (*ClearResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8a16ecfaf2b6a48f, []int{8}
}

func (m *ClearResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClearResponse.Unmarshal(m, b)
}
func (m *ClearResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClearResponse.Marshal(b, m, deterministic)
}
func (m *ClearResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClearResponse.Merge(m, src)
}
func (m *ClearResponse) XXX_Size() int {
	return xxx_messageInfo_ClearResponse.Size(m)
}
func (m *ClearResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ClearResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ClearResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Context)(nil), "recordprocessor.Context")
	proto.RegisterType((*RecordMove)(nil), "recordprocessor.RecordMove")
	proto.RegisterType((*Moves)(nil), "recordprocessor.Moves")
	proto.RegisterType((*MoveRequest)(nil), "recordprocessor.MoveRequest")
	proto.RegisterType((*MoveResponse)(nil), "recordprocessor.MoveResponse")
	proto.RegisterType((*ListRequest)(nil), "recordprocessor.ListRequest")
	proto.RegisterType((*ListResponse)(nil), "recordprocessor.ListResponse")
	proto.RegisterType((*ClearRequest)(nil), "recordprocessor.ClearRequest")
	proto.RegisterType((*ClearResponse)(nil), "recordprocessor.ClearResponse")
}

func init() { proto.RegisterFile("recordmover.proto", fileDescriptor_8a16ecfaf2b6a48f) }

var fileDescriptor_8a16ecfaf2b6a48f = []byte{
	// 482 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xad, 0x9b, 0x3a, 0x4d, 0xc6, 0x49, 0x2b, 0xf6, 0x64, 0xa5, 0x40, 0x23, 0x9f, 0x72, 0xb2,
	0x21, 0xdc, 0x90, 0x00, 0xa1, 0x02, 0x12, 0x50, 0x2e, 0x46, 0x9c, 0x2d, 0x67, 0x3d, 0x69, 0x2d,
	0x39, 0x9e, 0xb0, 0xbb, 0x89, 0xf8, 0x03, 0x3e, 0x92, 0xdf, 0xe0, 0x03, 0xd0, 0xee, 0xd8, 0x69,
	0x2a, 0x53, 0x22, 0x71, 0xf4, 0x7b, 0x6f, 0xde, 0xbe, 0x79, 0xbb, 0x32, 0x3c, 0x52, 0x28, 0x49,
	0x15, 0x2b, 0xda, 0xa2, 0x8a, 0xd7, 0x8a, 0x0c, 0x89, 0x73, 0x86, 0xd6, 0x8a, 0x24, 0x6a, 0x4d,
	0x6a, 0xf2, 0xfe, 0xa6, 0x34, 0xb7, 0x9b, 0x45, 0x2c, 0x69, 0x95, 0x2c, 0x14, 0x99, 0x5b, 0x54,
	0x15, 0xdd, 0x94, 0x32, 0x61, 0xa1, 0xa4, 0xaa, 0x42, 0x69, 0x4a, 0xaa, 0x13, 0x67, 0xd0, 0x81,
	0xd9, 0x37, 0xfa, 0xe9, 0xc1, 0xe9, 0x15, 0xd5, 0x06, 0x7f, 0x18, 0xf1, 0x0c, 0xfa, 0x0b, 0x5c,
	0x92, 0xc2, 0xd0, 0x9b, 0x7a, 0xb3, 0x60, 0x1e, 0xc6, 0x9d, 0xa1, 0xd4, 0x01, 0x69, 0xa3, 0x13,
	0x13, 0x18, 0x54, 0x24, 0x73, 0x4b, 0x85, 0xc7, 0x53, 0x6f, 0x36, 0x4c, 0x77, 0xdf, 0x22, 0x06,
	0x3f, 0x5f, 0x1a, 0x54, 0x61, 0xef, 0x80, 0x19, 0xcb, 0xa2, 0x5f, 0xc7, 0x00, 0x8c, 0x7c, 0xa1,
	0x2d, 0x8a, 0x4b, 0x08, 0xca, 0x5a, 0x9b, 0xbc, 0x96, 0x98, 0x95, 0x85, 0x4b, 0xe4, 0xa7, 0xd0,
	0x42, 0x1f, 0x0b, 0x2b, 0x58, 0x2a, 0x5a, 0x65, 0x4b, 0xaa, 0x0a, 0x54, 0xee, 0x78, 0x3f, 0x05,
	0x0b, 0x7d, 0x70, 0x88, 0xb8, 0x80, 0xa1, 0xa1, 0x96, 0xee, 0x39, 0x7a, 0x60, 0xe8, 0x8e, 0xb4,
	0xf5, 0x66, 0x45, 0x6e, 0x30, 0x3c, 0x99, 0x7a, 0xb3, 0x5e, 0x3a, 0xb0, 0xc0, 0xbb, 0xdc, 0xa0,
	0x2d, 0x82, 0xc3, 0x86, 0xfe, 0xa1, 0x22, 0x98, 0x10, 0x6f, 0xe0, 0x8c, 0x2b, 0xc9, 0x24, 0x97,
	0x19, 0xf6, 0xef, 0x4d, 0xee, 0xee, 0x2d, 0x6e, 0xca, 0x4e, 0xc7, 0xac, 0x6f, 0xbb, 0x7f, 0x05,
	0x63, 0x57, 0xc3, 0x6e, 0xfe, 0xf4, 0xc0, 0xfc, 0xc8, 0xc9, 0xdb, 0xf1, 0x4b, 0x08, 0xaa, 0x5c,
	0x9b, 0x6c, 0xb3, 0x76, 0x0b, 0x0d, 0xdc, 0x42, 0x60, 0xa1, 0x6f, 0x0e, 0x89, 0x5e, 0x82, 0x6f,
	0x6b, 0xd5, 0xe2, 0x39, 0xf8, 0x76, 0x4f, 0x1d, 0x7a, 0xd3, 0xde, 0x2c, 0x98, 0x5f, 0x74, 0x0e,
	0xb8, 0xbb, 0x83, 0x94, 0x95, 0xd1, 0x6b, 0x08, 0xdc, 0x27, 0x7e, 0xdf, 0xa0, 0x36, 0x22, 0x81,
	0x13, 0x8b, 0x37, 0x8f, 0xe4, 0x9f, 0x06, 0x4e, 0x18, 0x9d, 0xc1, 0x88, 0xe7, 0xf5, 0x9a, 0x6a,
	0x8d, 0xd1, 0x18, 0x82, 0xeb, 0x52, 0x9b, 0xc6, 0x2f, 0x7a, 0x0b, 0x23, 0xfe, 0x64, 0xfa, 0x7f,
	0x12, 0x26, 0x30, 0xba, 0xaa, 0x30, 0x57, 0x6d, 0xc4, 0x43, 0x8f, 0x27, 0x3a, 0x87, 0x71, 0x33,
	0xc0, 0x87, 0xce, 0x7f, 0x7b, 0xbc, 0xe4, 0x57, 0x54, 0xdb, 0x52, 0xa2, 0xf8, 0x7c, 0xef, 0x31,
	0x3e, 0xee, 0x64, 0xd8, 0x2b, 0x64, 0xf2, 0xe4, 0x01, 0xb6, 0x59, 0xf7, 0x48, 0x7c, 0x82, 0xa1,
	0xdd, 0x90, 0x2f, 0xa0, 0xeb, 0xb5, 0x57, 0xc6, 0x5f, 0xbc, 0xf6, 0xbb, 0x89, 0x8e, 0xc4, 0x35,
	0x0c, 0x5d, 0x72, 0x97, 0xab, 0xab, 0xde, 0xaf, 0x61, 0xf2, 0xf4, 0x21, 0xba, 0x75, 0x5b, 0xf4,
	0xdd, 0x5f, 0xe0, 0xc5, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x95, 0x63, 0x6e, 0x6b, 0x72, 0x04,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MoveServiceClient is the client API for MoveService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MoveServiceClient interface {
	RecordMove(ctx context.Context, in *MoveRequest, opts ...grpc.CallOption) (*MoveResponse, error)
	ListMoves(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	ClearMove(ctx context.Context, in *ClearRequest, opts ...grpc.CallOption) (*ClearResponse, error)
}

type moveServiceClient struct {
	cc *grpc.ClientConn
}

func NewMoveServiceClient(cc *grpc.ClientConn) MoveServiceClient {
	return &moveServiceClient{cc}
}

func (c *moveServiceClient) RecordMove(ctx context.Context, in *MoveRequest, opts ...grpc.CallOption) (*MoveResponse, error) {
	out := new(MoveResponse)
	err := c.cc.Invoke(ctx, "/recordprocessor.MoveService/RecordMove", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moveServiceClient) ListMoves(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/recordprocessor.MoveService/ListMoves", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moveServiceClient) ClearMove(ctx context.Context, in *ClearRequest, opts ...grpc.CallOption) (*ClearResponse, error) {
	out := new(ClearResponse)
	err := c.cc.Invoke(ctx, "/recordprocessor.MoveService/ClearMove", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MoveServiceServer is the server API for MoveService service.
type MoveServiceServer interface {
	RecordMove(context.Context, *MoveRequest) (*MoveResponse, error)
	ListMoves(context.Context, *ListRequest) (*ListResponse, error)
	ClearMove(context.Context, *ClearRequest) (*ClearResponse, error)
}

func RegisterMoveServiceServer(s *grpc.Server, srv MoveServiceServer) {
	s.RegisterService(&_MoveService_serviceDesc, srv)
}

func _MoveService_RecordMove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MoveServiceServer).RecordMove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recordprocessor.MoveService/RecordMove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MoveServiceServer).RecordMove(ctx, req.(*MoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MoveService_ListMoves_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MoveServiceServer).ListMoves(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recordprocessor.MoveService/ListMoves",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MoveServiceServer).ListMoves(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MoveService_ClearMove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClearRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MoveServiceServer).ClearMove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recordprocessor.MoveService/ClearMove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MoveServiceServer).ClearMove(ctx, req.(*ClearRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MoveService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "recordprocessor.MoveService",
	HandlerType: (*MoveServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RecordMove",
			Handler:    _MoveService_RecordMove_Handler,
		},
		{
			MethodName: "ListMoves",
			Handler:    _MoveService_ListMoves_Handler,
		},
		{
			MethodName: "ClearMove",
			Handler:    _MoveService_ClearMove_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "recordmover.proto",
}
