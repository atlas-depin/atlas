#!/usr/bin/env bash
set -euo pipefail

# governance-proposal.sh — Submit a storage module parameter change via governance.
#
# Usage:
#   ./governance-proposal.sh [flags]
#
# Flags (all optional — omitted fields keep current on-chain values):
#   --price-per-gb-day <int64>       e.g. 400
#   --provider-deposit <int64>       e.g. 10000
#   --proof-window-blocks <uint64>   e.g. 50
#   --proof-round-blocks <uint64>    e.g. 10
#   --title       "My Proposal"
#   --description "Detailed rationale"
#   --deposit     "100000000uatl"    (default: min deposit)
#   --from        <key-name>         (default: validator1)
#   --home        <node-home>        (default: ~/.atlas-devnet/node1)
#   --keyring-dir <path>             (default: ~/.atlas-devnet/keys)
#   --keyring-backend <backend>      (default: test)
#   --chain-id    <chain-id>         (default: atlas-1)
#   --node        <rpc-addr>         (default: http://localhost:26667)
#
# Examples:
#   ./storage-governance-proposal.sh --proof-round-blocks 20 --from alice \
#     --title "Speed up proof rounds" --description "Reduce round blocks from 10 to 20"
#
#   ./governance-proposal.sh --price-per-gb-day 500 --provider-deposit 20000

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
PRICE_PER_GB=""
PROVIDER_DEPOSIT=""
PROOF_WINDOW=""
PROOF_ROUND=""

# ── parse flags ─────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --price-per-gb-day)    PRICE_PER_GB="$2";    shift 2 ;;
    --provider-deposit)    PROVIDER_DEPOSIT="$2"; shift 2 ;;
    --proof-window-blocks) PROOF_WINDOW="$2";    shift 2 ;;
    --proof-round-blocks)  PROOF_ROUND="$2";     shift 2 ;;
    --title)               TITLE="$2";           shift 2 ;;
    --description)         DESC="$2";            shift 2 ;;
    --deposit)             DEPOSIT="$2";         shift 2 ;;
    --from)                FROM="$2";            shift 2 ;;
    --home)                HOME_DIR="$2";        shift 2 ;;
    --keyring-dir)         KEYRING_DIR="$2";     shift 2 ;;
    --keyring-backend)     KEYRING_BACKEND="$2"; shift 2 ;;
    --chain-id)            CHAIN_ID="$2";        shift 2 ;;
    --node)                NODE="$2";            shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# ── gov module address ──────────────────────────────────────────────────────
# Default for the atlas-1 devnet (derived from module name "gov" with prefix "atl").
# Override with --gov-addr or GOV_ADDR env var.
DEFAULT_GOV_ADDR="atl10d07y265gmmuvt4z0w9aw880jnsr700jltv948"
GOV_ADDR="${GOV_ADDR:-$DEFAULT_GOV_ADDR}"

echo "Using gov module address: $GOV_ADDR"

# ── build params JSON ──────────────────────────────────────────────────────
# Start with current on-chain params so omitted fields keep their values.
CURRENT_PARAMS=$($BIN q storage params \
  --node "$NODE" \
  --chain-id "$CHAIN_ID" \
  --keyring-dir "$KEYRING_DIR" \
  --keyring-backend "$KEYRING_BACKEND" \
  -o json 2>/dev/null | jq -r '.params // .' 2>/dev/null || echo "{}")

# Helper: use override if given, otherwise keep current value.
pick() {
  local key="$1" override="$2"
  if [[ -n "$override" ]]; then
    echo "$override"
  else
    echo "$CURRENT_PARAMS" | jq -r ".${key}" 2>/dev/null || echo "0"
  fi
}

PRICE_PER_GB=$(  pick "price_per_gb_day"     "$PRICE_PER_GB")
PROVIDER_DEPOSIT=$(pick "provider_deposit"   "$PROVIDER_DEPOSIT")
PROOF_WINDOW=$(  pick "proof_window_blocks"  "$PROOF_WINDOW")
PROOF_ROUND=$(   pick "proof_round_blocks"   "$PROOF_ROUND")

PARAMS_JSON=$(cat <<EOF
{
  "price_per_gb_day":     "$PRICE_PER_GB",
  "provider_deposit":    "$PROVIDER_DEPOSIT",
  "proof_window_blocks": "$PROOF_WINDOW",
  "proof_round_blocks":  "$PROOF_ROUND"
}
EOF
)

# ── build proposal message ──────────────────────────────────────────────────
MSG=$(cat <<EOF
{
  "@type": "/atlas.storage.v1.MsgUpdateParams",
  "authority": "$GOV_ADDR",
  "params": $PARAMS_JSON
}
EOF
)

# ── defaults for title / description / deposit ──────────────────────────────
if [[ -z "$TITLE" ]]; then
  TITLE="Update storage module parameters"
fi
if [[ -z "$DESC" ]]; then
  DESC=$(echo "Proposed storage parameter changes: $PARAMS_JSON" | jq -R -c .)
fi
if [[ -z "$DEPOSIT" ]]; then
  DEPOSIT="100000000uatl"
fi

# ── write proposal JSON ─────────────────────────────────────────────────────
PROPOSAL_FILE=$(mktemp /tmp/storage-gov-proposal-XXXXXX.json)
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
echo "Submitting governance proposal..."
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
