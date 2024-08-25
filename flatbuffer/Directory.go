// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuffer

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Directory struct {
	_tab flatbuffers.Table
}

func GetRootAsDirectory(buf []byte, offset flatbuffers.UOffsetT) *Directory {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Directory{}
	x.Init(buf, n+offset)
	return x
}

func FinishDirectoryBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsDirectory(buf []byte, offset flatbuffers.UOffsetT) *Directory {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Directory{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedDirectoryBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *Directory) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Directory) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Directory) Id() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Directory) MutateId(n uint64) bool {
	return rcv._tab.MutateUint64Slot(4, n)
}

func (rcv *Directory) ParentId() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Directory) MutateParentId(n uint64) bool {
	return rcv._tab.MutateUint64Slot(6, n)
}

func (rcv *Directory) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func DirectoryStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func DirectoryAddId(builder *flatbuffers.Builder, id uint64) {
	builder.PrependUint64Slot(0, id, 0)
}
func DirectoryAddParentId(builder *flatbuffers.Builder, parentId uint64) {
	builder.PrependUint64Slot(1, parentId, 0)
}
func DirectoryAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(name), 0)
}
func DirectoryEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
