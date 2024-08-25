// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuffers

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ChunkingParam struct {
	_tab flatbuffers.Table
}

func GetRootAsChunkingParam(buf []byte, offset flatbuffers.UOffsetT) *ChunkingParam {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ChunkingParam{}
	x.Init(buf, n+offset)
	return x
}

func FinishChunkingParamBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsChunkingParam(buf []byte, offset flatbuffers.UOffsetT) *ChunkingParam {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &ChunkingParam{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedChunkingParamBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *ChunkingParam) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ChunkingParam) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ChunkingParam) Unk0() uint16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ChunkingParam) MutateUnk0(n uint16) bool {
	return rcv._tab.MutateUint16Slot(4, n)
}

func (rcv *ChunkingParam) ChunkingVersion() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ChunkingParam) MutateChunkingVersion(n byte) bool {
	return rcv._tab.MutateByteSlot(6, n)
}

func (rcv *ChunkingParam) MinChunkSize() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ChunkingParam) MutateMinChunkSize(n uint32) bool {
	return rcv._tab.MutateUint32Slot(8, n)
}

func (rcv *ChunkingParam) ChunkSize() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ChunkingParam) MutateChunkSize(n uint32) bool {
	return rcv._tab.MutateUint32Slot(10, n)
}

func (rcv *ChunkingParam) MaxChunkSize() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ChunkingParam) MutateMaxChunkSize(n uint32) bool {
	return rcv._tab.MutateUint32Slot(12, n)
}

func ChunkingParamStart(builder *flatbuffers.Builder) {
	builder.StartObject(5)
}
func ChunkingParamAddUnk0(builder *flatbuffers.Builder, unk0 uint16) {
	builder.PrependUint16Slot(0, unk0, 0)
}
func ChunkingParamAddChunkingVersion(builder *flatbuffers.Builder, chunkingVersion byte) {
	builder.PrependByteSlot(1, chunkingVersion, 0)
}
func ChunkingParamAddMinChunkSize(builder *flatbuffers.Builder, minChunkSize uint32) {
	builder.PrependUint32Slot(2, minChunkSize, 0)
}
func ChunkingParamAddChunkSize(builder *flatbuffers.Builder, chunkSize uint32) {
	builder.PrependUint32Slot(3, chunkSize, 0)
}
func ChunkingParamAddMaxChunkSize(builder *flatbuffers.Builder, maxChunkSize uint32) {
	builder.PrependUint32Slot(4, maxChunkSize, 0)
}
func ChunkingParamEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
