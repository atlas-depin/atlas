package keeper

import (
	"context"
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

type fileIndexes struct {
	ByFileId *indexes.Unique[string, collections.Triple[string, string, string], types.File]
}

func NewFileIndexes(sb *collections.SchemaBuilder) fileIndexes {
	return fileIndexes{
		ByFileId: indexes.NewUnique[string, collections.Triple[string, string, string], types.File](
			sb, collections.NewPrefix(40), "by_fid",
			collections.StringKey, collections.TripleKeyCodec(collections.StringKey, collections.StringKey, collections.StringKey),
			func(pk collections.Triple[string, string, string], f types.File) (string, error) {
				return f.FileId, nil
			},
		),
	}
}

func (fi fileIndexes) IndexesList() []collections.Index[collections.Triple[string, string, string], types.File] {
	return []collections.Index[collections.Triple[string, string, string], types.File]{
		fi.ByFileId,
	}
}

type fileIdIndex struct{}

func (fa fileIdIndex) UnorderedKeys() bool { return false }
func (fa fileIdIndex) Index(_ context.Context, pk string, file types.File) (string, error) {
	return file.FileId, nil
}
func (fa fileIdIndex) Reference(_ context.Context, pk string, file types.File) (string, error) {
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
