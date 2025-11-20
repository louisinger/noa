package command

import (
	"encoding/hex"
	"fmt"

	"github.com/arkade-os/arkd/pkg/ark-lib/note"
	"github.com/arkade-os/arkd/pkg/ark-lib/script"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
)

func RunNoteFromTxid(txidString string) error {
	preimageHashBytes, err := chainhash.NewHashFromStr(txidString)
	if err != nil {
		return fmt.Errorf("failed to decode preimage hash: %w", err)
	}
	if len(preimageHashBytes) != 32 {
		return fmt.Errorf("preimage hash must be 32 bytes")
	}

	hash := [32]byte(preimageHashBytes.CloneBytes())
	vtxoScript := script.TapscriptsVtxoScript{
		Closures: []script.Closure{&note.NoteClosure{PreimageHash: hash}},
	}

	tapkey, _, err := vtxoScript.TapTree()
	if err != nil {
		return fmt.Errorf("failed to get tapkey: %w", err)
	}

	pkScript, err := txscript.PayToTaprootScript(tapkey)
	if err != nil {
		return fmt.Errorf("failed to create pk script: %w", err)
	}

	var output string

	// Tapkey
	output += fmt.Sprintf("\n%s\n",
		sectionStyle.Render("Tapkey:"),
	)
	tapkeyBytes := schnorr.SerializePubKey(tapkey)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("hex:"),
		valueStyle.Render(hex.EncodeToString(tapkeyBytes)),
	)

	// Script
	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("Script:"),
	)
	scriptHex := hex.EncodeToString(pkScript)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("hex:"),
		valueStyle.Render(scriptHex),
	)

	disasm, err := txscript.DisasmString(pkScript)
	if err == nil {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("asm:"),
			valueStyle.Render(disasm),
		)
	}

	fmt.Print(output)
	return nil
}
