// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.1
// 	protoc        v3.21.12
// source: proto/monitoringNotification.proto

package proto

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

type MonitoringDeploymentRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	JobName        string                 `protobuf:"bytes,1,opt,name=job_name,json=jobName,proto3" json:"job_name,omitempty"`
	JobHash        string                 `protobuf:"bytes,2,opt,name=job_hash,json=jobHash,proto3" json:"job_hash,omitempty"`
	InstanceNumber int32                  `protobuf:"varint,3,opt,name=instance_number,json=instanceNumber,proto3" json:"instance_number,omitempty"`
	Resource       *ResourceInfo          `protobuf:"bytes,4,opt,name=resource,proto3" json:"resource,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *MonitoringDeploymentRequest) Reset() {
	*x = MonitoringDeploymentRequest{}
	mi := &file_proto_monitoringNotification_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MonitoringDeploymentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MonitoringDeploymentRequest) ProtoMessage() {}

func (x *MonitoringDeploymentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_monitoringNotification_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MonitoringDeploymentRequest.ProtoReflect.Descriptor instead.
func (*MonitoringDeploymentRequest) Descriptor() ([]byte, []int) {
	return file_proto_monitoringNotification_proto_rawDescGZIP(), []int{0}
}

func (x *MonitoringDeploymentRequest) GetJobName() string {
	if x != nil {
		return x.JobName
	}
	return ""
}

func (x *MonitoringDeploymentRequest) GetJobHash() string {
	if x != nil {
		return x.JobHash
	}
	return ""
}

func (x *MonitoringDeploymentRequest) GetInstanceNumber() int32 {
	if x != nil {
		return x.InstanceNumber
	}
	return 0
}

func (x *MonitoringDeploymentRequest) GetResource() *ResourceInfo {
	if x != nil {
		return x.Resource
	}
	return nil
}

type ResourceInfo struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Cpu           string                 `protobuf:"bytes,1,opt,name=cpu,proto3" json:"cpu,omitempty"`
	Memory        string                 `protobuf:"bytes,2,opt,name=memory,proto3" json:"memory,omitempty"`
	Gpu           string                 `protobuf:"bytes,3,opt,name=gpu,proto3" json:"gpu,omitempty"`
	Disk          string                 `protobuf:"bytes,4,opt,name=disk,proto3" json:"disk,omitempty"`
	Network       *NetworkInfo           `protobuf:"bytes,5,opt,name=network,proto3" json:"network,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ResourceInfo) Reset() {
	*x = ResourceInfo{}
	mi := &file_proto_monitoringNotification_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ResourceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceInfo) ProtoMessage() {}

func (x *ResourceInfo) ProtoReflect() protoreflect.Message {
	mi := &file_proto_monitoringNotification_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceInfo.ProtoReflect.Descriptor instead.
func (*ResourceInfo) Descriptor() ([]byte, []int) {
	return file_proto_monitoringNotification_proto_rawDescGZIP(), []int{1}
}

func (x *ResourceInfo) GetCpu() string {
	if x != nil {
		return x.Cpu
	}
	return ""
}

func (x *ResourceInfo) GetMemory() string {
	if x != nil {
		return x.Memory
	}
	return ""
}

func (x *ResourceInfo) GetGpu() string {
	if x != nil {
		return x.Gpu
	}
	return ""
}

func (x *ResourceInfo) GetDisk() string {
	if x != nil {
		return x.Disk
	}
	return ""
}

func (x *ResourceInfo) GetNetwork() *NetworkInfo {
	if x != nil {
		return x.Network
	}
	return nil
}

type NetworkInfo struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BandwidthIn   string                 `protobuf:"bytes,1,opt,name=bandwidth_in,json=bandwidthIn,proto3" json:"bandwidth_in,omitempty"`
	BandwidthOut  string                 `protobuf:"bytes,2,opt,name=bandwidth_out,json=bandwidthOut,proto3" json:"bandwidth_out,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *NetworkInfo) Reset() {
	*x = NetworkInfo{}
	mi := &file_proto_monitoringNotification_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NetworkInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkInfo) ProtoMessage() {}

func (x *NetworkInfo) ProtoReflect() protoreflect.Message {
	mi := &file_proto_monitoringNotification_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkInfo.ProtoReflect.Descriptor instead.
func (*NetworkInfo) Descriptor() ([]byte, []int) {
	return file_proto_monitoringNotification_proto_rawDescGZIP(), []int{2}
}

func (x *NetworkInfo) GetBandwidthIn() string {
	if x != nil {
		return x.BandwidthIn
	}
	return ""
}

func (x *NetworkInfo) GetBandwidthOut() string {
	if x != nil {
		return x.BandwidthOut
	}
	return ""
}

type MonitoringResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Acknowledged  bool                   `protobuf:"varint,1,opt,name=acknowledged,proto3" json:"acknowledged,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MonitoringResponse) Reset() {
	*x = MonitoringResponse{}
	mi := &file_proto_monitoringNotification_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MonitoringResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MonitoringResponse) ProtoMessage() {}

func (x *MonitoringResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_monitoringNotification_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MonitoringResponse.ProtoReflect.Descriptor instead.
func (*MonitoringResponse) Descriptor() ([]byte, []int) {
	return file_proto_monitoringNotification_proto_rawDescGZIP(), []int{3}
}

func (x *MonitoringResponse) GetAcknowledged() bool {
	if x != nil {
		return x.Acknowledged
	}
	return false
}

func (x *MonitoringResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type MonitoringDeletionRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	JobName        string                 `protobuf:"bytes,1,opt,name=job_name,json=jobName,proto3" json:"job_name,omitempty"`
	InstanceNumber int32                  `protobuf:"varint,2,opt,name=instance_number,json=instanceNumber,proto3" json:"instance_number,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *MonitoringDeletionRequest) Reset() {
	*x = MonitoringDeletionRequest{}
	mi := &file_proto_monitoringNotification_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MonitoringDeletionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MonitoringDeletionRequest) ProtoMessage() {}

func (x *MonitoringDeletionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_monitoringNotification_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MonitoringDeletionRequest.ProtoReflect.Descriptor instead.
func (*MonitoringDeletionRequest) Descriptor() ([]byte, []int) {
	return file_proto_monitoringNotification_proto_rawDescGZIP(), []int{4}
}

func (x *MonitoringDeletionRequest) GetJobName() string {
	if x != nil {
		return x.JobName
	}
	return ""
}

func (x *MonitoringDeletionRequest) GetInstanceNumber() int32 {
	if x != nil {
		return x.InstanceNumber
	}
	return 0
}

var File_proto_monitoringNotification_proto protoreflect.FileDescriptor

var file_proto_monitoringNotification_proto_rawDesc = []byte{
	0x0a, 0x22, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69,
	0x6e, 0x67, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67,
	0x22, 0xb2, 0x01, 0x0a, 0x1b, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x44,
	0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x19, 0x0a, 0x08, 0x6a, 0x6f, 0x62, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6a, 0x6f, 0x62, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x6a,
	0x6f, 0x62, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6a,
	0x6f, 0x62, 0x48, 0x61, 0x73, 0x68, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12,
	0x34, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x18, 0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x91, 0x01, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x70, 0x75, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x63, 0x70, 0x75, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x6d, 0x6f,
	0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x70, 0x75, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67,
	0x70, 0x75, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x69, 0x73, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x64, 0x69, 0x73, 0x6b, 0x12, 0x31, 0x0a, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72,
	0x6b, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x22, 0x55, 0x0a, 0x0b, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x61, 0x6e, 0x64,
	0x77, 0x69, 0x64, 0x74, 0x68, 0x5f, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x62, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x64, 0x74, 0x68, 0x49, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x62,
	0x61, 0x6e, 0x64, 0x77, 0x69, 0x64, 0x74, 0x68, 0x5f, 0x6f, 0x75, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x62, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x64, 0x74, 0x68, 0x4f, 0x75, 0x74,
	0x22, 0x52, 0x0a, 0x12, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x63, 0x6b, 0x6e, 0x6f, 0x77,
	0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x61, 0x63,
	0x6b, 0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x5f, 0x0a, 0x19, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69,
	0x6e, 0x67, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x19, 0x0a, 0x08, 0x6a, 0x6f, 0x62, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6a, 0x6f, 0x62, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x27, 0x0a, 0x0f,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x32, 0xcd, 0x01, 0x0a, 0x11, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x5d, 0x0a, 0x10, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12,
	0x27, 0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x4d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x59, 0x0a, 0x0e, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x2e, 0x6d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67,
	0x2e, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x50, 0x5a, 0x4e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x6e, 0x7a, 0x6c, 0x6e, 0x73, 0x6b, 0x2f, 0x6f, 0x70, 0x65,
	0x6e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x2d, 0x63, 0x6f, 0x6d, 0x70, 0x6f,
	0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x2f,
	0x6f, 0x61, 0x6b, 0x65, 0x73, 0x74, 0x72, 0x61, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x6f,
	0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_monitoringNotification_proto_rawDescOnce sync.Once
	file_proto_monitoringNotification_proto_rawDescData = file_proto_monitoringNotification_proto_rawDesc
)

func file_proto_monitoringNotification_proto_rawDescGZIP() []byte {
	file_proto_monitoringNotification_proto_rawDescOnce.Do(func() {
		file_proto_monitoringNotification_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_monitoringNotification_proto_rawDescData)
	})
	return file_proto_monitoringNotification_proto_rawDescData
}

var file_proto_monitoringNotification_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_monitoringNotification_proto_goTypes = []any{
	(*MonitoringDeploymentRequest)(nil), // 0: monitoring.MonitoringDeploymentRequest
	(*ResourceInfo)(nil),                // 1: monitoring.ResourceInfo
	(*NetworkInfo)(nil),                 // 2: monitoring.NetworkInfo
	(*MonitoringResponse)(nil),          // 3: monitoring.MonitoringResponse
	(*MonitoringDeletionRequest)(nil),   // 4: monitoring.MonitoringDeletionRequest
}
var file_proto_monitoringNotification_proto_depIdxs = []int32{
	1, // 0: monitoring.MonitoringDeploymentRequest.resource:type_name -> monitoring.ResourceInfo
	2, // 1: monitoring.ResourceInfo.network:type_name -> monitoring.NetworkInfo
	0, // 2: monitoring.MonitoringService.NotifyDeployment:input_type -> monitoring.MonitoringDeploymentRequest
	4, // 3: monitoring.MonitoringService.NotifyDeletion:input_type -> monitoring.MonitoringDeletionRequest
	3, // 4: monitoring.MonitoringService.NotifyDeployment:output_type -> monitoring.MonitoringResponse
	3, // 5: monitoring.MonitoringService.NotifyDeletion:output_type -> monitoring.MonitoringResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_monitoringNotification_proto_init() }
func file_proto_monitoringNotification_proto_init() {
	if File_proto_monitoringNotification_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_monitoringNotification_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_monitoringNotification_proto_goTypes,
		DependencyIndexes: file_proto_monitoringNotification_proto_depIdxs,
		MessageInfos:      file_proto_monitoringNotification_proto_msgTypes,
	}.Build()
	File_proto_monitoringNotification_proto = out.File
	file_proto_monitoringNotification_proto_rawDesc = nil
	file_proto_monitoringNotification_proto_goTypes = nil
	file_proto_monitoringNotification_proto_depIdxs = nil
}
