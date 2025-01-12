package ostree

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/function61/gokit/os/osutil"
	"github.com/joonas-fi/joonas-sys/pkg/common"
	"github.com/joonas-fi/joonas-sys/pkg/filelocations"
	"github.com/joonas-fi/joonas-sys/pkg/gostree"
	"github.com/samber/lo"
)

func ListVersions(root filelocations.Root) ([]CheckoutWithLabel, error) {
	versionsFromCheckouts, err := listVersionsFromCheckouts(root)
	if err != nil {
		return nil, err
	}

	versionsFromRepo, err := listVersionsFromRepo(root)
	if err != nil {
		return nil, err
	}

	// do not add duplicates by sourcing also from commit log if that is already committed
	versionsFromRepoNotInCheckouts := lo.Filter(versionsFromRepo, func(versionFromRepo CheckoutWithLabel, _ int) bool {
		return !lo.ContainsBy(versionsFromCheckouts, func(checkout CheckoutWithLabel) bool {
			return checkout.CommitShort == versionFromRepo.CommitShort
		})
	})

	allVersions := append([]CheckoutWithLabel{}, versionsFromCheckouts...)
	allVersions = append(allVersions, versionsFromRepoNotInCheckouts...)

	sort.Slice(allVersions, func(i, j int) bool {
		// newest to oldest
		return allVersions[i].Timestamp.After(allVersions[j].Timestamp)
	})

	return allVersions, nil
}

func listVersionsFromRepo(root filelocations.Root) ([]CheckoutWithLabel, error) {
	withErr := func(err error) ([]CheckoutWithLabel, error) { return nil, fmt.Errorf("listVersionsFromRepo: %w", err) }

	repo := gostree.Open(root.App(common.AppOSRepo))

	commitID, err := repo.ResolveRef(ostreeBranchNameX8664, remoteNameParam())
	if err != nil {
		return withErr(err)
	}
	commitLog, err := repo.ReadParentCommits(commitID)
	if err != nil {
		return withErr(err)
	}

	return lo.Map(commitLog, func(commit gostree.CommitWithID, _ int) CheckoutWithLabel {
		labelComponents := []string{commitShort(commit.ID), commit.GetTimestamp().Format(time.RFC3339), commit.Subject}

		return CheckoutWithLabel{
			CommitShort: commitShort(commit.ID),
			Commit:      commit.ID, // needs to be filled for commits
			Label:       checkedOutIndicator(false) + strings.Join(labelComponents, " - "),
			Timestamp:   commit.GetTimestamp(),
		}
	}), nil
}

func EnsureCheckedOut(ctx context.Context, cwl CheckoutWithLabel) (string, error) {
	checkout := filelocations.Sysroot.Checkout(cwl.CommitShort)

	exists, err := osutil.Exists(checkout)
	if err != nil {
		return "", err
	}
	if !exists {
		slog.Warn("not checked out; proceeding to checkout", "checkout", checkout)

		if err := checkoutRootFS(ctx, cwl.Commit); err != nil {
			return "", err
		}
	}

	return checkout, nil
}

func checkedOutIndicator(set bool) string {
	if set {
		return "✅"
	} else {
		return "  "
	}
}

func commitShort(id string) string {
	return id[:7] // 7 hexits is unique enough
}
