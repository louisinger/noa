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

```bash
noa address <address_ark>
```

This command decodes an ARK address and displays:
- Address ARK
- Version and HRP (Human Readable Part)
- Public Keys (signer and tapkey)
- Script information (hex and asm)

