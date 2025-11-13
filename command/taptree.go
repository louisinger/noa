package command

import (
	"encoding/hex"
	"fmt"

	"github.com/arkade-os/arkd/pkg/ark-lib/script"
	"github.com/arkade-os/arkd/pkg/ark-lib/txutils"
	"github.com/btcsuite/btcd/txscript"
)

func RunTaptreeDecode(input string) error {
	bytesInput, err := hex.DecodeString(input)
	if err != nil {
		return fmt.Errorf("failed to decode input: %w", err)
	}
	taptree, err := txutils.DecodeTapTree(bytesInput)
	if err != nil {
		return fmt.Errorf("failed to decode taptree: %w", err)
	}

	var output string

	// Print scripts in taptree
	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("TapTree Scripts:"),
	)
	for i, scriptHex := range taptree {
		scriptBytes, err := hex.DecodeString(scriptHex)
		if err != nil {
			return fmt.Errorf("failed to decode script [%d]: %w", i, err)
		}

		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render(fmt.Sprintf("[%d]:", i)),
		)
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  hex:"),
			valueStyle.Render(scriptHex),
		)

		disasm, err := txscript.DisasmString(scriptBytes)
		if err == nil {
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("  asm:"),
				valueStyle.Render(disasm),
			)
		}
	}

	// Get tapkey and create pk script
	vtxoScript, err := script.ParseVtxoScript(taptree)
	if err != nil {
		return fmt.Errorf("failed to parse vtxo script: %w", err)
	}

	tapkey, _, err := vtxoScript.TapTree()
	if err != nil {
		return fmt.Errorf("failed to get tapkey: %w", err)
	}

	// Create pk script from tapkey
	pkScript, err := txscript.PayToTaprootScript(tapkey)
	if err != nil {
		return fmt.Errorf("failed to create pk script: %w", err)
	}

	// Print pk script
	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("PkScript:"),
	)

	hexStr := hex.EncodeToString(pkScript)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("hex:"),
		valueStyle.Render(hexStr),
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

func RunTaptreeEncode(input []string) error {
	var output string

	// Print input scripts
	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("Input Scripts:"),
	)
	for i, scriptHex := range input {
		scriptBytes, err := hex.DecodeString(scriptHex)
		if err != nil {
			return fmt.Errorf("failed to decode input script [%d]: %w", i, err)
		}

		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render(fmt.Sprintf("[%d]:", i)),
		)
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  hex:"),
			valueStyle.Render(scriptHex),
		)

		disasm, err := txscript.DisasmString(scriptBytes)
		if err == nil {
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("  asm:"),
				valueStyle.Render(disasm),
			)
		}
	}

	// Encode taptree
	taptree := txutils.TapTree(input)
	bytes, err := taptree.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode taptree: %w", err)
	}

	// Print encoded output
	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("Encoded TapTree:"),
	)
	hexStr := hex.EncodeToString(bytes)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("hex:"),
		valueStyle.Render(hexStr),
	)

	fmt.Print(output)
	return nil
}
