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
