// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuffer

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type File struct {
	_tab flatbuffers.Table
}

func GetRootAsFile(buf []byte, offset flatbuffers.UOffsetT) *File {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &File{}
	x.Init(buf, n+offset)
	return x
}

func FinishFileBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsFile(buf []byte, offset flatbuffers.UOffsetT) *File {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &File{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedFileBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *File) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *File) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *File) Id() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateId(n int64) bool {
	return rcv._tab.MutateInt64Slot(4, n)
}

func (rcv *File) DirectoryId() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateDirectoryId(n int64) bool {
	return rcv._tab.MutateInt64Slot(6, n)
}

func (rcv *File) Size() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateSize(n uint32) bool {
	return rcv._tab.MutateUint32Slot(8, n)
}

func (rcv *File) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *File) TagBitmask() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateTagBitmask(n uint64) bool {
	return rcv._tab.MutateUint64Slot(12, n)
}

func (rcv *File) Unk5() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateUnk5(n byte) bool {
	return rcv._tab.MutateByteSlot(14, n)
}

func (rcv *File) Unk6() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateUnk6(n byte) bool {
	return rcv._tab.MutateByteSlot(16, n)
}

func (rcv *File) ChunkIds(j int) int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetInt64(a + flatbuffers.UOffsetT(j*8))
	}
	return 0
}

func (rcv *File) ChunkIdsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *File) MutateChunkIds(j int, n int64) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.MutateInt64(a+flatbuffers.UOffsetT(j*8), n)
	}
	return false
}

func (rcv *File) Unk8() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateUnk8(n byte) bool {
	return rcv._tab.MutateByteSlot(20, n)
}

func (rcv *File) Symlink() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *File) Unk10() uint16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		return rcv._tab.GetUint16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateUnk10(n uint16) bool {
	return rcv._tab.MutateUint16Slot(24, n)
}

func (rcv *File) ChunkingParamId() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(26))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutateChunkingParamId(n byte) bool {
	return rcv._tab.MutateByteSlot(26, n)
}

func (rcv *File) Permissions() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(28))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *File) MutatePermissions(n byte) bool {
	return rcv._tab.MutateByteSlot(28, n)
}

func FileStart(builder *flatbuffers.Builder) {
	builder.StartObject(13)
}
func FileAddId(builder *flatbuffers.Builder, id int64) {
	builder.PrependInt64Slot(0, id, 0)
}
func FileAddDirectoryId(builder *flatbuffers.Builder, directoryId int64) {
	builder.PrependInt64Slot(1, directoryId, 0)
}
func FileAddSize(builder *flatbuffers.Builder, size uint32) {
	builder.PrependUint32Slot(2, size, 0)
}
func FileAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(name), 0)
}
func FileAddTagBitmask(builder *flatbuffers.Builder, tagBitmask uint64) {
	builder.PrependUint64Slot(4, tagBitmask, 0)
}
func FileAddUnk5(builder *flatbuffers.Builder, unk5 byte) {
	builder.PrependByteSlot(5, unk5, 0)
}
func FileAddUnk6(builder *flatbuffers.Builder, unk6 byte) {
	builder.PrependByteSlot(6, unk6, 0)
}
func FileAddChunkIds(builder *flatbuffers.Builder, chunkIds flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(7, flatbuffers.UOffsetT(chunkIds), 0)
}
func FileStartChunkIdsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(8, numElems, 8)
}
func FileAddUnk8(builder *flatbuffers.Builder, unk8 byte) {
	builder.PrependByteSlot(8, unk8, 0)
}
func FileAddSymlink(builder *flatbuffers.Builder, symlink flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(9, flatbuffers.UOffsetT(symlink), 0)
}
func FileAddUnk10(builder *flatbuffers.Builder, unk10 uint16) {
	builder.PrependUint16Slot(10, unk10, 0)
}
func FileAddChunkingParamId(builder *flatbuffers.Builder, chunkingParamId byte) {
	builder.PrependByteSlot(11, chunkingParamId, 0)
}
func FileAddPermissions(builder *flatbuffers.Builder, permissions byte) {
	builder.PrependByteSlot(12, permissions, 0)
}
func FileEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
