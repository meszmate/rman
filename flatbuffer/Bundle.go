// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuffer

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Bundle struct {
	_tab flatbuffers.Table
}

func GetRootAsBundle(buf []byte, offset flatbuffers.UOffsetT) *Bundle {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Bundle{}
	x.Init(buf, n+offset)
	return x
}

func FinishBundleBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsBundle(buf []byte, offset flatbuffers.UOffsetT) *Bundle {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Bundle{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedBundleBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *Bundle) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Bundle) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Bundle) BundleId() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Bundle) MutateBundleId(n uint64) bool {
	return rcv._tab.MutateUint64Slot(4, n)
}

func (rcv *Bundle) Chunks(obj *Chunk, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Bundle) ChunksLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func BundleStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func BundleAddBundleId(builder *flatbuffers.Builder, bundleId uint64) {
	builder.PrependUint64Slot(0, bundleId, 0)
}
func BundleAddChunks(builder *flatbuffers.Builder, chunks flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(chunks), 0)
}
func BundleStartChunksVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func BundleEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
