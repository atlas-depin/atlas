package types

// storage module event types
const (
	EventTypeBuyStorage = "buy_storage"

	EventTypeNewFile    = "new_file"
	EventTypeDeleteFile = "delete_file"

	EventTypeNewProvider = "new_provider"

	EventTypeNewChallengeRound    = "challenge_round_start"
	EventTypeNewChallengeWindow   = "challenge_window_start"
	EventTypeMissedProofsDetected = "proofs_missed"

	AttributeValueCategory = ModuleName

	AttributeKeyBuyer       = "buyer" // buy storage
	AttributeKeyReceiver    = "recipient"
	AttributeKeyBytesBought = "bytes_bought"
	AttributeKeyTimeBought  = "days_bought"

	AttributeKeySigner    = "signer" // sign storage deal
	AttributeKeyFile      = "file"
	AttributeKeyProvider  = "provider"
	AttributeKeyProviders = "providers"
	AttributeKeyPayOnce   = "pay_once"
	AttributeKeyStart     = "start"
)
