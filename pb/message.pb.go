// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

package pb

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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// status
// 0 - Alive, 1 - Suspect, 2 - Confirm
type MbrStatsMsg_Type int32

const (
	MbrStatsMsg_Alive   MbrStatsMsg_Type = 0
	MbrStatsMsg_Suspect MbrStatsMsg_Type = 1
	MbrStatsMsg_Confirm MbrStatsMsg_Type = 2
)

var MbrStatsMsg_Type_name = map[int32]string{
	0: "Alive",
	1: "Suspect",
	2: "Confirm",
}

var MbrStatsMsg_Type_value = map[string]int32{
	"Alive":   0,
	"Suspect": 1,
	"Confirm": 2,
}

func (x MbrStatsMsg_Type) String() string {
	return proto.EnumName(MbrStatsMsg_Type_name, int32(x))
}

func (MbrStatsMsg_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{7, 0}
}

type Message struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// address of source member who sent this message
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	// Types that are valid to be assigned to Payload:
	//	*Message_Ping
	//	*Message_Ack
	//	*Message_Nack
	//	*Message_IndirectPing
	//	*Message_Membership
	Payload              isMessage_Payload `protobuf_oneof:"payload"`
	PiggyBack            *PiggyBack        `protobuf:"bytes,8,opt,name=piggyBack,proto3" json:"piggyBack,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Message) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type isMessage_Payload interface {
	isMessage_Payload()
}

type Message_Ping struct {
	Ping *Ping `protobuf:"bytes,3,opt,name=ping,proto3,oneof"`
}

type Message_Ack struct {
	Ack *Ack `protobuf:"bytes,4,opt,name=ack,proto3,oneof"`
}

type Message_Nack struct {
	Nack *Nack `protobuf:"bytes,5,opt,name=nack,proto3,oneof"`
}

type Message_IndirectPing struct {
	IndirectPing *IndirectPing `protobuf:"bytes,6,opt,name=indirect_ping,json=indirectPing,proto3,oneof"`
}

type Message_Membership struct {
	Membership *Membership `protobuf:"bytes,7,opt,name=membership,proto3,oneof"`
}

func (*Message_Ping) isMessage_Payload() {}

func (*Message_Ack) isMessage_Payload() {}

func (*Message_Nack) isMessage_Payload() {}

func (*Message_IndirectPing) isMessage_Payload() {}

func (*Message_Membership) isMessage_Payload() {}

func (m *Message) GetPayload() isMessage_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *Message) GetPing() *Ping {
	if x, ok := m.GetPayload().(*Message_Ping); ok {
		return x.Ping
	}
	return nil
}

func (m *Message) GetAck() *Ack {
	if x, ok := m.GetPayload().(*Message_Ack); ok {
		return x.Ack
	}
	return nil
}

func (m *Message) GetNack() *Nack {
	if x, ok := m.GetPayload().(*Message_Nack); ok {
		return x.Nack
	}
	return nil
}

func (m *Message) GetIndirectPing() *IndirectPing {
	if x, ok := m.GetPayload().(*Message_IndirectPing); ok {
		return x.IndirectPing
	}
	return nil
}

func (m *Message) GetMembership() *Membership {
	if x, ok := m.GetPayload().(*Message_Membership); ok {
		return x.Membership
	}
	return nil
}

func (m *Message) GetPiggyBack() *PiggyBack {
	if m != nil {
		return m.PiggyBack
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Message) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Message_OneofMarshaler, _Message_OneofUnmarshaler, _Message_OneofSizer, []interface{}{
		(*Message_Ping)(nil),
		(*Message_Ack)(nil),
		(*Message_Nack)(nil),
		(*Message_IndirectPing)(nil),
		(*Message_Membership)(nil),
	}
}

func _Message_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Message)
	// payload
	switch x := m.Payload.(type) {
	case *Message_Ping:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Ping); err != nil {
			return err
		}
	case *Message_Ack:
		b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Ack); err != nil {
			return err
		}
	case *Message_Nack:
		b.EncodeVarint(5<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Nack); err != nil {
			return err
		}
	case *Message_IndirectPing:
		b.EncodeVarint(6<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.IndirectPing); err != nil {
			return err
		}
	case *Message_Membership:
		b.EncodeVarint(7<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Membership); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Message.Payload has unexpected type %T", x)
	}
	return nil
}

func _Message_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Message)
	switch tag {
	case 3: // payload.ping
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Ping)
		err := b.DecodeMessage(msg)
		m.Payload = &Message_Ping{msg}
		return true, err
	case 4: // payload.ack
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Ack)
		err := b.DecodeMessage(msg)
		m.Payload = &Message_Ack{msg}
		return true, err
	case 5: // payload.nack
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Nack)
		err := b.DecodeMessage(msg)
		m.Payload = &Message_Nack{msg}
		return true, err
	case 6: // payload.indirect_ping
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(IndirectPing)
		err := b.DecodeMessage(msg)
		m.Payload = &Message_IndirectPing{msg}
		return true, err
	case 7: // payload.membership
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Membership)
		err := b.DecodeMessage(msg)
		m.Payload = &Message_Membership{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Message_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Message)
	// payload
	switch x := m.Payload.(type) {
	case *Message_Ping:
		s := proto.Size(x.Ping)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Message_Ack:
		s := proto.Size(x.Ack)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Message_Nack:
		s := proto.Size(x.Nack)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Message_IndirectPing:
		s := proto.Size(x.IndirectPing)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Message_Membership:
		s := proto.Size(x.Membership)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type Ping struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ping) Reset()         { *m = Ping{} }
func (m *Ping) String() string { return proto.CompactTextString(m) }
func (*Ping) ProtoMessage()    {}
func (*Ping) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{1}
}

func (m *Ping) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ping.Unmarshal(m, b)
}
func (m *Ping) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ping.Marshal(b, m, deterministic)
}
func (m *Ping) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ping.Merge(m, src)
}
func (m *Ping) XXX_Size() int {
	return xxx_messageInfo_Ping.Size(m)
}
func (m *Ping) XXX_DiscardUnknown() {
	xxx_messageInfo_Ping.DiscardUnknown(m)
}

var xxx_messageInfo_Ping proto.InternalMessageInfo

type Ack struct {
	Payload              string   `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ack) Reset()         { *m = Ack{} }
func (m *Ack) String() string { return proto.CompactTextString(m) }
func (*Ack) ProtoMessage()    {}
func (*Ack) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{2}
}

func (m *Ack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ack.Unmarshal(m, b)
}
func (m *Ack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ack.Marshal(b, m, deterministic)
}
func (m *Ack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ack.Merge(m, src)
}
func (m *Ack) XXX_Size() int {
	return xxx_messageInfo_Ack.Size(m)
}
func (m *Ack) XXX_DiscardUnknown() {
	xxx_messageInfo_Ack.DiscardUnknown(m)
}

var xxx_messageInfo_Ack proto.InternalMessageInfo

func (m *Ack) GetPayload() string {
	if m != nil {
		return m.Payload
	}
	return ""
}

type Nack struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nack) Reset()         { *m = Nack{} }
func (m *Nack) String() string { return proto.CompactTextString(m) }
func (*Nack) ProtoMessage()    {}
func (*Nack) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{3}
}

func (m *Nack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nack.Unmarshal(m, b)
}
func (m *Nack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nack.Marshal(b, m, deterministic)
}
func (m *Nack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nack.Merge(m, src)
}
func (m *Nack) XXX_Size() int {
	return xxx_messageInfo_Nack.Size(m)
}
func (m *Nack) XXX_DiscardUnknown() {
	xxx_messageInfo_Nack.DiscardUnknown(m)
}

var xxx_messageInfo_Nack proto.InternalMessageInfo

type IndirectPing struct {
	// indirect-ping's target address
	Target               string   `protobuf:"bytes,2,opt,name=target,proto3" json:"target,omitempty"`
	Nack                 bool     `protobuf:"varint,3,opt,name=nack,proto3" json:"nack,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IndirectPing) Reset()         { *m = IndirectPing{} }
func (m *IndirectPing) String() string { return proto.CompactTextString(m) }
func (*IndirectPing) ProtoMessage()    {}
func (*IndirectPing) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{4}
}

func (m *IndirectPing) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IndirectPing.Unmarshal(m, b)
}
func (m *IndirectPing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IndirectPing.Marshal(b, m, deterministic)
}
func (m *IndirectPing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndirectPing.Merge(m, src)
}
func (m *IndirectPing) XXX_Size() int {
	return xxx_messageInfo_IndirectPing.Size(m)
}
func (m *IndirectPing) XXX_DiscardUnknown() {
	xxx_messageInfo_IndirectPing.DiscardUnknown(m)
}

var xxx_messageInfo_IndirectPing proto.InternalMessageInfo

func (m *IndirectPing) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *IndirectPing) GetNack() bool {
	if m != nil {
		return m.Nack
	}
	return false
}

type PiggyBack struct {
	MbrStatsMsg          *MbrStatsMsg `protobuf:"bytes,1,opt,name=mbrStatsMsg,proto3" json:"mbrStatsMsg,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *PiggyBack) Reset()         { *m = PiggyBack{} }
func (m *PiggyBack) String() string { return proto.CompactTextString(m) }
func (*PiggyBack) ProtoMessage()    {}
func (*PiggyBack) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{5}
}

func (m *PiggyBack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PiggyBack.Unmarshal(m, b)
}
func (m *PiggyBack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PiggyBack.Marshal(b, m, deterministic)
}
func (m *PiggyBack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PiggyBack.Merge(m, src)
}
func (m *PiggyBack) XXX_Size() int {
	return xxx_messageInfo_PiggyBack.Size(m)
}
func (m *PiggyBack) XXX_DiscardUnknown() {
	xxx_messageInfo_PiggyBack.DiscardUnknown(m)
}

var xxx_messageInfo_PiggyBack proto.InternalMessageInfo

func (m *PiggyBack) GetMbrStatsMsg() *MbrStatsMsg {
	if m != nil {
		return m.MbrStatsMsg
	}
	return nil
}

// Membership has a list of Member status messages
type Membership struct {
	MbrStatsMsgs         []*MbrStatsMsg `protobuf:"bytes,1,rep,name=mbrStatsMsgs,proto3" json:"mbrStatsMsgs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Membership) Reset()         { *m = Membership{} }
func (m *Membership) String() string { return proto.CompactTextString(m) }
func (*Membership) ProtoMessage()    {}
func (*Membership) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{6}
}

func (m *Membership) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Membership.Unmarshal(m, b)
}
func (m *Membership) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Membership.Marshal(b, m, deterministic)
}
func (m *Membership) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Membership.Merge(m, src)
}
func (m *Membership) XXX_Size() int {
	return xxx_messageInfo_Membership.Size(m)
}
func (m *Membership) XXX_DiscardUnknown() {
	xxx_messageInfo_Membership.DiscardUnknown(m)
}

var xxx_messageInfo_Membership proto.InternalMessageInfo

func (m *Membership) GetMbrStatsMsgs() []*MbrStatsMsg {
	if m != nil {
		return m.MbrStatsMsgs
	}
	return nil
}

// Member status message
type MbrStatsMsg struct {
	Type                 MbrStatsMsg_Type `protobuf:"varint,1,opt,name=type,proto3,enum=pb.MbrStatsMsg_Type" json:"type,omitempty"`
	Id                   string           `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Incarnation          uint32           `protobuf:"varint,3,opt,name=incarnation,proto3" json:"incarnation,omitempty"`
	Address              string           `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *MbrStatsMsg) Reset()         { *m = MbrStatsMsg{} }
func (m *MbrStatsMsg) String() string { return proto.CompactTextString(m) }
func (*MbrStatsMsg) ProtoMessage()    {}
func (*MbrStatsMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{7}
}

func (m *MbrStatsMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MbrStatsMsg.Unmarshal(m, b)
}
func (m *MbrStatsMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MbrStatsMsg.Marshal(b, m, deterministic)
}
func (m *MbrStatsMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MbrStatsMsg.Merge(m, src)
}
func (m *MbrStatsMsg) XXX_Size() int {
	return xxx_messageInfo_MbrStatsMsg.Size(m)
}
func (m *MbrStatsMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_MbrStatsMsg.DiscardUnknown(m)
}

var xxx_messageInfo_MbrStatsMsg proto.InternalMessageInfo

func (m *MbrStatsMsg) GetType() MbrStatsMsg_Type {
	if m != nil {
		return m.Type
	}
	return MbrStatsMsg_Alive
}

func (m *MbrStatsMsg) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *MbrStatsMsg) GetIncarnation() uint32 {
	if m != nil {
		return m.Incarnation
	}
	return 0
}

func (m *MbrStatsMsg) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func init() {
	proto.RegisterEnum("pb.MbrStatsMsg_Type", MbrStatsMsg_Type_name, MbrStatsMsg_Type_value)
	proto.RegisterType((*Message)(nil), "pb.Message")
	proto.RegisterType((*Ping)(nil), "pb.Ping")
	proto.RegisterType((*Ack)(nil), "pb.Ack")
	proto.RegisterType((*Nack)(nil), "pb.Nack")
	proto.RegisterType((*IndirectPing)(nil), "pb.IndirectPing")
	proto.RegisterType((*PiggyBack)(nil), "pb.PiggyBack")
	proto.RegisterType((*Membership)(nil), "pb.Membership")
	proto.RegisterType((*MbrStatsMsg)(nil), "pb.MbrStatsMsg")
}

func init() { proto.RegisterFile("message.proto", fileDescriptor_33c57e4bae7b9afd) }

var fileDescriptor_33c57e4bae7b9afd = []byte{
	// 420 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x52, 0x5f, 0xab, 0xd3, 0x30,
	0x14, 0x5f, 0xbb, 0xde, 0x75, 0x3d, 0xdd, 0xe6, 0x38, 0x88, 0x14, 0x04, 0x1d, 0x7d, 0x1a, 0x5c,
	0x18, 0xba, 0xfb, 0x20, 0xf8, 0x20, 0xec, 0xfa, 0x32, 0x1f, 0x26, 0x92, 0xeb, 0xbb, 0xa4, 0x6d,
	0xac, 0x61, 0xb7, 0x69, 0x48, 0xa2, 0xb0, 0x6f, 0x25, 0xf8, 0x05, 0x25, 0x59, 0xee, 0x9a, 0x89,
	0x8f, 0xe7, 0xf7, 0x27, 0xe7, 0x9c, 0x5f, 0x0e, 0xcc, 0x3b, 0xa6, 0x35, 0x6d, 0xd9, 0x46, 0xaa,
	0xde, 0xf4, 0x18, 0xcb, 0xaa, 0xfc, 0x13, 0x43, 0x7a, 0x38, 0xa3, 0xb8, 0x80, 0x98, 0x37, 0x45,
	0xb4, 0x8a, 0xd6, 0x19, 0x89, 0x79, 0x83, 0x05, 0xa4, 0xb4, 0x69, 0x14, 0xd3, 0xba, 0x88, 0x1d,
	0xf8, 0x54, 0xe2, 0x2b, 0x48, 0x24, 0x17, 0x6d, 0x31, 0x5e, 0x45, 0xeb, 0x7c, 0x3b, 0xdd, 0xc8,
	0x6a, 0xf3, 0x85, 0x8b, 0x76, 0x3f, 0x22, 0x0e, 0xc7, 0x97, 0x30, 0xa6, 0xf5, 0xb1, 0x48, 0x1c,
	0x9d, 0x5a, 0x7a, 0x57, 0x1f, 0xf7, 0x23, 0x62, 0x51, 0x6b, 0x16, 0x96, 0xbd, 0x19, 0xcc, 0x9f,
	0xa9, 0xa3, 0x1d, 0x8e, 0xef, 0x60, 0xce, 0x45, 0xc3, 0x15, 0xab, 0xcd, 0x37, 0xd7, 0x65, 0xe2,
	0x84, 0x4b, 0x2b, 0xfc, 0xe4, 0x09, 0xdf, 0x6d, 0xc6, 0x83, 0x1a, 0xdf, 0x00, 0x74, 0xac, 0xab,
	0x98, 0xd2, 0x3f, 0xb8, 0x2c, 0x52, 0xe7, 0x5a, 0x58, 0xd7, 0xe1, 0x82, 0xee, 0x47, 0x24, 0xd0,
	0xe0, 0x2d, 0x64, 0x92, 0xb7, 0xed, 0xe9, 0xde, 0xce, 0x33, 0x75, 0x86, 0xf9, 0x79, 0x19, 0x0f,
	0x92, 0x81, 0xbf, 0xcf, 0x20, 0x95, 0xf4, 0xf4, 0xd8, 0xd3, 0xa6, 0x9c, 0x40, 0x62, 0x3b, 0x96,
	0xaf, 0x61, 0xbc, 0xab, 0x8f, 0x36, 0x28, 0xcf, 0x3c, 0x05, 0x15, 0x08, 0xed, 0x6e, 0xe5, 0x7b,
	0x98, 0x85, 0xa3, 0xe3, 0x0b, 0x98, 0x18, 0xaa, 0x5a, 0x66, 0xbc, 0xc1, 0x57, 0x88, 0x3e, 0x1b,
	0x1b, 0xec, 0xf4, 0x9c, 0x47, 0xf9, 0x01, 0xb2, 0xcb, 0x3c, 0xf8, 0x16, 0xf2, 0xae, 0x52, 0x0f,
	0x86, 0x1a, 0x7d, 0xd0, 0xad, 0xfb, 0xac, 0x7c, 0xfb, 0xcc, 0x2d, 0x39, 0xc0, 0x24, 0xd4, 0x94,
	0x3b, 0x80, 0x21, 0x00, 0xbc, 0x83, 0x59, 0x40, 0xea, 0x22, 0x5a, 0x8d, 0xff, 0xf7, 0xc2, 0x95,
	0xa8, 0xfc, 0x1d, 0x41, 0x1e, 0xb0, 0xb8, 0x86, 0xc4, 0x9c, 0x24, 0x73, 0xed, 0x17, 0xdb, 0xe7,
	0xff, 0x98, 0x37, 0x5f, 0x4f, 0x92, 0x11, 0xa7, 0xf0, 0x37, 0x15, 0x5f, 0x6e, 0x6a, 0x05, 0x39,
	0x17, 0x35, 0x55, 0x82, 0x1a, 0xde, 0x0b, 0xb7, 0xe7, 0x9c, 0x84, 0x50, 0x78, 0x75, 0xc9, 0xd5,
	0xd5, 0x95, 0xb7, 0x90, 0xd8, 0x97, 0x31, 0x83, 0x9b, 0xdd, 0x23, 0xff, 0xc5, 0x96, 0x23, 0xcc,
	0x21, 0x7d, 0xf8, 0xa9, 0x25, 0xab, 0xcd, 0x32, 0xb2, 0xc5, 0xc7, 0x5e, 0x7c, 0xe7, 0xaa, 0x5b,
	0xc6, 0xd5, 0xc4, 0xdd, 0xf8, 0xdd, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x1b, 0x02, 0xe4, 0xb3,
	0xf4, 0x02, 0x00, 0x00,
}
