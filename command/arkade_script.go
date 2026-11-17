package command

import (
	"strings"

	"github.com/arkade-os/emulator/pkg/arkade"
	"github.com/btcsuite/btcd/txscript/v2"
)

// Use the emulator's canonical names instead of maintaining a separate opcode table.
var arkadeOpcodeNames = func() map[byte]string {
	names := make(map[byte]string)
	for name, opcode := range arkade.OpcodeByName {
		switch name {
		case "OP_FALSE", "OP_TRUE", "OP_NOP2", "OP_NOP3", "OP_MERKLEBRANCHVERIFY":
			continue // aliases of canonical names already in the upstream table
		}
		names[opcode] = name
	}
	return names
}()

// disasmArkadeScript uses the emulator's tokenizer and opcode names. Pushes and
// small integers retain the compact formatting used by Bitcoin's disassembler.
func disasmArkadeScript(script []byte) (string, error) {
	var tokens []string
	tokenizer := arkade.MakeScriptTokenizer(0, script)
	var start int32
	for tokenizer.Next() {
		name, ok := arkadeOpcodeNames[tokenizer.Opcode()]
		if !ok || tokenizer.Opcode() <= arkade.OP_16 {
			var err error
			name, err = txscript.DisasmString(script[start:tokenizer.ByteIndex()])
			if err != nil {
				return "", err
			}
		}
		tokens = append(tokens, name)
		start = tokenizer.ByteIndex()
	}
	if err := tokenizer.Err(); err != nil {
		return "", err
	}
	return strings.Join(tokens, " "), nil
}
