// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuffers

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Tag struct {
	_tab flatbuffers.Table
}

func GetRootAsTag(buf []byte, offset flatbuffers.UOffsetT) *Tag {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Tag{}
	x.Init(buf, n+offset)
	return x
}

func FinishTagBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsTag(buf []byte, offset flatbuffers.UOffsetT) *Tag {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Tag{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedTagBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *Tag) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Tag) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Tag) Id() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Tag) MutateId(n byte) bool {
	return rcv._tab.MutateByteSlot(4, n)
}

func (rcv *Tag) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func TagStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func TagAddId(builder *flatbuffers.Builder, id byte) {
	builder.PrependByteSlot(0, id, 0)
}
func TagAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(name), 0)
}
func TagEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
