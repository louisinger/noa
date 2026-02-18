# Ark Protocol Concepts

> Source: [Arkade Documentation](https://docs.arkadeos.com)

## VTXOs (Virtual Transaction Outputs)

VTXOs are Bitcoin's UTXOs reimagined for offchain execution — programmable, offchain objects that mirror Bitcoin's UTXO structure but exist within Arkade's Virtual Mempool.

**Key properties:**
- Each VTXO is backed by a presigned, unbroadcast Bitcoin transaction
- Represents a user's claim to a specific portion of value within a batch output
- Includes defined ownership, value, and spending conditions
- Supports both collaborative and unilateral exit paths

**VTXO Structure (Two Spending Paths):**

1. **Collaborative Path**: Default for Arkade transactions. Requires owner's signature + cosignature from the Arkade Signer.
2. **Exit Path (Unilateral)**: Exclusively controlled by VTXO owner. Allows moving offchain funds onchain at any time. Protected by CSV delay to prevent double-spending.

**VTXO States:**

| State | Description |
|-------|-------------|
| **Spent** | Used as input to another Arkade transaction, no longer active |
| **Recoverable** | Valid in protocol but cannot exit unilaterally (sub-dust, expired, notes) |
| **Confirmed (Settled)** | Anchored onchain, full Bitcoin security with unilateral exit rights |
| **Unconfirmed** | Included in commitment tx, awaiting Bitcoin confirmation |
| **Preconfirmed** | Created offchain, validated and cosigned, but not yet settled onchain |

**VTXO Expiration:**
- Every VTXO is tied to the lifetime of its batch output
- Users must periodically renew VTXOs before expiry to maintain unilateral exit rights
- If expiration is missed, operator regains control and issues an Arkade Note in exchange

## Batch Outputs

A batch output consolidates multiple users' ownership claims into a single Bitcoin output through a tree of presigned virtual transactions.

**Technical Structure:**
- Locked by a taproot script with an n-of-n MuSig2 internal key
- All VTXO owners are cosigners
- Unspendable key path with two script paths:
  1. **Unroll path**: Split batch output into separate VTXO branches
  2. **Sweep path**: Operator can spend after batch expiry timeout

**Virtual Transaction Tree:**
- Leaves of the tree are the VTXOs (user ownership claims)
- Enables selective unrolling — users can exit without affecting others
- MuSig2 allows tree to appear as single signature onchain

## Batch Swaps

Batch swaps aggregate multiple Arkade transactions into a single onchain commitment transaction.

**Why settle via batch swap?**
1. **VTXO Renewal**: Must renew before batch expiry or operator can claim
2. **Chain Depth**: Deeply chained VTXOs have high unilateral exit costs
3. **Trust Surface**: Preconfirmed VTXOs rely on operator integrity

**Commitment Transaction Structure:**
- **Connector Output**: Dust-amount output ensuring atomicity
- **Batch Output**: Aggregates all participants' new VTXOs

**Forfeit Transactions** protect operators against user fraud:
- Inputs: One connector input + one VTXO input
- Outputs: Anchor output (fee management) + forfeit output (to operator)
- Only claimable if new batch confirms onchain

## Arkade Transactions

Virtual Bitcoin transactions that execute offchain in the virtual mempool.

**Properties:**
- Consume existing VTXOs as inputs, produce new VTXOs as outputs
- Instant confirmation through operator cosignature (preconfirmation)
- Can be chained indefinitely without waiting for onchain confirmation
- **Warning**: Longer chains = higher unilateral exit costs

## Unilateral Exit

**Exit Process:**
1. Publish complete transaction path from batch output to VTXO
2. "Unroll" the virtual tree structure with sequential Bitcoin transactions

**Exit Costs:**

| Chain Depth | Transactions Required |
|-------------|----------------------|
| Direct child of Batch Output | 1 Bitcoin tx |
| Second-level VTXO | 2 Bitcoin txs |
| Third-level VTXO | 3 Bitcoin txs |

**Economic Consideration:** For small VTXOs with high fees, exit costs may exceed balance, stranding funds until fees decrease.

## Transaction Finality

**Preconfirmation** — Transaction cosigned by Arkade Signer. Instant execution. Relies on operator integrity.

**Bitcoin Finality** — VTXO anchored to blockchain via batch swap. Full Bitcoin security guarantees.

## Ark Addresses

Ark addresses encode:
- **Version**: Protocol version (currently 0)
- **HRP**: Human-readable prefix (`ark` mainnet, `tark` testnet, `sark` signet)
- **Signer**: User's public key for signing transactions
- **Tapkey**: The taproot output key derived from the VTXO script tree

| HRP | Network |
|-----|---------|
| `ark` | Bitcoin mainnet |
| `tark` | Bitcoin testnet |
| `sark` | Bitcoin signet |
| `bcrt` | Bitcoin regtest |

## Closures (Script Types)

Closures are spending conditions in Ark VTXO scripts:

| Closure Type | Description |
|--------------|-------------|
| **MultisigClosure** | n-of-n multisig, all specified keys must sign |
| **CLTVMultisigClosure** | Multisig + absolute timelock (block height or timestamp) |
| **CSVMultisigClosure** | Multisig + relative timelock (e.g., "24 hours after confirmation") |
| **ConditionMultisigClosure** | Multisig + custom condition (e.g., hash preimage for HTLC) |
| **ConditionCSVMultisigClosure** | Combines CSV timelock + custom condition |

**Locktime Types:**
- `Blocks`: Locktime measured in block height
- `Seconds`: Locktime measured in 512-second intervals (CSV) or Unix timestamp (CLTV)

## Taptrees

Ark uses taproot script trees to encode multiple spending paths:
- **Forfeit path**: Server can claim after timeout (CSV)
- **Unilateral exit**: User can exit onchain (with server cosign or after delay)
- **Collaborative path**: Keypath spend with user + server

## ARK PSBT Fields

Custom PSBT fields used in Ark transactions:

| Field | Description |
|-------|-------------|
| `ConditionWitness` | Witness stack for condition scripts (e.g., preimages) |
| `CosignerPublicKey` | Identifies which key should cosign |
| `VtxoTaprootTree` | The full taptree for VTXO reconstruction |
| `VtxoTreeExpiry` | When the VTXO tree expires (Blocks or Seconds) |

## Glossary

| Term | Definition |
|------|------------|
| **Arkade Operator** | Coordinator who validates transactions, aggregates into commitment txs. Cannot unilaterally spend user VTXOs. |
| **Arkade Signer** | Independent entity managing cosigning keys. Operates in TEE, isolated from operator. |
| **Virtual Mempool** | Offchain execution engine processing transactions in real-time using DAG structure. |
| **Intent** | BIP322-based signed message proving ownership, used for batch swap participation or delegation. |
| **Connector Output** | Dust-amount output ensuring atomicity between old/new states during batch swaps. |
| **Batch Expiry** | Timeout after which operator can sweep unclaimed batch outputs. |
| **Preconfirmation** | Instant confirmation via operator cosignature, before onchain settlement. |
| **MuSig2** | Schnorr multi-signature scheme enabling compact onchain footprint. |
| **TEE** | Trusted Execution Environment (hardware enclave) protecting signer keys. |
