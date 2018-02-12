package reflect2

import (
	"unsafe"
	"reflect"
)

// sliceHeader is a safe version of SliceHeader used within this package.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

type UnsafeSliceType struct {
	unsafeType
	elemRType  unsafe.Pointer
	pElemRType unsafe.Pointer
	elemSize   uintptr
}

func newUnsafeSliceType(cfg *frozenConfig, type1 reflect.Type) SliceType {
	elemType := type1.Elem()
	return &UnsafeSliceType{
		unsafeType: *newUnsafeType(cfg, type1),
		pElemRType: unpackEFace(reflect.PtrTo(elemType)).data,
		elemRType:  unpackEFace(elemType).data,
		elemSize:   elemType.Size(),
	}
}

func (type2 *UnsafeSliceType) MakeSlice(length int, cap int) interface{} {
	return packEFace(type2.ptrRType, type2.UnsafeMakeSlice(length, cap))
}

func (type2 *UnsafeSliceType) UnsafeMakeSlice(length int, cap int) unsafe.Pointer {
	header := &sliceHeader{unsafe_NewArray(type2.elemRType, cap), length, cap}
	return unsafe.Pointer(header)
}

func (type2 *UnsafeSliceType) Len(obj interface{}) int {
	objEFace := unpackEFace(obj)
	assertType("SliceType.Len argument 1", type2.ptrRType, objEFace.rtype)
	return type2.UnsafeLen(objEFace.data)
}

func (type2 *UnsafeSliceType) UnsafeLen(obj unsafe.Pointer) int {
	header := (*sliceHeader)(obj)
	return header.Len
}

func (type2 *UnsafeSliceType) Set(obj interface{}, index int, elem interface{}) {
	objEFace := unpackEFace(obj)
	assertType("SliceType.Set argument 1", type2.ptrRType, objEFace.rtype)
	elemEFace := unpackEFace(elem)
	assertType("SliceType.Set argument 3", type2.pElemRType, elemEFace.rtype)
	type2.UnsafeSet(objEFace.data, index, elemEFace.data)
}

func (type2 *UnsafeSliceType) UnsafeSet(obj unsafe.Pointer, index int, elem unsafe.Pointer) {
	header := (*sliceHeader)(obj)
	elemPtr := arrayAt(header.Data, index, type2.elemSize, "i < s.Len")
	typedmemmove(type2.elemRType, elemPtr, elem)
}

func (type2 *UnsafeSliceType) Get(obj interface{}, index int) interface{} {
	objEFace := unpackEFace(obj)
	assertType("SliceType.Get argument 1", type2.ptrRType, objEFace.rtype)
	elemPtr := type2.UnsafeGet(objEFace.data, index)
	return packEFace(type2.pElemRType, elemPtr)
}

func (type2 *UnsafeSliceType) UnsafeGet(obj unsafe.Pointer, index int) unsafe.Pointer {
	header := (*sliceHeader)(obj)
	return arrayAt(header.Data, index, type2.elemSize, "i < s.Len")
}

func (type2 *UnsafeSliceType) Append(obj interface{}, elem interface{}) interface{} {
	objEFace := unpackEFace(obj)
	assertType("SliceType.Append argument 1", type2.ptrRType, objEFace.rtype)
	elemEFace := unpackEFace(elem)
	assertType("SliceType.Append argument 2", type2.pElemRType, elemEFace.rtype)
	ptr := type2.UnsafeAppend(objEFace.data, elemEFace.data)
	return packEFace(type2.ptrRType, ptr)
}

func (type2 *UnsafeSliceType) UnsafeAppend(obj unsafe.Pointer, elem unsafe.Pointer) unsafe.Pointer {
	header := (*sliceHeader)(obj)
	if header.Cap == header.Len {
		header = type2.grow(header, header.Len+1)
	}
	type2.UnsafeSet(unsafe.Pointer(header), header.Len, elem)
	header.Len += 1
	return unsafe.Pointer(header)
}

func (type2 *UnsafeSliceType) grow(header *sliceHeader, expectedCap int) *sliceHeader {
	newCap := calcNewCap(header.Cap, expectedCap)
	newHeader := (*sliceHeader)(type2.UnsafeMakeSlice(header.Len, newCap))
	typedslicecopy(type2.elemRType, *newHeader, *header)
	return newHeader
}

func calcNewCap(cap int, expectedCap int) int {
	if cap == 0 {
		cap = expectedCap
	} else {
		for cap < expectedCap {
			if cap < 1024 {
				cap += cap
			} else {
				cap += cap / 4
			}
		}
	}
	return cap
}
