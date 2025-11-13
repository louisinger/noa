package command

import (
	"encoding/hex"
	"fmt"

	arklib "github.com/arkade-os/arkd/pkg/ark-lib"
	"github.com/arkade-os/arkd/pkg/ark-lib/script"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
)

func RunScript(scriptHex string) error {
	// Decode hex string to bytes
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return fmt.Errorf("failed to decode hex string: %w", err)
	}

	var output string

	// ASM disassembly
	disasm, err := txscript.DisasmString(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to disassemble script: %w", err)
	}

	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("asm:"),
		valueStyle.Render(disasm),
	)

	closure, err := script.DecodeClosure(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to decode closure: %w", err)
	}

	output += sectionStyle.Render("\nClosure: ")

	// Print closure type and fields using type switch
	output += formatClosure(closure)

	fmt.Print(output)
	return nil
}

// formatClosure formats a closure based on its concrete type
func formatClosure(closure script.Closure) string {
	var output string

	switch c := closure.(type) {
	case *script.MultisigClosure:
		output += fmt.Sprintf("%s\n",
			valueStyle.Render("MultisigClosure"),
		)
		output += formatMultisigClosure(c)

	case *script.CLTVMultisigClosure:
		output += fmt.Sprintf("%s\n",
			valueStyle.Render("CLTVMultisigClosure"),
		)
		output += formatMultisigClosure(&c.MultisigClosure)
		output += fmt.Sprintf("%s\n",
			sectionStyle.Render("Locktime:"),
		)
		output += formatAbsoluteLocktime(c.Locktime)

	case *script.CSVMultisigClosure:
		output += fmt.Sprintf("%s\n",
			valueStyle.Render("CSVMultisigClosure"),
		)
		output += formatMultisigClosure(&c.MultisigClosure)
		output += fmt.Sprintf("%s\n",
			sectionStyle.Render("Locktime:"),
		)
		output += formatRelativeLocktime(c.Locktime)

	case *script.ConditionMultisigClosure:
		output += fmt.Sprintf("%s\n",
			valueStyle.Render("ConditionMultisigClosure"),
		)
		output += formatMultisigClosure(&c.MultisigClosure)
		output += fmt.Sprintf("%s\n",
			sectionStyle.Render("Condition:"),
		)
		output += formatConditionScript(c.Condition)

	case *script.ConditionCSVMultisigClosure:
		output += fmt.Sprintf("%s\n",
			valueStyle.Render("ConditionCSVMultisigClosure"),
		)
		output += formatMultisigClosure(&c.CSVMultisigClosure.MultisigClosure)
		output += fmt.Sprintf("%s\n",
			sectionStyle.Render("Locktime:"),
		)
		output += formatRelativeLocktime(c.CSVMultisigClosure.Locktime)
		output += fmt.Sprintf("%s\n",
			sectionStyle.Render("Condition:"),
		)
		output += formatConditionScript(c.Condition)

	default:
		output += fmt.Sprintf("%s\n",
			valueStyle.Render(fmt.Sprintf("%T", closure)),
		)
		output += fmt.Sprintf("%s\n",
			valueStyle.Render(fmt.Sprintf("%+v", closure)),
		)
	}

	return output
}

// formatMultisigClosure formats the common MultisigClosure fields
func formatMultisigClosure(m *script.MultisigClosure) string {
	var output string

	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("PubKeys:"),
	)
	for i, pubKey := range m.PubKeys {
		if pubKey != nil {
			pubKeyBytes := schnorr.SerializePubKey(pubKey)
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render(fmt.Sprintf("[%d]:", i)),
				valueStyle.Render(hex.EncodeToString(pubKeyBytes)),
			)
		} else {
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render(fmt.Sprintf("[%d]:", i)),
				valueStyle.Render("<nil>"),
			)
		}
	}

	return output
}

// formatAbsoluteLocktime formats an AbsoluteLocktime
func formatAbsoluteLocktime(lt arklib.AbsoluteLocktime) string {
	var output string
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("Type:"),
		valueStyle.Render(formatAbsoluteLocktimeType(lt)),
	)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("Value:"),
		valueStyle.Render(fmt.Sprintf("%d", uint32(lt))),
	)
	return output
}

// formatAbsoluteLocktimeType formats the type of an AbsoluteLocktime
func formatAbsoluteLocktimeType(lt arklib.AbsoluteLocktime) string {
	if lt.IsSeconds() {
		return "Seconds"
	}
	return "Blocks"
}

// formatRelativeLocktime formats a RelativeLocktime
func formatRelativeLocktime(lt arklib.RelativeLocktime) string {
	var output string
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("Type:"),
		valueStyle.Render(formatRelativeLocktimeType(lt.Type)),
	)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("Value:"),
		valueStyle.Render(fmt.Sprintf("%d", lt.Value)),
	)
	return output
}

// formatRelativeLocktimeType formats a RelativeLocktimeType
func formatRelativeLocktimeType(t arklib.RelativeLocktimeType) string {
	switch t {
	case arklib.LocktimeTypeSecond:
		return "Seconds"
	default:
		// Default (0) is blocks
		return "Blocks"
	}
}

// formatConditionScript formats a condition script (byte slice) as both hex and asm
func formatConditionScript(condition []byte) string {
	var output string

	// Hex
	hexStr := hex.EncodeToString(condition)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("hex:"),
		valueStyle.Render(hexStr),
	)

	// ASM disassembly
	disasm, err := txscript.DisasmString(condition)
	if err == nil {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("asm:"),
			valueStyle.Render(disasm),
		)
	}

	return output
}
