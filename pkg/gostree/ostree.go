// Go-native OSTree library
package gostree

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joonas-fi/joonas-sys/pkg/gvariant"
	"github.com/samber/lo"
)

// https://ostreedev.github.io/ostree/repo/#core-object-types-and-data-model

type CommitWithID struct {
	ID string
	Commit
}

func Open(at string) *repoReader {
	return &repoReader{at}
}

type repoReader struct {
	path string
}

// tries to first resolve ref from local heads, then from remotes. (this is how `$ ostree log <ref>` operates)
func (r *repoReader) ResolveRef(ref string, FIXMEremoteNames []string) (string, error) {
	withErr := func(err error) (string, error) { return "", fmt.Errorf("ResolveRef: %w", err) }

	refTypesToTry := append([]string{"refs/heads"}, lo.Map(FIXMEremoteNames, func(remoteName string, _ int) string { return "refs/remotes/" + remoteName })...)

	for _, refTypeToTry := range refTypesToTry {
		rawRef, err := os.ReadFile(filepath.Join(r.path, refTypeToTry, ref))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			} else { // actually unexpected error
				return withErr(err)
			}
		}

		return strings.TrimSuffix(string(rawRef), "\n"), nil
	}

	return withErr(fmt.Errorf("ref '%s' not found", ref))
}

func (r *repoReader) ReadCommit(id string) (*Commit, error) {
	commitBytes, err := os.ReadFile(r.objectPath(id) + ".commit")
	if err != nil {
		return nil, err
	}

	// to debug from official ostree use something like:
	// $ ostree show 6807cc4e843e516c36e60ed7d849848b7150160c964d693c802d42d0511cf204 --raw

	commit := &Commit{}
	return commit, gvariant.Unmarshal(commitBytes, commit, binary.BigEndian)
}

func (r *repoReader) ReadDirtree(id string) (*Dirtree, error) {
	dirtreeBytes, err := os.ReadFile(r.objectPath(id) + ".dirtree")
	if err != nil {
		return nil, err
	}

	dirtree := &Dirtree{}
	// return dirtree, gvariant.Unmarshal(dirtreeBytes, dirtree, binary.BigEndian)
	panic("endianness partly wrong here")
	return dirtree, gvariant.Unmarshal(dirtreeBytes, dirtree, binary.LittleEndian)
}

// powers things like `$ ostree log`
func (r *repoReader) ReadParentCommits(id string) ([]CommitWithID, error) {
	withErr := func(err error) ([]CommitWithID, error) { return nil, fmt.Errorf("ReadParentCommits: %w", err) }

	current, err := r.ReadCommit(id)
	if err != nil {
		return withErr(err)
	}

	all := []CommitWithID{CommitWithID{ID: id, Commit: *current}}

	for {
		parentID := fmt.Sprintf("%x", current.ParentCommit)
		parent, err := r.ReadCommit(parentID)
		if err != nil {
			slog.Warn("assuming << History beyond this commit not fetched >> due to", "err", err)

			// assume reached "root" commit already.
			// TODO: structural way to report same as ostree ~"did not fetch all of history"
			return all, nil
		}

		all = append(all, CommitWithID{ID: parentID, Commit: *parent})

		current = parent
	}
}

// NOTE: caller must add content-type specific suffix like ".commit" | ".dirtree" | ...
func (r *repoReader) objectPath(checksum string) string {
	// 6807cc4e843e516c36e60ed7d849848b7150160c964d693c802d42d0511cf204 =>
	// <root>/objects/68/07cc4e843e516c36e60ed7d849848b7150160c964d693c802d42d0511cf204
	return filepath.Join(r.path, "objects", checksum[0:2], checksum[2:])
}
