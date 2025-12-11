package keeper

// import (
// 	"encoding/binary"
// 	"hash"

// 	"github.com/cespare/xxhash"
// 	"github.com/wealdtech/go-merkletree/v2"
// 	"github.com/zeebo/blake3"
// )

// func HashChunk(chunk []byte) []byte {
// 	// Using SHA256 as placeholder - replace with actual xxhash64
// 	// In production: use "github.com/cespare/xxhash"
// 	hash := xxhash.Sum64(chunk)
// 	return hash[:]
// }

// // VerifyMerkleProof verifies a chunk against Merkle root
// func VerifyMerkleProof(root []byte, proof *MerkleProof) (bool, error) {
// 	// Hash the chunk first
// 	chunkHash := HashChunk(proof.Chunk)

// 	// Create proof struct for the library
// 	merkleProof := &merkletree.Proof{
// 		Hash:   chunkHash,
// 		Index:  proof.Index,
// 		Hashes: proof.Siblings,
// 	}

// 	// Verify using Blake3
// 	return merkletree.VerifyProof(chunkHash, merkleProof, root, blake3.New())
// }

// /// *** VRF Methods *** ///

// type VerifiableRandomness struct {
// 	blockHash []byte
// 	hasher    hash.Hash
// 	counter   uint64
// }

// func NewVerifiableRandomness(blockHash []byte) *VerifiableRandomness {
// 	hasher := blake3.New()
// 	return &VerifiableRandomness{
// 		blockHash: blockHash,
// 		hasher:    hasher,
// 		counter:   0,
// 	}
// }

// // NextRandom generates next random number in sequence
// func (vr *VerifiableRandomness) NextRandom(max uint64) uint64 {
// 	// Mix counter with block hash
// 	vr.hasher.Reset()
// 	vr.hasher.Write(vr.blockHash)

// 	counterBytes := make([]byte, 8)
// 	binary.BigEndian.PutUint64(counterBytes, vr.counter)
// 	vr.hasher.Write(counterBytes)

// 	hash := vr.hasher.Sum(nil)

// 	// Convert first 8 bytes to uint64
// 	result := binary.BigEndian.Uint64(hash[:8])

// 	vr.counter++
// 	return result % max
// }
