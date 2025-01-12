package gostree

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/function61/gokit/testing/assert"
	"github.com/joonas-fi/joonas-sys/pkg/gvariant"
)

//go:embed testdata/7a2e51473983e374c7c6e331ae376b62cea277dd6b67c58b4dfbaa5f164b176b.commit
var exampleCommitSerialized []byte

func TestDeserializeCommit(t *testing.T) {
	exampleCommit := Commit{}
	assert.Ok(t, gvariant.Unmarshal(exampleCommitSerialized, &exampleCommit, binary.BigEndian))

	assert.Equal(t, len(exampleCommit.Metadata), 1)
	assert.Equal(t, exampleCommit.Metadata[0].Key, "ostree.ref-binding")
	assert.Equal(t, exampleCommit.Metadata[0].Value.Format, "as")
	// FIXME: decoding fails due to garbage here?
	assert.Equal(t, string(exampleCommit.Metadata[0].Value.Data), "deploy/app/fi.joonas.os/x86_64/stable\x00&")
	assert.Equal(t, fmt.Sprintf("%x", exampleCommit.ParentCommit), "8fc4f77c86abba308795e62f58711bad8620afb9c41562f123de135187e5e746")
	assert.Equal(t, exampleCommit.Subject, "start adding Varasto mounts")
	assert.Equal(t, exampleCommit.Body, "")
	assert.Equal(t, exampleCommit.GetTimestamp().Format(time.RFC3339), "2024-12-29T08:11:07Z")
	assert.Equal(t, fmt.Sprintf("%x", exampleCommit.Dirtree), "cffc8909064594608f42782e34de049b1f060251f60a0fbcc08cda3d8e24396a")
	assert.Equal(t, fmt.Sprintf("%x", exampleCommit.DirtreeMeta), "2a28dac42b76c2015ee3c41cc4183bb8b5c790fd21fa5cfa0802c6e11fd0edbe")
}

//go:embed testdata/cffc8909064594608f42782e34de049b1f060251f60a0fbcc08cda3d8e24396a.dirtree
var cffc8909064594608f42782e34de049b1f060251f60a0fbcc08cda3d8e24396adirtree []byte

func TestDeserializeDirtree(t *testing.T) {
	exampleDirtree := Dirtree{}
	assert.Ok(t, gvariant.Unmarshal(cffc8909064594608f42782e34de049b1f060251f60a0fbcc08cda3d8e24396adirtree, &exampleDirtree, binary.LittleEndian))

	assert.Equal(t, fmt.Sprintf("files=%d directories=%d", len(exampleDirtree.Files), len(exampleDirtree.Directories)),
		"files=4 directories=19")

	// name looks like a dir but is probably a symlink
	assert.Equal(t, exampleDirtree.Files[0].Name, "bin")
	assert.Equal(t, fmt.Sprintf("%x", exampleDirtree.Files[0].Checksum), "389846c2702216e1367c8dfb68326a6b93ccf5703c89c93979052a9bf359608e")

	assert.Equal(t, exampleDirtree.Directories[0].Name, "boot")
	assert.Equal(t, fmt.Sprintf("%x", exampleDirtree.Directories[0].DirtreeChecksum), "2aa605f14d5ca3261f5636b956d9122efe569f2ce47d8a6c4408265277a02641")
	assert.Equal(t, fmt.Sprintf("%x", exampleDirtree.Directories[0].DirmetaChecksum), "2a28dac42b76c2015ee3c41cc4183bb8b5c790fd21fa5cfa0802c6e11fd0edbe")
}

//go:embed testdata/28dac42b76c2015ee3c41cc4183bb8b5c790fd21fa5cfa0802c6e11fd0edbe.dirmeta
var exampleDirmetaSerialized []byte

func TestDeserializeDirmeta(t *testing.T) {
	// copy(exampleDirmetaSerialized, []byte{0xe8, 0x03, 0x00, 0x00})
	exampleDirmeta := dirmetaStruct{}
	assert.Ok(t, gvariant.Unmarshal(exampleDirmetaSerialized, &exampleDirmeta, binary.BigEndian))

	assert.Equal(t, exampleDirmeta.UID, 1000)
	assert.Equal(t, exampleDirmeta.GID, 1000)
	assert.Equal(t, exampleDirmeta.Unknown, 16893)
}
