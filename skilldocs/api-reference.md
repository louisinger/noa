# Ark REST API Reference

**Default mainnet server:** `https://arkade.computer`

```bash
ARK_SERVER="https://arkade.computer"
# Testnet: ARK_SERVER="https://testnet.arkade.computer"
```

---

## IndexerService Endpoints

### Get VTXOs

```bash
# By script (hex-encoded pk_script)
curl "$ARK_SERVER/v1/indexer/vtxos?scripts=<script_hex>"

# By outpoint
curl "$ARK_SERVER/v1/indexer/vtxos?outpoints=<txid>:<vout>"

# Filter flags (combinable)
curl "$ARK_SERVER/v1/indexer/vtxos?scripts=<hex>&spendable_only=true"
curl "$ARK_SERVER/v1/indexer/vtxos?scripts=<hex>&spent_only=true"
curl "$ARK_SERVER/v1/indexer/vtxos?scripts=<hex>&recoverable_only=true"

# Time range (Unix ms)
curl "$ARK_SERVER/v1/indexer/vtxos?scripts=<hex>&after=1700000000000&before=1710000000000"
```

**IndexerVtxo response fields:**

| Field | Meaning |
|-------|---------|
| `is_preconfirmed` | VTXO exists but round not yet confirmed onchain |
| `is_swept` | Server swept this VTXO after expiry |
| `is_unrolled` | VTXO was unilaterally exited onchain |
| `is_spent` | VTXO was spent (offchain or onchain) |
| `spent_by` | Txid of the spending transaction |
| `expires_at` | Unix timestamp when VTXO expires |
| `commitment_txids` | List of commitment txs this VTXO appeared in |
| `settled_by` | Txid that settled this VTXO onchain |
| `ark_txid` | The Ark virtual transaction ID |

### Get VTXO Chain (Spending History)

```bash
curl "$ARK_SERVER/v1/indexer/vtxo/<txid>/<vout>/chain"
```

**Response:** Chain of transactions with `txid`, `expires_at`, `type`, `spends`.

**IndexerChainedTxType values:**

| Type | Meaning |
|------|---------|
| `COMMITMENT` | Onchain commitment transaction (round) |
| `ARK` | Offchain Ark virtual transaction |
| `TREE` | VTXO tree transaction |
| `CHECKPOINT` | Checkpoint transaction |

### Get Commitment Transaction

```bash
curl "$ARK_SERVER/v1/indexer/commitmentTx/<txid>"
```

**Response fields:** `started_at`, `ended_at`, `batches` (amounts + expiry), `total_input_amount`, `total_output_amount`, `total_input_vtxos`, `total_output_vtxos`.

### Get Forfeit Transactions

```bash
curl "$ARK_SERVER/v1/indexer/commitmentTx/<txid>/forfeitTxs"
```

### Get Connectors Tree

```bash
curl "$ARK_SERVER/v1/indexer/commitmentTx/<txid>/connectors"
```

### Get VTXO Tree

```bash
# Full tree
curl "$ARK_SERVER/v1/indexer/batch/<txid>/<vout>/tree"

# Tree leaves (VTXO outpoints)
curl "$ARK_SERVER/v1/indexer/batch/<txid>/<vout>/tree/leaves"
```

### Get Virtual Transactions

```bash
# One or multiple txids (comma-separated)
curl "$ARK_SERVER/v1/indexer/virtualTx/<txid1>,<txid2>"
```

### Get Batch Sweep Transactions

```bash
curl "$ARK_SERVER/v1/indexer/batch/<txid>/<vout>/sweepTxs"
```

---

## ArkService Endpoints

### Get Server Info

```bash
curl "$ARK_SERVER/v1/info"
```

**Response:** `pubkey`, `vtxo_tree_expiry` (seconds), `unilateral_exit_delay` (seconds), `network`.

### Get Boarding Address

```bash
curl -X POST "$ARK_SERVER/v1/boarding/address" \
  -H "Content-Type: application/json" \
  -d '{"pubkey": "<user_pubkey_hex>"}'
```

---

## Combined Workflows

### Fetch and Decode a VTXO

```bash
# 1. Fetch from indexer
RESPONSE=$(curl -s "$ARK_SERVER/v1/indexer/vtxos?scripts=<script_hex>")

# 2. Inspect status fields
echo "$RESPONSE" | jq '.vtxos[] | {outpoint, is_spent, is_swept, expires_at}'

# 3. Decode script locally with noa
SCRIPT=$(echo "$RESPONSE" | jq -r '.vtxos[0].script')
noa script "$SCRIPT"
```

### Trace VTXO History

```bash
# 1. Get spending chain
CHAIN=$(curl -s "$ARK_SERVER/v1/indexer/vtxo/<txid>/<vout>/chain")

# 2. Fetch all virtual txs in the chain
TXIDS=$(echo "$CHAIN" | jq -r '.chain[].txid' | tr '\n' ',')
TXS=$(curl -s "$ARK_SERVER/v1/indexer/virtualTx/$TXIDS")

# 3. Decode each locally
echo "$TXS" | jq -r '.txs[]' | while read tx; do
  noa psbt decode "$tx"
done
```
