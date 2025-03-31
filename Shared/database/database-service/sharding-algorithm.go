package database

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func MyUUIDSuffixAlgorithm(columnValue any) (string, error) {
	var parsed uuid.UUID
	switch v := columnValue.(type) {
	case string:
		// Already a string, parse as UUID:
		u, err := uuid.Parse(strings.TrimSpace(v))
		if err != nil {
			return "", fmt.Errorf("failed to parse '%s' as UUID: %v", v, err)
		}
		parsed = u

	case uuid.UUID:
		// Already a UUID value
		parsed = v

	case *uuid.UUID:
		// A pointer to UUID
		if v == nil {
			return "", fmt.Errorf("nil *uuid.UUID")
		}
		parsed = *v

	case []byte:
		// Possibly got raw bytes
		u, err := uuid.FromBytes(v)
		if err != nil {
			return "", fmt.Errorf("failed to parse columnValue bytes as UUID: %v", err)
		}
		parsed = u

	default:
		return "", fmt.Errorf("unsupported type %T; need string, uuid.UUID, or *uuid.UUID", v)
	}

	// Now "parsed" is a proper uuid.UUID
	// Next, do your shard logic. For example:
	shardIndex := hashUUID(parsed) % 64
	suffix := fmt.Sprintf("_%04d", shardIndex)
	return suffix, nil
}

// For example, hash the first 8 bytes of a SHA-256 of the UUID:
func hashUUID(u uuid.UUID) uint64 {
	h := sha256.Sum256(u[:]) // 16-byte input
	val := binary.BigEndian.Uint64(h[:8])
	return val
}
