#!/usr/bin/env bash
set -euo pipefail

# gov-gov.sh — Change x/gov parameters through a governance proposal.
#
# Usage:
#   ./gov-gov.sh [flags]
#
# Flags (all optional — omitted fields keep current on-chain values):
#   --voting-period        "15s"         (e.g. "300s", "1h")
#   --max-deposit-period   "15s"         (e.g. "86400s", "2d")
#   --min-deposit          "100000000"   (amount in uatl)
#   --expedited-voting-period    "10s"
#   --expedited-min-deposit      "101000000"
#   --min-initial-deposit-ratio  "0.000000000000000000"
#   --quorum              "0.334000000000000000"
#   --threshold           "0.500000000000000000"
#   --veto-threshold      "0.334000000000000000"
#   --title       "My Proposal"
#   --description "Detailed rationale"
#   --deposit     "100000000uatl"        (proposal deposit, default: 100 ATL)
#   --from        <key-name>            (default: validator1)
#   --home        <node-home>           (default: ~/.atlas-devnet/node1)
#   --chain-id    <chain-id>            (default: atlas-1)
#   --node        <rpc-addr>            (default: http://localhost:26667)
#
# Examples:
#   ./gov-gov.sh --voting-period "30s" --max-deposit-period "30s"
#   ./gov-gov.sh --min-deposit "50000000" --max-deposit-period "60s" --from bob

BIN="${BIN:-atlasd}"

# ── defaults ────────────────────────────────────────────────────────────────
TITLE=""
DESC=""
DEPOSIT=""
FROM="validator1"
HOME_DIR="${HOME}/.atlas-devnet/node1"
KEYRING_DIR="${HOME}/.atlas-devnet/keys"
KEYRING_BACKEND="test"
CHAIN_ID="atlas-1"
NODE="http://localhost:26667"

# param overrides (empty = keep current)
VOTING_PERIOD=""
MAX_DEPOSIT_PERIOD=""
MIN_DEPOSIT_AMOUNT=""
EXPEDITED_VOTING=""
EXPEDITED_MIN_DEPOSIT=""
MIN_INITIAL_RATIO=""
QUORUM=""
THRESHOLD=""
VETO_THRESHOLD=""
BURN_VOTE_QUORUM=""
BURN_PROPOSAL_DEPOSIT=""
BURN_VOTE_VETO=""

# ── parse flags ─────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --voting-period)             VOTING_PERIOD="$2";         shift 2 ;;
    --max-deposit-period)        MAX_DEPOSIT_PERIOD="$2";    shift 2 ;;
    --min-deposit)               MIN_DEPOSIT_AMOUNT="$2";    shift 2 ;;
    --expedited-voting-period)   EXPEDITED_VOTING="$2";      shift 2 ;;
    --expedited-min-deposit)     EXPEDITED_MIN_DEPOSIT="$2"; shift 2 ;;
    --min-initial-deposit-ratio) MIN_INITIAL_RATIO="$2";     shift 2 ;;
    --quorum)                    QUORUM="$2";                shift 2 ;;
    --threshold)                 THRESHOLD="$2";             shift 2 ;;
    --veto-threshold)            VETO_THRESHOLD="$2";        shift 2 ;;
    --burn-vote-quorum)          BURN_VOTE_QUORUM="$2";      shift 2 ;;
    --burn-proposal-deposit)     BURN_PROPOSAL_DEPOSIT="$2"; shift 2 ;;
    --burn-vote-veto)            BURN_VOTE_VETO="$2";        shift 2 ;;
    --title)                     TITLE="$2";                 shift 2 ;;
    --description)               DESC="$2";                  shift 2 ;;
    --deposit)                   DEPOSIT="$2";               shift 2 ;;
    --from)                      FROM="$2";                  shift 2 ;;
    --home)                      HOME_DIR="$2";              shift 2 ;;
    --keyring-dir)               KEYRING_DIR="$2";           shift 2 ;;
    --keyring-backend)           KEYRING_BACKEND="$2";       shift 2 ;;
    --chain-id)                  CHAIN_ID="$2";              shift 2 ;;
    --node)                      NODE="$2";                  shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# ── gov module address ──────────────────────────────────────────────────────
# Default for the atlas-1 devnet (derived from module name "gov" with prefix "atl").
# Override with GOV_ADDR env var.
DEFAULT_GOV_ADDR="atl10d07y265gmmuvt4z0w9aw880jnsr700jltv948"
GOV_ADDR="${GOV_ADDR:-$DEFAULT_GOV_ADDR}"
echo "Using gov module address: $GOV_ADDR"

# ── fetch current gov params ────────────────────────────────────────────────
CUR=$($BIN q gov params \
  --node "$NODE" --chain-id "$CHAIN_ID" \
  --keyring-dir "$KEYRING_DIR" --keyring-backend "$KEYRING_BACKEND" \
  -o json 2>/dev/null || echo "{}")

pick_val() {
  local key="$1" override="$2"
  if [[ -n "$override" ]]; then
    echo "$override"
  else
    echo "$CUR" | jq -r ".${key} // empty" 2>/dev/null || echo ""
  fi
}

VOTING_PERIOD=$(   pick_val "voting_period"              "$VOTING_PERIOD")
MAX_DEPOSIT_PERIOD=$(pick_val "max_deposit_period"       "$MAX_DEPOSIT_PERIOD")
EXPEDITED_VOTING=$(pick_val "expedited_voting_period"    "$EXPEDITED_VOTING")
MIN_INITIAL_RATIO=$(pick_val "min_initial_deposit_ratio" "$MIN_INITIAL_RATIO")
QUORUM=$(          pick_val "quorum"                     "$QUORUM")
THRESHOLD=$(       pick_val "threshold"                  "$THRESHOLD")
VETO_THRESHOLD=$(  pick_val "veto_threshold"             "$VETO_THRESHOLD")

# booleans — need careful handling since jq outputs "true"/"false" string
BURN_VOTE_QUORUM_NEW=$(pick_val "burn_vote_quorum" "$BURN_VOTE_QUORUM")
[[ "$BURN_VOTE_QUORUM_NEW" == "" ]] && BURN_VOTE_QUORUM_NEW="false"

BURN_PROPOSAL_DEPOSIT_NEW=$(pick_val "burn_proposal_deposit" "$BURN_PROPOSAL_DEPOSIT")
[[ "$BURN_PROPOSAL_DEPOSIT_NEW" == "" ]] && BURN_PROPOSAL_DEPOSIT_NEW="false"

BURN_VOTE_VETO_NEW=$(pick_val "burn_vote_veto" "$BURN_VOTE_VETO")
[[ "$BURN_VOTE_VETO_NEW" == "" ]] && BURN_VOTE_VETO_NEW="false"

# min_deposit — handle coin list
if [[ -n "$MIN_DEPOSIT_AMOUNT" ]]; then
  MIN_DEPOSIT_JSON=$(cat <<COINEOF
[{"denom": "uatl", "amount": "$MIN_DEPOSIT_AMOUNT"}]
COINEOF
)
else
  MIN_DEPOSIT_JSON=$(echo "$CUR" | jq '.min_deposit // [{"denom":"uatl","amount":"100000000"}]' 2>/dev/null)
fi

# expedited min deposit
if [[ -n "$EXPEDITED_MIN_DEPOSIT" ]]; then
  EXP_MIN_DEPOSIT_JSON=$(cat <<COINEOF
[{"denom": "uatl", "amount": "$EXPEDITED_MIN_DEPOSIT"}]
COINEOF
)
else
  EXP_MIN_DEPOSIT_JSON=$(echo "$CUR" | jq '.expedited_min_deposit // [{"denom":"uatl","amount":"101000000"}]' 2>/dev/null)
fi

# ── build params JSON ──────────────────────────────────────────────────────
PARAMS_JSON=$(cat <<EOF
{
  "min_deposit":              $MIN_DEPOSIT_JSON,
  "max_deposit_period":       "$MAX_DEPOSIT_PERIOD",
  "voting_period":            "$VOTING_PERIOD",
  "expedited_voting_period":  "$EXPEDITED_VOTING",
  "expedited_min_deposit":    $EXP_MIN_DEPOSIT_JSON,
  "expedited_threshold":      "0.667000000000000000",
  "min_initial_deposit_ratio": "$MIN_INITIAL_RATIO",
  "quorum":                   "$QUORUM",
  "threshold":                "$THRESHOLD",
  "veto_threshold":           "$VETO_THRESHOLD",
  "burn_vote_quorum":         $BURN_VOTE_QUORUM_NEW,
  "burn_proposal_deposit":    $BURN_PROPOSAL_DEPOSIT_NEW,
  "burn_vote_veto":           $BURN_VOTE_VETO_NEW
}
EOF
)

# ── build proposal message ──────────────────────────────────────────────────
MSG=$(cat <<EOF
{
  "@type": "/cosmos.gov.v1.MsgUpdateParams",
  "authority": "$GOV_ADDR",
  "params": $PARAMS_JSON
}
EOF
)

# ── defaults for title / description / deposit ──────────────────────────────
if [[ -z "$TITLE" ]]; then
  TITLE="Update governance module parameters"
fi
if [[ -z "$DESC" ]]; then
  DESC=$(echo "Proposed x/gov parameter changes: $PARAMS_JSON" | jq -R -c .)
fi
if [[ -z "$DEPOSIT" ]]; then
  DEPOSIT="100000000uatl"
fi

# ── write proposal JSON ─────────────────────────────────────────────────────
PROPOSAL_FILE=$(mktemp /tmp/gov-gov-proposal-XXXXXX.json)
trap 'rm -f "$PROPOSAL_FILE"' EXIT

cat > "$PROPOSAL_FILE" <<EOF
{
  "title": "$TITLE",
  "description": $(echo "$DESC" | jq -R -c .),
  "deposit": "$DEPOSIT",
  "messages": [$MSG]
}
EOF

echo ""
echo "Proposal JSON:"
jq . "$PROPOSAL_FILE"

echo ""
echo "══════════════════════════════════════════════════════"
echo "Submitting governance parameter-change proposal..."
echo "══════════════════════════════════════════════════════"

$BIN tx gov submit-proposal "$PROPOSAL_FILE" \
  --from "$FROM" \
  --home "$HOME_DIR" \
  --keyring-dir "$KEYRING_DIR" \
  --keyring-backend "$KEYRING_BACKEND" \
  --chain-id "$CHAIN_ID" \
  --node "$NODE" \
  --fees 10000uatl \
  --gas auto \
  --gas-adjustment 1.5 \
  --yes

echo ""
echo "To vote:"
echo "  $BIN tx gov vote 1 yes --from $FROM --home $HOME_DIR --keyring-dir $KEYRING_DIR --keyring-backend $KEYRING_BACKEND --chain-id $CHAIN_ID --node $NODE --fees 10000uatl --yes"
echo ""
echo "To check proposal status:"
echo "  $BIN q gov proposal 1 --node $NODE --chain-id $CHAIN_ID --home $HOME_DIR"
