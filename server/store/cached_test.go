package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoredIndexCompareTo(t *testing.T) {
	v1 := 1
	v2 := 2
	v3 := 3
	v5 := 5
	before := StoredIndex[int]{
		Data: []CachedIndexEntry[int]{
			{Key: "key1", ValueHash: "hash1", data: &v1},
			{Key: "key2", ValueHash: "hash2", data: &v2},
			{Key: "key3", ValueHash: "hash3", data: &v3},
		},
	}
	after := StoredIndex[int]{
		Data: []CachedIndexEntry[int]{
			{Key: "key1", ValueHash: "hash1", data: &v1},
			{Key: "key3", ValueHash: "hash5", data: &v5},
			{Key: "key4", ValueHash: "hash4", data: &v3},
		},
	}

	change, remove := before.compareTo(&after)
	require.Equal(t, []CachedIndexEntry[int]{
		{Key: "key2", ValueHash: "hash2", data: &v2},
	}, remove)
	require.Equal(t, []CachedIndexEntry[int]{
		{Key: "key3", ValueHash: "hash5", data: &v5},
		{Key: "key4", ValueHash: "hash4", data: &v3},
	}, change)
}
