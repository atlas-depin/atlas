package types

// storage module event types
const (
	EventTypeBuyStorage = "buy_storage"

	EventTypeNewFile    = "new_file"
	EventTypeDeleteFile = "delete_file"

	EventTypeNewProvider        = "new_provider"
	EventTypeExpandSubscription = "expand_subscription"

	EventTypeNewChallengeRound    = "challenge_round_start"
	EventTypeNewChallengeWindow   = "challenge_window_start"
	EventTypeMissedProofsDetected = "proofs_missed"
	EventTypeChallengeAttempt     = "challenge_attempt"

	AttributeValueCategory = ModuleName

	AttributeKeyBuyer          = "buyer" // buy storage
	AttributeKeyReceiver       = "recipient"
	AttributeKeyBytesBought    = "bytes_bought"
	AttributeKeyTimeBought     = "days_bought"
	AttributeKeySubscriptionId = "subscription_id"
	AttributeKeyAmountPaid     = "amount_paid"
	AttributeKeyExpiration     = "expiration"
	AttributeKeySpaceAvailable = "space_available"

	AttributeKeySigner      = "signer" // sign storage deal
	AttributeKeyFile        = "file_id"
	AttributeKeyProvider    = "provider"
	AttributeKeyProviders   = "providers"
	AttributeKeyPayOnce     = "pay_once"
	AttributeKeyStart       = "start"
	AttributeKeyChallengeId = "challenge_id"
	AttributeKeyProofChunk  = "chunk"
)
