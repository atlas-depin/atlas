package types

import "cosmossdk.io/collections"

type ProviderStatus int32

const (
	ProviderStatusActive ProviderStatus = iota
	ProviderStatusJailed
	ProviderStatusSlashed
)

// Challenge tracks file storage proofs
type Challenge struct {
	FileId        string `json:"file_id"`
	Provider      string `json:"provider_addr"` // [TODO]: use sdk.AccAddress
	BinRoot       []byte `json:"bin_root"`
	ChallengeSeed string `json:"challenge_seed"`
	BlockHeight   int64  `json:"block_height"` // Used as randomness
	ResponseDue   int64  `json:"response_due"`
}

type ChallengeStatus int32

const (
	ChallengeStatusPending ChallengeStatus = iota
	ChallengeStatusResponded
	ChallengeStatusPassed
	ChallengeStatusFailed
	ChallengeStatusExpired
)

var MaxTripleSentinel = collections.Join3(
	string([]byte{0xff, 0xff, 0xff, 0xff}), // ~4×255
	string([]byte{0xff, 0xff, 0xff, 0xff}),
	string([]byte{0xff, 0xff, 0xff, 0xff}),
)

// // Params defines the module parameters
// type Params struct {
// 	MinProviderDeposit sdk.Coin       `json:"min_provider_deposit"`
// 	ChallengeDuration  int64          `json:"challenge_duration"` // blocks
// 	SlashPercentage    math.LegacyDec `json:"slash_percentage"`
// 	AuditPercentage    math.LegacyDec `json:"audit_percentage"` // % of files to audit
// }
