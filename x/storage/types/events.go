package types

// storage module event types
const (
	EventTypeBuyStorage = "buy_storage"

	EventTypeSignContract         = "post_file"
	EventTypeCancelContract       = "delete_file"
	EventTypeNewChallengeRound    = "challenge_round_start"
	EventTypeNewChallengeWindow   = "challenge_window_start"
	EventTypeMissedProofsDetected = "proofs_missed"

	AttributeValueCategory = ModuleName

	AttributeKeyBuyer       = "buyer" // buy storage
	AttributeKeyReceiver    = "recipient"
	AttributeKeyBytesBought = "bytes_bought"
	AttributeKeyTimeBought  = "days_bought"

	AttributeKeySigner   = "signer" // sign storage deal
	AttributeKeyContract = "file"
	AttributeKeyPayOnce  = "pay_once"
	AttributeKeyStart    = "start"
)
