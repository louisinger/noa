# TypeScript SDK Debugging (@arkade-os/sdk)

```bash
npm install @arkade-os/sdk
```

## Key Classes

| Class | Purpose |
|-------|---------|
| `Wallet` | Full wallet — send, receive, inspect |
| `ReadonlyWallet` | Query-only — balance, VTXOs, history without signing |
| `VtxoManager` | VTXO renewal, recovery, expiry monitoring |
| `RestArkProvider` | Ark server communication |
| `EsploraProvider` | Blockchain data (confirmations, fee rates) |
| `ArkAddress` | Address encoding/decoding/inspection |

## Inspecting Wallet State

```typescript
import { Wallet, MnemonicIdentity, RestArkProvider, EsploraProvider } from '@arkade-os/sdk';

const wallet = await Wallet.create({
  identity: new MnemonicIdentity(mnemonic),
  arkProvider: new RestArkProvider('https://arkade.computer'),
  esploraProvider: new EsploraProvider('https://blockstream.info/api'),
});

// Balance breakdown
const balance = await wallet.getBalance();
// { settled, preconfirmed, available, recoverable, total, boarding }

// All VTXOs with status
const vtxos = await wallet.getVtxos();
// Each VTXO: { txid, vout, value, status, virtualStatus, isSpent, isUnrolled, expires_at, ... }

// Transaction history
const history = await wallet.getTransactionHistory();
```

## Diagnosing Balance Issues

```typescript
const balance = await wallet.getBalance();

// Funds exist but unavailable?
if (balance.recoverable > 0) {
  // VTXOs are swept, sub-dust, or expired — recoverable
  const { total, subdustBalance, vtxoCount } = await vtxoManager.getRecoverableBalance();
  await vtxoManager.recoverVtxos();  // Reclaim them
}

// VTXOs expiring soon?
const expiring = await vtxoManager.getExpiringVtxos();
if (expiring.length > 0) {
  await vtxoManager.renewVtxos();  // Extend before expiry
}
```

## Decoding an Ark Address (TypeScript)

```typescript
import { ArkAddress } from '@arkade-os/sdk';

const addr = ArkAddress.decode('ark1q...');
// addr.serverPubKey  — server's public key (Uint8Array, 32 bytes)
// addr.vtxoTaprootKey — VTXO taproot key (Uint8Array, 32 bytes)
// addr.pkScript      — the Bitcoin output script
// addr.subdustPkScript — sub-dust variant

// Cross-validate with noa:
// noa address ark1q...
```

## Validating Transaction Graphs

```typescript
import { validateVtxoTxGraph, verifyTapscriptSignatures } from '@arkade-os/sdk';

// Validate full VTXO tx graph
await validateVtxoTxGraph(graph, roundTransaction, sweepTapTreeRoot);

// Verify tapscript signatures on specific input
await verifyTapscriptSignatures(tx, inputIndex, requiredSigners);
```

## Monitoring Server & Chain State

```typescript
const arkProvider = new RestArkProvider('https://arkade.computer');
const esploraProvider = new EsploraProvider('https://blockstream.info/api');

// Server config
const info = await arkProvider.getInfo();
// info.vtxo_tree_expiry, info.unilateral_exit_delay, info.pubkey, info.network

// VTXO confirmation status
const status = await esploraProvider.getTxStatus(txid);
// status.confirmed, status.block_height, status.block_time

// Is a batch output still unspent onchain?
const outspends = await esploraProvider.getTxOutspends(commitmentTxid);
// outspends[vout].spent — false means batch output still available

// Current fee rates
const feeRate = await esploraProvider.getFeeRate();
```

## Real-Time Event Monitoring

```typescript
// Watch for incoming funds
const incoming = await waitForIncomingFunds(wallet);

// Stream settlement events (for round participation)
const stream = await arkProvider.getEventStream();
for await (const event of stream) {
  console.log(event);  // settlement, rejection, etc.
}

// Watch onboarding progress
await ramps.onboard(feeInfo, boardingUtxos, amount, (event) => {
  console.log('Onboard event:', event);
});
```

## Common Client-Side Issues

### "Balance shows 0 but I have funds"
1. Check `balance.recoverable` — VTXOs may be swept/expired
2. Check `balance.preconfirmed` — round may not be confirmed yet
3. Check `balance.boarding.unconfirmed` — onchain deposit not yet mined
4. Call `wallet.getVtxos()` and inspect each `virtualStatus`

### "VTXO not spendable"
1. `wallet.getVtxos()` → find VTXO → check `status` and `virtualStatus`
2. If `isUnrolled: true` — VTXO exited onchain, check with `esploraProvider.getTxStatus()`
3. If `isSpent: true` — check `spentBy` field for the spending tx

### "Transaction broadcast failed"
1. `esploraProvider.getFeeRate()` — is fee rate sufficient?
2. `validateVtxoTxGraph()` — is the tx graph valid?
3. Check if VTXO inputs are still unspent: `esploraProvider.getTxOutspends()`

### "Round participation rejected"
1. `arkProvider.getInfo()` — check server is responsive
2. Check `balance.boarding` — ensure boarding UTXOs have enough confirmations
3. `arkProvider.getPendingTxs(intent)` — inspect intent state
