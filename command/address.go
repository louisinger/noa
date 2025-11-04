package command

import (
	"encoding/hex"
	"fmt"

	arklib "github.com/arkade-os/arkd/pkg/ark-lib"
	"github.com/btcsuite/btcd/txscript"
	"github.com/charmbracelet/lipgloss"
)

var (
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	subLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			MarginLeft(4).
			MarginRight(1)

	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Bold(true)

	addressLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				MarginRight(1)

	commonLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				MarginRight(1)
)

func RunAddress(addressArk string) error {
	decoded, err := arklib.DecodeAddressV0(addressArk)
	if err != nil {
		return fmt.Errorf("failed to decode address: %w", err)
	}

	var output string

	// Address ARK
	output += fmt.Sprintf("\n%s%s\n",
		addressLabelStyle.Render("Address:"),
		valueStyle.Render(addressArk),
	)

	// Version and HRP
	output += fmt.Sprintf("%s%s  %s%s\n",
		commonLabelStyle.Render("Version:"),
		valueStyle.Render(fmt.Sprintf("%d", decoded.Version)),
		commonLabelStyle.Render("HRP:"),
		valueStyle.Render(decoded.HRP),
	)

	// Public Keys
	output += fmt.Sprintf("%s\n",
		sectionStyle.Render("Public Keys:"),
	)

	// Signer Public Key
	if decoded.Signer != nil {
		signerBytes := decoded.Signer.SerializeCompressed()
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("signer:"),
			valueStyle.Render(hex.EncodeToString(signerBytes)),
		)
	} else {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("signer:"),
			valueStyle.Render("<nil>"),
		)
	}

	// VTXO Taproot Key
	if decoded.VtxoTapKey != nil {
		vtxoTapKeyBytes := decoded.VtxoTapKey.SerializeCompressed()
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("tapkey:"),
			valueStyle.Render(hex.EncodeToString(vtxoTapKeyBytes)),
		)
	} else {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("tapkey:"),
			valueStyle.Render("<nil>"),
		)
	}

	pkScript, err := decoded.GetPkScript()
	if err == nil && pkScript != nil {
		output += fmt.Sprintf("%s\n",
			sectionStyle.Render("Script:"),
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
	}

	fmt.Print(output)
	return nil
}
