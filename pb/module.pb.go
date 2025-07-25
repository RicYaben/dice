// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.12.4
// source: proto/module.proto

package pb

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Empty struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_proto_module_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{0}
}

type Fields struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Fields        []byte                 `protobuf:"bytes,1,opt,name=fields,proto3" json:"fields,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Fields) Reset() {
	*x = Fields{}
	mi := &file_proto_module_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Fields) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fields) ProtoMessage() {}

func (x *Fields) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fields.ProtoReflect.Descriptor instead.
func (*Fields) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{1}
}

func (x *Fields) GetFields() []byte {
	if x != nil {
		return x.Fields
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            uint32                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Event) Reset() {
	*x = Event{}
	mi := &file_proto_module_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{2}
}

func (x *Event) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Event) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

type HandleRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	AddServer     uint32                 `protobuf:"varint,1,opt,name=add_server,json=addServer,proto3" json:"add_server,omitempty"`
	Event         *Event                 `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandleRequest) Reset() {
	*x = HandleRequest{}
	mi := &file_proto_module_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandleRequest) ProtoMessage() {}

func (x *HandleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandleRequest.ProtoReflect.Descriptor instead.
func (*HandleRequest) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{3}
}

func (x *HandleRequest) GetAddServer() uint32 {
	if x != nil {
		return x.AddServer
	}
	return 0
}

func (x *HandleRequest) GetEvent() *Event {
	if x != nil {
		return x.Event
	}
	return nil
}

type IDRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            uint32                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *IDRequest) Reset() {
	*x = IDRequest{}
	mi := &file_proto_module_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IDRequest) ProtoMessage() {}

func (x *IDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IDRequest.ProtoReflect.Descriptor instead.
func (*IDRequest) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{4}
}

func (x *IDRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type QueryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Query         string                 `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	mi := &file_proto_module_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{5}
}

func (x *QueryRequest) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

type QueryResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Hosts         []*Host                `protobuf:"bytes,1,rep,name=hosts,proto3" json:"hosts,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QueryResponse) Reset() {
	*x = QueryResponse{}
	mi := &file_proto_module_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResponse) ProtoMessage() {}

func (x *QueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryResponse.ProtoReflect.Descriptor instead.
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{6}
}

func (x *QueryResponse) GetHosts() []*Host {
	if x != nil {
		return x.Hosts
	}
	return nil
}

type LabelHostRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	HostId        uint32                 `protobuf:"varint,1,opt,name=host_id,json=hostId,proto3" json:"host_id,omitempty"`
	Label         string                 `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LabelHostRequest) Reset() {
	*x = LabelHostRequest{}
	mi := &file_proto_module_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LabelHostRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelHostRequest) ProtoMessage() {}

func (x *LabelHostRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_module_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelHostRequest.ProtoReflect.Descriptor instead.
func (*LabelHostRequest) Descriptor() ([]byte, []int) {
	return file_proto_module_proto_rawDescGZIP(), []int{7}
}

func (x *LabelHostRequest) GetHostId() uint32 {
	if x != nil {
		return x.HostId
	}
	return 0
}

func (x *LabelHostRequest) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

var File_proto_module_proto protoreflect.FileDescriptor

const file_proto_module_proto_rawDesc = "" +
	"\n" +
	"\x12proto/module.proto\x12\x05proto\x1a\x12proto/models.proto\"\a\n" +
	"\x05Empty\" \n" +
	"\x06Fields\x12\x16\n" +
	"\x06fields\x18\x01 \x01(\fR\x06fields\"+\n" +
	"\x05Event\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\rR\x02id\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\"R\n" +
	"\rHandleRequest\x12\x1d\n" +
	"\n" +
	"add_server\x18\x01 \x01(\rR\taddServer\x12\"\n" +
	"\x05event\x18\x02 \x01(\v2\f.proto.EventR\x05event\"\x1b\n" +
	"\tIDRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\rR\x02id\"$\n" +
	"\fQueryRequest\x12\x14\n" +
	"\x05query\x18\x01 \x01(\tR\x05query\"2\n" +
	"\rQueryResponse\x12!\n" +
	"\x05hosts\x18\x01 \x03(\v2\v.proto.HostR\x05hosts\"A\n" +
	"\x10LabelHostRequest\x12\x17\n" +
	"\ahost_id\x18\x01 \x01(\rR\x06hostId\x12\x14\n" +
	"\x05label\x18\x02 \x01(\tR\x05label2\x8a\x01\n" +
	"\x06Module\x12'\n" +
	"\tPropagate\x12\f.proto.Empty\x1a\f.proto.Empty\x12)\n" +
	"\n" +
	"Properties\x12\f.proto.Empty\x1a\r.proto.Fields\x12,\n" +
	"\x06Handle\x12\x14.proto.HandleRequest\x1a\f.proto.Empty2\x9f\x03\n" +
	"\aAdapter\x12(\n" +
	"\aGetHost\x12\x10.proto.IDRequest\x1a\v.proto.Host\x12,\n" +
	"\tGetSource\x12\x10.proto.IDRequest\x1a\r.proto.Source\x12(\n" +
	"\aGetScan\x12\x10.proto.IDRequest\x1a\v.proto.Scan\x12&\n" +
	"\bAddLabel\x12\f.proto.Label\x1a\f.proto.Empty\x122\n" +
	"\x0eAddFingerprint\x12\x12.proto.Fingerprint\x1a\f.proto.Empty\x12$\n" +
	"\aAddScan\x12\v.proto.Scan\x1a\f.proto.Empty\x12(\n" +
	"\tAddSource\x12\r.proto.Source\x1a\f.proto.Empty\x122\n" +
	"\tLabelHost\x12\x17.proto.LabelHostRequest\x1a\f.proto.Empty\x122\n" +
	"\x05Query\x12\x13.proto.QueryRequest\x1a\x14.proto.QueryResponseB\x05Z\x03pb/b\x06proto3"

var (
	file_proto_module_proto_rawDescOnce sync.Once
	file_proto_module_proto_rawDescData []byte
)

func file_proto_module_proto_rawDescGZIP() []byte {
	file_proto_module_proto_rawDescOnce.Do(func() {
		file_proto_module_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_module_proto_rawDesc), len(file_proto_module_proto_rawDesc)))
	})
	return file_proto_module_proto_rawDescData
}

var file_proto_module_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_proto_module_proto_goTypes = []any{
	(*Empty)(nil),            // 0: proto.Empty
	(*Fields)(nil),           // 1: proto.Fields
	(*Event)(nil),            // 2: proto.Event
	(*HandleRequest)(nil),    // 3: proto.HandleRequest
	(*IDRequest)(nil),        // 4: proto.IDRequest
	(*QueryRequest)(nil),     // 5: proto.QueryRequest
	(*QueryResponse)(nil),    // 6: proto.QueryResponse
	(*LabelHostRequest)(nil), // 7: proto.LabelHostRequest
	(*Host)(nil),             // 8: proto.Host
	(*Label)(nil),            // 9: proto.Label
	(*Fingerprint)(nil),      // 10: proto.Fingerprint
	(*Scan)(nil),             // 11: proto.Scan
	(*Source)(nil),           // 12: proto.Source
}
var file_proto_module_proto_depIdxs = []int32{
	2,  // 0: proto.HandleRequest.event:type_name -> proto.Event
	8,  // 1: proto.QueryResponse.hosts:type_name -> proto.Host
	0,  // 2: proto.Module.Propagate:input_type -> proto.Empty
	0,  // 3: proto.Module.Properties:input_type -> proto.Empty
	3,  // 4: proto.Module.Handle:input_type -> proto.HandleRequest
	4,  // 5: proto.Adapter.GetHost:input_type -> proto.IDRequest
	4,  // 6: proto.Adapter.GetSource:input_type -> proto.IDRequest
	4,  // 7: proto.Adapter.GetScan:input_type -> proto.IDRequest
	9,  // 8: proto.Adapter.AddLabel:input_type -> proto.Label
	10, // 9: proto.Adapter.AddFingerprint:input_type -> proto.Fingerprint
	11, // 10: proto.Adapter.AddScan:input_type -> proto.Scan
	12, // 11: proto.Adapter.AddSource:input_type -> proto.Source
	7,  // 12: proto.Adapter.LabelHost:input_type -> proto.LabelHostRequest
	5,  // 13: proto.Adapter.Query:input_type -> proto.QueryRequest
	0,  // 14: proto.Module.Propagate:output_type -> proto.Empty
	1,  // 15: proto.Module.Properties:output_type -> proto.Fields
	0,  // 16: proto.Module.Handle:output_type -> proto.Empty
	8,  // 17: proto.Adapter.GetHost:output_type -> proto.Host
	12, // 18: proto.Adapter.GetSource:output_type -> proto.Source
	11, // 19: proto.Adapter.GetScan:output_type -> proto.Scan
	0,  // 20: proto.Adapter.AddLabel:output_type -> proto.Empty
	0,  // 21: proto.Adapter.AddFingerprint:output_type -> proto.Empty
	0,  // 22: proto.Adapter.AddScan:output_type -> proto.Empty
	0,  // 23: proto.Adapter.AddSource:output_type -> proto.Empty
	0,  // 24: proto.Adapter.LabelHost:output_type -> proto.Empty
	6,  // 25: proto.Adapter.Query:output_type -> proto.QueryResponse
	14, // [14:26] is the sub-list for method output_type
	2,  // [2:14] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_proto_module_proto_init() }
func file_proto_module_proto_init() {
	if File_proto_module_proto != nil {
		return
	}
	file_proto_models_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_module_proto_rawDesc), len(file_proto_module_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_proto_module_proto_goTypes,
		DependencyIndexes: file_proto_module_proto_depIdxs,
		MessageInfos:      file_proto_module_proto_msgTypes,
	}.Build()
	File_proto_module_proto = out.File
	file_proto_module_proto_goTypes = nil
	file_proto_module_proto_depIdxs = nil
}
