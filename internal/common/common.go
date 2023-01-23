package common

import (
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"
)

// TreeEntryFromCoreClaim convert core.Claim to merkletree.Entry
func TreeEntryFromCoreClaim(c core.Claim) merkletree.Entry {
	index, value := c.RawSlots()

	var e merkletree.Entry

	e.Data[0] = ElemBytesCoreToMT(index[0])
	e.Data[1] = ElemBytesCoreToMT(index[1])
	e.Data[2] = ElemBytesCoreToMT(index[2])
	e.Data[3] = ElemBytesCoreToMT(index[3])

	e.Data[4] = ElemBytesCoreToMT(value[0])
	e.Data[5] = ElemBytesCoreToMT(value[1])
	e.Data[6] = ElemBytesCoreToMT(value[2])
	e.Data[7] = ElemBytesCoreToMT(value[3])

	return e
}

// ElemBytesCoreToMT coverts core.ElemBytes to merkletree.ElemBytes
func ElemBytesCoreToMT(ebCore core.ElemBytes) merkletree.ElemBytes {
	var ebMT merkletree.ElemBytes
	copy(ebMT[:], ebCore[:])
	return ebMT
}
