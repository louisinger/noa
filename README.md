# noa

Your Ark companion.

## Installation

```bash
go install github.com/louisinger/noa@latest
```

Or build from source:

```bash
make install
```

## Usage

### address

```bash
noa address <address_ark>
```

Decodes an ARK address and displays:
- Address ARK
- Version and HRP (Human Readable Part)
- Public Keys (signer and tapkey)
- Script information (hex and asm)

### script

```bash
noa script <script_hex>
```

Decodes a script and displays:
- ASM disassembly
- Ark Closure information (type and fields)

### note

#### fromTxid

```bash
noa note fromTxid <txid_string>
```

Generates a note closure from a transaction ID (32-byte hash) and displays:
- Tapkey (hex)
- Script (hex and asm)

The txid is used as the preimage hash for the note closure, which is then used to generate the taproot tapkey and corresponding script.

### taptree

#### decode

```bash
noa taptree decode <taptree_hex>
```

Decodes a taptree (hex-encoded) and displays:
- All scripts in the taptree (hex and asm)
- Output script (hex and asm)

#### encode

```bash
noa taptree encode <script1_hex> [script2_hex] ...
```

Encodes one or more scripts into a taptree and displays the encoded taptree (hex).

### psbt

#### decode

```bash
noa psbt decode <psbt_base64>
```

Decodes a PSBT (Partially Signed Bitcoin Transaction) from base64 or hex format and displays:
- Global transaction information (version, locktime, txid)
- Inputs with:
  - Previous outpoint and sequence
  - Redeem scripts and witness scripts
  - BIP32 derivation paths
  - Witness UTXO information
  - **ARK PSBT fields** (when present):
    - ConditionWitness
    - CosignerPublicKey
    - VtxoTaprootTree
    - VtxoTreeExpiry
- Outputs with:
  - Value and script (hex and asm)
  - Redeem scripts and witness scripts
  - BIP32 derivation paths

The command automatically detects whether the input is base64 or hex encoded.
