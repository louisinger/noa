# arkd Server Reference

## Repository Structure

```
arkd/
├── api-spec/           # Protocol Buffer API definitions
│   └── protobuf/ark/v1/
│       ├── service.proto      # Main ArkService
│       ├── indexer.proto      # IndexerService
│       └── admin.proto        # AdminService
├── pkg/
│   ├── ark-lib/        # Reusable data structures and functions
│   │   ├── script/     # Closure types (Multisig, CSV, CLTV, Condition)
│   │   ├── note/       # Note closure implementation
│   │   └── txutils/    # PSBT and taptree utilities
│   ├── arkd-wallet/    # Bitcoin wallet (liquidity provider + signer)
│   └── ark-cli/        # CLI wallet implementation
├── internal/
│   ├── core/
│   │   ├── application/  # Core business logic
│   │   ├── domain/       # Models and events
│   │   └── ports/        # Interface definitions
│   └── infrastructure/   # Port implementations (DB, cache, etc.)
└── test/e2e/           # Integration tests
```

## Key Configuration (Environment Variables)

| Variable | Description | Default |
|----------|-------------|---------|
| `ARKD_VTXO_TREE_EXPIRY` | VTXO tree expiry (seconds) | 604672 (7 days) |
| `ARKD_UNILATERAL_EXIT_DELAY` | Exit delay (seconds) | 86400 (24 hours) |
| `ARKD_BOARDING_EXIT_DELAY` | Boarding exit delay (seconds) | 7776000 (3 months) |
| `ARKD_SESSION_DURATION` | Batch session timeout (seconds) | 30 |
| `ARKD_ROUND_MAX_PARTICIPANTS_COUNT` | Max participants per round | 128 |
| `ARKD_ROUND_MIN_PARTICIPANTS_COUNT` | Min participants per round | 1 |
| `ARKD_BAN_DURATION` | Ban duration (seconds) | 300 (5 mins) |
| `ARKD_BAN_THRESHOLD` | Crimes to trigger ban | 3 |
| `ARKD_DB_TYPE` | Database type | postgres |
| `ARKD_LIVE_STORE_TYPE` | Cache type | redis |
| `ARKD_SCHEDULER_TYPE` | Scheduler type | gocron |

## Core Components

### Transaction Builder (`ARKD_TX_BUILDER_TYPE`)
Currently supports `covenantless` mode — creates VTXO trees without covenants, uses taproot script trees, relies on server cosigning for collaborative spends.

### Scheduler (`ARKD_SCHEDULER_TYPE`)
- `gocron`: Time-based round scheduling
- `block`: Block-based round scheduling

### Database (`ARKD_DB_TYPE`)
- `postgres`: Production
- `sqlite`: Lightweight
- `badger`: Embedded key-value store

### Cache (`ARKD_LIVE_STORE_TYPE`)
- `redis`: Production (conflict retry support)
- `inmemory`: Development/testing only

## Common Server Issues

### Round Not Starting

**Symptoms:** No new rounds being created

1. Check scheduler type and configuration
2. Verify `ARKD_ROUND_MIN_PARTICIPANTS_COUNT` — may be waiting for participants
3. Check wallet balance — server needs liquidity
4. Enable trace logging: `ARKD_LOG_LEVEL=6`

### VTXO Not Found in Indexer

**Symptoms:** Client can't find their VTXO

1. Check if round was confirmed onchain
2. Verify `is_preconfirmed` status
3. Check if VTXO was swept (`is_swept=true`)
4. Verify correct network (mainnet vs testnet)

### Transaction Rejected by Server

**Symptoms:** Server returns error during round participation

1. Check if user is banned (`ARKD_BAN_DURATION`, `ARKD_BAN_THRESHOLD`)
2. Verify amounts within limits: `ARKD_UTXO_MIN_AMOUNT` / `ARKD_UTXO_MAX_AMOUNT`, `ARKD_VTXO_MIN_AMOUNT` / `ARKD_VTXO_MAX_AMOUNT`
3. Check session timing — may have exceeded `ARKD_SESSION_DURATION`
4. Verify PSBT format and signatures

### Sweep Not Happening

**Symptoms:** Expired VTXOs not being swept

1. Check `ARKD_VTXO_TREE_EXPIRY` configuration
2. Verify server wallet has funds for sweep tx fees
3. Check for pending sweeps in indexer
4. Look for unilateral exits that may have unrolled the tree

## Debugging Tips

### Enable Trace Logging
```bash
export ARKD_LOG_LEVEL=6
```

### Check Round State
```bash
curl -X GET "http://localhost:7071/v1/admin/round/current"
```

### Inspect Database (Postgres)
```sql
-- Check pending VTXOs
SELECT * FROM vtxos WHERE spent = false;

-- Check recent rounds
SELECT * FROM rounds ORDER BY created_at DESC LIMIT 10;
```

### Test with Regtest
```bash
nigiri start
make run-wallet
make run-light  # sqlite + inmemory cache
```

## Code Navigation

| What | Where |
|------|-------|
| Closure encoding | `pkg/ark-lib/script/closure.go`, `multisig.go`, `csv_multisig.go` |
| PSBT handling | `pkg/ark-lib/txutils/psbt.go`, `taptree.go` |
| Round logic | `internal/core/application/round.go`, `domain/round.go` |
| Indexer implementation | `internal/interface/grpc/indexer.go` |

## Closure Implementation Details

```go
// MultisigClosure — simple n-of-n
type MultisigClosure struct { PubKeys []*btcec.PublicKey }

// CSVMultisigClosure — adds OP_CHECKSEQUENCEVERIFY
type CSVMultisigClosure struct {
    MultisigClosure
    Locktime RelativeLocktime  // Type: Blocks (0) or Seconds (1, in 512s intervals)
}

// CLTVMultisigClosure — adds OP_CHECKLOCKTIMEVERIFY
type CLTVMultisigClosure struct {
    MultisigClosure
    Locktime AbsoluteLocktime  // Block height or Unix timestamp
}

// ConditionMultisigClosure — adds arbitrary condition script
type ConditionMultisigClosure struct {
    MultisigClosure
    Condition []byte  // Custom script (HTLCs, notes, etc.)
}
```

## PSBT Field Constants

```go
// pkg/ark-lib/txutils/
const (
    ConditionWitnessField    = 0x00
    CosignerPublicKeyField   = 0x01
    VtxoTaprootTreeField     = 0x02
    VtxoTreeExpiryField      = 0x03
)
```

## API Protocol Buffers (`api-spec/protobuf/ark/v1/`)

**service.proto (ArkService):** `GetInfo`, `RegisterInputsForNextRound`, `SubmitSignedForfeitTxs`, `GetRound`, `Ping`

**indexer.proto (IndexerService):** `GetVtxos`, `GetVtxoChain`, subscription endpoints for real-time updates

**admin.proto (AdminService):** Server administration, wallet management — requires authentication

## arkd-wallet Integration

```bash
ARKD_WALLET_NBXPLORER_URL=http://localhost:32838
ARKD_WALLET_SIGNER_KEY=<private_key_hex>  # If using wallet as signer
```

Key files: `pkg/arkd-wallet/` — uses NBXplorer for blockchain data.
