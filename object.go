package canopen

// Index represents the 2-byte index in an object index.
type Index struct {
	B0 byte
	B1 byte
}

// ObjectIndex represents the index of an object.
type ObjectIndex struct {
	Index    Index
	SubIndex byte
}

// NewObjectIndex returns an object index from a 2-byte index and 1-byte sub index.
func NewObjectIndex(index uint16, subIndex uint8) ObjectIndex {
	return ObjectIndex{
		Index: Index{
			B0: byte(index & 0xFF),
			B1: byte(index >> 8),
		},
		SubIndex: subIndex,
	}
}

func (objectIndex *ObjectIndex) Compare(other ObjectIndex) bool {
	return objectIndex.Index.B0 == other.Index.B0 && objectIndex.Index.B1 == other.Index.B1 && objectIndex.SubIndex == other.SubIndex
}

func (objectIndex *ObjectIndex) Bytes() []byte {
	return []byte{
		objectIndex.Index.B0,
		objectIndex.Index.B1,
		objectIndex.SubIndex,
	}
}

func (index *Index) Index() uint16 {
	return uint16(index.B0) + (uint16(index.B1) << 8)
}
