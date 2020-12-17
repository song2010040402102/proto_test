package util

import (
	"github.com/golang/protobuf/proto"

	"proto_test/protobuf"
)

func ProtoMsg2Buff(t int32, msg proto.Message) []byte {
	data, _ := proto.Marshal(msg)
	buff := make([]byte, 4, len(data)+4)
	buff[0] = byte(t & 0xff)
	buff[1] = byte(t >> 8 & 0xff)
	buff[2] = byte(t >> 16 & 0xff)
	buff[3] = byte(t >> 24 & 0xff)
	return append(buff, data...)
}

func Buff2ProtoMsg(buff []byte) (int32, proto.Message) {
	t := int32(buff[0]) | int32(buff[1])<<8 | int32(buff[2])<<16 | int32(buff[3])<<24
	msg := protobuf.GetMessageObjectById(t)
	proto.Unmarshal(buff[4:], msg)
	return t, msg
}
