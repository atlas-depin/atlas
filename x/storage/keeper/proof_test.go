package keeper_test

import (
	"encoding/hex"
	"time"

	"cosmossdk.io/math"

	"atlas/x/storage/types"

	merkletree "github.com/ahm23/go-merkletree-xxh"
	"github.com/zeebo/blake3"
)

// buildProof constructs a valid merkle proof for the leaf at index, matching the
// keeper's Prove() verification path: the tree is built over the blake3 hashes
// of each chunk using the XXH128 tree config.
func buildProof(chunks [][]byte, index int) (root []byte, p types.Proof) {
	leaves := make([][]byte, 0, len(chunks))
	for _, c := range chunks {
		h := blake3.Sum256(c)
		leaves = append(leaves, h[:])
	}
	tree, err := merkletree.New(&merkletree.Config{XXH128: true}, leaves)
	if err != nil {
		panic(err)
	}
	proof, err := tree.Proof(index)
	if err != nil {
		panic(err)
	}
	return tree.Root, types.Proof{
		Index:  uint64(index),
		Path:   proof.PathBits,
		Hashes: proof.Siblings,
		Data:   chunks[index],
	}
}

func testChunks() [][]byte {
	return [][]byte{
		[]byte("chunk-0-data"),
		[]byte("chunk-1-data"),
		[]byte("chunk-2-data"),
		[]byte("chunk-3-data"),
	}
}

func (suite *KeeperTestSuite) TestProve_Valid() {
	chunks := testChunks()
	root, p := buildProof(chunks, 1)

	suite.Require().True(suite.keeper.Prove(suite.ctx, root, p.Index, p))
}

func (suite *KeeperTestSuite) TestProve_TamperedData() {
	chunks := testChunks()
	root, p := buildProof(chunks, 1)

	p.Data = []byte("tampered")
	suite.Require().False(suite.keeper.Prove(suite.ctx, root, p.Index, p))
}

func (suite *KeeperTestSuite) TestProve_IndexMismatch() {
	chunks := testChunks()
	root, p := buildProof(chunks, 1)

	suite.Require().False(suite.keeper.Prove(suite.ctx, root, p.Index+1, p))
}

func (suite *KeeperTestSuite) TestProve_EmptyRoot() {
	chunks := testChunks()
	_, p := buildProof(chunks, 0)

	suite.Require().False(suite.keeper.Prove(suite.ctx, []byte{}, p.Index, p))
}

func (suite *KeeperTestSuite) TestAttemptAddProvider_Success_BecomesActive() {
	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	chunks := testChunks()
	root, p := buildProof(chunks, 0)

	// Seed the file directly (no count increment — the count only moves on first proof).
	file := types.File{
		Fid:      "claim-active",
		Merkle:   hex.EncodeToString(root),
		Creator:  owner,
		Status:   "STAGED",
		FileSize: 1024,
		Replicas: 1,
	}
	suite.Require().NoError(suite.keeper.SetNewFile(suite.ctx, file))

	err := suite.keeper.AttemptAddProvider(suite.ctx, file, provider, p)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Require().Equal("ACTIVE", stored.Status)
	suite.Require().Equal([]string{provider}, stored.Providers)

	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), count)
}

func (suite *KeeperTestSuite) TestAttemptAddProvider_Success_BecomesStray() {
	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	chunks := testChunks()
	root, p := buildProof(chunks, 0)

	// Replicas=3, only one provider -> STRAY
	file := suite.createTestFile("claim-stray", owner, 1024, 3, nil)
	file.Status = "STAGED"
	file.Merkle = hex.EncodeToString(root)
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	err := suite.keeper.AttemptAddProvider(suite.ctx, file, provider, p)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Require().Equal("STRAY", stored.Status)
	suite.Require().Equal([]string{provider}, stored.Providers)
}

func (suite *KeeperTestSuite) TestAttemptAddProvider_InvalidProof() {
	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	file := suite.createTestFile("claim-invalid", owner, 1024, 1, nil)
	file.Status = "STAGED"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	badProof := types.Proof{Index: 0, Data: []byte("wrong"), Hashes: [][]byte{[]byte("garbage")}}

	err := suite.keeper.AttemptAddProvider(suite.ctx, file, provider, badProof)
	suite.Require().ErrorIs(err, types.ErrProofInvalid)

	// File untouched.
	stored, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Require().Empty(stored.Providers)
}

func (suite *KeeperTestSuite) TestAttemptChallenge_MissingChallenge() {
	err := suite.keeper.AttemptChallenge(suite.ctx, suite.CreateTestAddress(0), types.Proof{ChallengeId: "missing"})
	suite.Require().NoError(err, "missing challenges are handled silently")
}

func (suite *KeeperTestSuite) TestAttemptChallenge_MissingFile() {
	provider := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "ch-missing-file",
		FileId:      "no-such-file",
		Provider:    provider,
	}))

	err := suite.keeper.AttemptChallenge(suite.ctx, provider, types.Proof{ChallengeId: "ch-missing-file"})
	suite.Require().NoError(err)

	// Challenge is cleaned up even when the file is gone.
	_, err = suite.keeper.GetChallenge(suite.ctx, provider, "ch-missing-file")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestAttemptChallenge_BadProvider() {
	owner := suite.CreateTestAddress(0)
	fileProvider := suite.CreateTestAddress(1)
	attacker := suite.CreateTestAddress(2)

	file := suite.createTestFile("ch-bad-provider", owner, 1024, 1, []string{fileProvider})
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "ch-bad",
		FileId:      file.Fid,
		Provider:    attacker,
	}))

	err := suite.keeper.AttemptChallenge(suite.ctx, attacker, types.Proof{ChallengeId: "ch-bad"})
	suite.Require().ErrorIs(err, types.ErrProofFileBadProvider)
}

func (suite *KeeperTestSuite) TestAttemptChallenge_InvalidProof() {
	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	file := suite.createTestFile("ch-invalid-proof", owner, 1024, 1, []string{provider})
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "ch-invalid",
		FileId:      file.Fid,
		Provider:    provider,
		ChunkIndex:  2,
	}))

	// Proof for a different chunk than the challenge demands.
	badProof := types.Proof{ChallengeId: "ch-invalid", Index: 0, Data: []byte("wrong")}

	err := suite.keeper.AttemptChallenge(suite.ctx, provider, badProof)
	suite.Require().ErrorIs(err, types.ErrProofInvalid)
}

func (suite *KeeperTestSuite) TestAttemptChallenge_Success() {
	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	chunks := testChunks()
	root, p := buildProof(chunks, 1)

	file := suite.createTestFile("ch-success", owner, 1024, 1, []string{provider})
	file.Merkle = hex.EncodeToString(root)
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "ch-ok",
		FileId:      file.Fid,
		Provider:    provider,
		ChunkIndex:  1,
	}))

	p.ChallengeId = "ch-ok"
	err := suite.keeper.AttemptChallenge(suite.ctx, provider, p)
	suite.Require().NoError(err)

	// Challenge consumed.
	_, err = suite.keeper.GetChallenge(suite.ctx, provider, "ch-ok")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestAttemptChallenge_ExpiredSubscription() {
	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	// Expired subscription.
	end := suite.ctx.BlockTime().Add(-1 * time.Hour)
	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime().Add(-24 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	file := suite.createTestFileWithSub("ch-expired", owner, subId, 1024, 1, []string{provider})
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "ch-expired",
		FileId:      file.Fid,
		Provider:    provider,
	}))

	err = suite.keeper.AttemptChallenge(suite.ctx, provider, types.Proof{ChallengeId: "ch-expired"})
	suite.Require().NoError(err)

	// File expired and challenge removed.
	_, err = suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().ErrorIs(err, types.ErrFileNotFound)
	_, err = suite.keeper.GetChallenge(suite.ctx, provider, "ch-expired")
	suite.Require().Error(err)
}
