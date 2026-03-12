package keeper

import (
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

type fileIndexes struct {
	ByFileId *indexes.Unique[string, collections.Triple[string, string, string], types.File]
	ByStatus *indexes.Multi[string, collections.Triple[string, string, string], types.File]
}

func NewFileIndexes(sb *collections.SchemaBuilder) fileIndexes {
	return fileIndexes{
		ByFileId: indexes.NewUnique(
			sb, collections.NewPrefix(40), "by_fid",
			collections.StringKey, collections.TripleKeyCodec(collections.StringKey, collections.StringKey, collections.StringKey),
			func(pk collections.Triple[string, string, string], f types.File) (string, error) {
				return f.Fid, nil
			},
		),
		ByStatus: indexes.NewMulti(
			sb, collections.NewPrefix(41), "by_status",
			collections.StringKey, collections.TripleKeyCodec(collections.StringKey, collections.StringKey, collections.StringKey),
			func(pk collections.Triple[string, string, string], f types.File) (string, error) {
				return f.Status, nil
			},
		),
	}
}

func (fi fileIndexes) IndexesList() []collections.Index[collections.Triple[string, string, string], types.File] {
	return []collections.Index[collections.Triple[string, string, string], types.File]{
		fi.ByFileId,
		fi.ByStatus,
	}
}

// type challengeIndex struct {
// 	ByProvider *indexes.Multi[string, collections.Pair[string, string], types.StorageChallenge]
// }

// func NewChallengeIndexes(sb *collections.SchemaBuilder) challengeIndex {
// 	return challengeIndex{
// 		ByProvider: indexes.NewMulti[string, collections.Pair[string, string], types.StorageChallenge](
// 			sb, collections.NewPrefix(80), "by_provider",
// 			collections.StringKey, collections.PairKeyCodec(collections.StringKey, collections.StringKey),
// 			func(pk collections.Pair[string, string], f types.StorageChallenge) (string, error) {
// 				return f.Provider, nil
// 			},
// 		),
// 	}
// }

// func (fi challengeIndex) IndexesList() []collections.Index[collections.Pair[string, string], types.StorageChallenge] {
// 	return []collections.Index[collections.Pair[string, string], types.StorageChallenge]{
// 		fi.ByProvider,
// 	}
// }
// func (fa challengeIndex) UnorderedKeys() bool { return false }
// func (fa challengeIndex) Index(_ context.Context, pk string, file types.File) (string, error) {
// 	return file.Fid, nil
// }
// func (fa challengeIndex) Reference(_ context.Context, pk string, file types.File) (string, error) {
// 	return pk, nil
// }
