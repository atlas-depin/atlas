package types

import (
	"time"

	"cosmossdk.io/math"
)

type StorageSubscription struct {
	Start          time.Time
	End            time.Time
	SpaceAvailable int64
	SpaceUsed      int64
	Credits        math.Int
}

type StoragePaymentInfo struct {
	Start          time.Time `protobuf:"bytes,1,opt,name=start,proto3,stdtime" json:"start"`
	End            time.Time `protobuf:"bytes,2,opt,name=end,proto3,stdtime" json:"end"`
	SpaceAvailable int64     `protobuf:"varint,3,opt,name=spaceAvailable,proto3" json:"spaceAvailable,omitempty"`
	SpaceUsed      int64     `protobuf:"varint,4,opt,name=spaceUsed,proto3" json:"spaceUsed,omitempty"`
	Address        string    `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
}

type StorageAccountInfo struct {
	Credits math.Int
}
