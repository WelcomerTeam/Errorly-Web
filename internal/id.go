package errorly

import (
	"sync"
	"time"
)

// NewIDGenerator returns an IDGenerator
func NewIDGenerator(initialEpoch int64, shardID int64) *IDGenerator {
	return &IDGenerator{
		initialEpoch: initialEpoch,
		shardID:      shardID,
		sequence:     0,
	}
}

// IDGenerator contains the structure of the id generator.
// IDs are comprised of a similar structure to intagram where 41 bits
// contain the timestamp then a further 23 bits for a shard and sequence.
type IDGenerator struct {
	sync.Mutex
	initialEpoch int64
	shardID      int64
	sequence     int64
}

// GenerateID returns a new id that is int64
func (id *IDGenerator) GenerateID() int64 {
	id.Lock()
	defer id.Unlock()

	ms := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	id.sequence++ // We only want 10 bits
	return ((ms - id.initialEpoch) << 23) | (id.shardID << 10) | (id.sequence % 1024)
}
