// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.27.3
// source: internal/actor.proto

package msg

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address     string `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`         // 服务地址
	Name        string `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`               // 服务名称
	ServiceType string `protobuf:"bytes,3,opt,name=ServiceType,proto3" json:"ServiceType,omitempty"` // 服务类型
	ServiceUid  string `protobuf:"bytes,4,opt,name=ServiceUid,proto3" json:"ServiceUid,omitempty"`   // 服务实例在集群的唯一标识(serverId:serviceName@Uid)
	State       int32  `protobuf:"varint,5,opt,name=State,proto3" json:"State,omitempty"`            // 服务状态(0: 正常, 1: 退休)
	ServerId    int32  `protobuf:"varint,6,opt,name=ServerId,proto3" json:"ServerId,omitempty"`      // 服务ID
	Version     int64  `protobuf:"varint,7,opt,name=Version,proto3" json:"Version,omitempty"`        // 服务版本号
	RpcType     string `protobuf:"bytes,8,opt,name=RpcType,proto3" json:"RpcType,omitempty"`         // rpc类型(默认使用rpcx)
	NodeUid     string `protobuf:"bytes,9,opt,name=NodeUid,proto3" json:"NodeUid,omitempty"`         // 节点唯一标识(这个主要是用来区分本地服务用的)
}

func (x *PID) Reset() {
	*x = PID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_actor_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PID) ProtoMessage() {}

func (x *PID) ProtoReflect() protoreflect.Message {
	mi := &file_internal_actor_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PID.ProtoReflect.Descriptor instead.
func (*PID) Descriptor() ([]byte, []int) {
	return file_internal_actor_proto_rawDescGZIP(), []int{0}
}

func (x *PID) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *PID) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PID) GetServiceType() string {
	if x != nil {
		return x.ServiceType
	}
	return ""
}

func (x *PID) GetServiceUid() string {
	if x != nil {
		return x.ServiceUid
	}
	return ""
}

func (x *PID) GetState() int32 {
	if x != nil {
		return x.State
	}
	return 0
}

func (x *PID) GetServerId() int32 {
	if x != nil {
		return x.ServerId
	}
	return 0
}

func (x *PID) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *PID) GetRpcType() string {
	if x != nil {
		return x.RpcType
	}
	return ""
}

func (x *PID) GetNodeUid() string {
	if x != nil {
		return x.NodeUid
	}
	return ""
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TypeId        int32             `protobuf:"varint,1,opt,name=TypeId,proto3" json:"TypeId,omitempty"`                                                                                                      // 消息类型ID
	TypeName      string            `protobuf:"bytes,2,opt,name=TypeName,proto3" json:"TypeName,omitempty"`                                                                                                   // 消息类型名称
	Sender        *PID              `protobuf:"bytes,3,opt,name=Sender,proto3" json:"Sender,omitempty"`                                                                                                       // 发送者
	Receiver      *PID              `protobuf:"bytes,4,opt,name=Receiver,proto3" json:"Receiver,omitempty"`                                                                                                   // 接收者
	Method        string            `protobuf:"bytes,5,opt,name=Method,proto3" json:"Method,omitempty"`                                                                                                       // 调用方法
	Request       []byte            `protobuf:"bytes,6,opt,name=Request,proto3" json:"Request,omitempty"`                                                                                                     // 方法参数
	Response      []byte            `protobuf:"bytes,7,opt,name=Response,proto3" json:"Response,omitempty"`                                                                                                   // 方法返回值
	Err           string            `protobuf:"bytes,8,opt,name=Err,proto3" json:"Err,omitempty"`                                                                                                             // 错误信息
	MessageHeader map[string]string `protobuf:"bytes,9,rep,name=MessageHeader,proto3" json:"MessageHeader,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // 消息头(额外信息)
	Reply         bool              `protobuf:"varint,10,opt,name=Reply,proto3" json:"Reply,omitempty"`                                                                                                       // 是否是回复
	ReqId         uint64            `protobuf:"varint,11,opt,name=ReqId,proto3" json:"ReqId,omitempty"`                                                                                                       // 请求ID
	NeedResp      bool              `protobuf:"varint,12,opt,name=NeedResp,proto3" json:"NeedResp,omitempty"`                                                                                                 // 是否需要回复
	CompressType  int32             `protobuf:"varint,13,opt,name=CompressType,proto3" json:"CompressType,omitempty"`                                                                                         // 压缩类型(0无压缩)
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_actor_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_internal_actor_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_internal_actor_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetTypeId() int32 {
	if x != nil {
		return x.TypeId
	}
	return 0
}

func (x *Message) GetTypeName() string {
	if x != nil {
		return x.TypeName
	}
	return ""
}

func (x *Message) GetSender() *PID {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *Message) GetReceiver() *PID {
	if x != nil {
		return x.Receiver
	}
	return nil
}

func (x *Message) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Message) GetRequest() []byte {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *Message) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *Message) GetErr() string {
	if x != nil {
		return x.Err
	}
	return ""
}

func (x *Message) GetMessageHeader() map[string]string {
	if x != nil {
		return x.MessageHeader
	}
	return nil
}

func (x *Message) GetReply() bool {
	if x != nil {
		return x.Reply
	}
	return false
}

func (x *Message) GetReqId() uint64 {
	if x != nil {
		return x.ReqId
	}
	return 0
}

func (x *Message) GetNeedResp() bool {
	if x != nil {
		return x.NeedResp
	}
	return false
}

func (x *Message) GetCompressType() int32 {
	if x != nil {
		return x.CompressType
	}
	return 0
}

var File_internal_actor_proto protoreflect.FileDescriptor

var file_internal_actor_proto_rawDesc = []byte{
	0x0a, 0x14, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x22, 0xf5, 0x01,
	0x0a, 0x03, 0x50, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x55, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x55, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x18, 0x0a, 0x07, 0x52, 0x70, 0x63, 0x54, 0x79, 0x70, 0x65, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x52, 0x70, 0x63, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x4e,
	0x6f, 0x64, 0x65, 0x55, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4e, 0x6f,
	0x64, 0x65, 0x55, 0x69, 0x64, 0x22, 0xe0, 0x03, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x54, 0x79, 0x70, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x06, 0x54, 0x79, 0x70, 0x65, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x54, 0x79, 0x70,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x54, 0x79, 0x70,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x06, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x50, 0x49,
	0x44, 0x52, 0x06, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x26, 0x0a, 0x08, 0x52, 0x65, 0x63,
	0x65, 0x69, 0x76, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x50, 0x49, 0x44, 0x52, 0x08, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
	0x72, 0x12, 0x16, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x10, 0x0a, 0x03, 0x45, 0x72, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x45, 0x72,
	0x72, 0x12, 0x47, 0x0a, 0x0d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0d, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x52, 0x65, 0x71, 0x49, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x05, 0x52, 0x65, 0x71, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x65, 0x65, 0x64, 0x52, 0x65,
	0x73, 0x70, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x4e, 0x65, 0x65, 0x64, 0x52, 0x65,
	0x73, 0x70, 0x12, 0x22, 0x0a, 0x0c, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65,
	0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0x1a, 0x40, 0x0a, 0x12, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x3b, 0x6d, 0x73,
	0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_actor_proto_rawDescOnce sync.Once
	file_internal_actor_proto_rawDescData = file_internal_actor_proto_rawDesc
)

func file_internal_actor_proto_rawDescGZIP() []byte {
	file_internal_actor_proto_rawDescOnce.Do(func() {
		file_internal_actor_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_actor_proto_rawDescData)
	})
	return file_internal_actor_proto_rawDescData
}

var file_internal_actor_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_internal_actor_proto_goTypes = []interface{}{
	(*PID)(nil),     // 0: actor.PID
	(*Message)(nil), // 1: actor.Message
	nil,             // 2: actor.Message.MessageHeaderEntry
}
var file_internal_actor_proto_depIdxs = []int32{
	0, // 0: actor.Message.Sender:type_name -> actor.PID
	0, // 1: actor.Message.Receiver:type_name -> actor.PID
	2, // 2: actor.Message.MessageHeader:type_name -> actor.Message.MessageHeaderEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_internal_actor_proto_init() }
func file_internal_actor_proto_init() {
	if File_internal_actor_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_actor_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PID); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_actor_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_actor_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_actor_proto_goTypes,
		DependencyIndexes: file_internal_actor_proto_depIdxs,
		MessageInfos:      file_internal_actor_proto_msgTypes,
	}.Build()
	File_internal_actor_proto = out.File
	file_internal_actor_proto_rawDesc = nil
	file_internal_actor_proto_goTypes = nil
	file_internal_actor_proto_depIdxs = nil
}