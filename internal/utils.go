package errorly

import (
	"encoding/binary"

	"github.com/btcsuite/btcutil/base58"
)

func idFromUInt64(i uint64) string {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buf, i)

	return base58.Encode(buf)
}

func uint64FromID(id string) uint64 {
	return binary.LittleEndian.Uint64(base58.Decode(id))
}
