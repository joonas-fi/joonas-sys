package gostree

import (
	"time"

	"github.com/joonas-fi/joonas-sys/pkg/gvariant"
)

// ------------- commit

type pair struct {
	Key   string
	Value gvariant.Variant
}

type someTuple struct {
	K string
	V []byte
}

// "Schema" defined here
// https://github.com/ostreedev/ostree/blob/8aaea0c65ddcf32d4f52efb707185c25354c5e42/src/libostree/ostree-repo-commit.c#L2965
type Commit struct {
	Metadata     []pair      // @a{sv}
	ParentCommit []byte      // @ay
	Something    []someTuple // @a(say)
	Subject      string      // s
	Body         string      // s
	Timestamp    uint64      // t
	Dirtree      []byte      // @ay
	DirtreeMeta  []byte      // @ay
}

func (c Commit) GetTimestamp() time.Time {
	return time.Unix(int64(c.Timestamp), 0)
}

// ------------- dirtree

type objectEntry struct {
	Name     string
	Checksum []byte
}

type dirEntry struct {
	Name            string
	DirtreeChecksum []byte
	DirmetaChecksum []byte
}

// https://ostreedev.github.io/ostree/repo/#dirtree-objects
type Dirtree struct {
	Files       []objectEntry // "content objects"
	Directories []dirEntry
}

// ------------- dirmeta

type dirmetaItem struct {
	Unknown1 []byte
	Unknown2 []byte
}

type dirmetaStruct struct {
	UID      uint32
	GID      uint32
	Unknown  uint32        // mode?
	Children []dirmetaItem // @a(ayay)
}
