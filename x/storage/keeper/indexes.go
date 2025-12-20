package keeper

import (
	"context"
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

type fileIndexes struct {
	ByOwner        collections.Index[string, types.File]
	BySubscription collections.Index[string, types.File]
}

func NewFileIndexes(sb *collections.SchemaBuilder) fileIndexes {
	return fileIndexes{
		ByOwner: indexes.NewMulti[string, string, types.File](
			sb, collections.NewPrefix(40), "by_owner",
			collections.StringKey, collections.StringKey,
			func(pk string, f types.File) (string, error) {
				return f.Creator, nil
			},
		),
		BySubscription: indexes.NewMulti[string, string, types.File](
			sb, collections.NewPrefix(41), "by_subscription",
			collections.StringKey, collections.StringKey,
			func(pk string, f types.File) (string, error) {
				return f.Subscription, nil
			},
		),
	}
}

func (fi fileIndexes) IndexesList() []collections.Index[string, types.File] {
	return []collections.Index[string, types.File]{
		fi.ByOwner,
		fi.BySubscription,
	}
}

type fileAddressIndex struct{}

func (fa fileAddressIndex) UnorderedKeys() bool { return false }
func (fa fileAddressIndex) Index(_ context.Context, pk string, file types.File) (string, error) {
	return file.Creator, nil
}
func (fa fileAddressIndex) Reference(_ context.Context, pk string, file types.File) (string, error) {
	return pk, nil
}

type fileSubscriptionIndex struct{}

func (fs fileSubscriptionIndex) UnorderedKeys() bool { return false }
func (fs fileSubscriptionIndex) Index(_ context.Context, pk string, file types.File) (string, error) {
	return file.Subscription, nil
}
func (fs fileSubscriptionIndex) Reference(_ context.Context, pk string, file types.File) (string, error) {
	return pk, nil
}

// type fileProviderIndex struct{}

// func (fp fileProviderIndex) UnorderedKeys() bool { return false }
// func (fp fileProviderIndex) Index(_ context.Context, pk string, file types.File) ([]string, error) {
// 	if len(file.Providers) == 0 {
// 		return nil, collections.ErrNotFound
// 	}
// 	return file.Providers, nil
// }
// func (fp fileProviderIndex) Reference(_ context.Context, pk string, file types.File) (string, error) {
// 	return pk, nil
// }
