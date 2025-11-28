package command

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/arkade-os/arkd/pkg/ark-lib/txutils"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"
)

func RunPsbtDecode(psbtInput string) error {
	var psbtBytes []byte
	var err error

	// Try base64 first (most common for PSBT)
	psbtBytes, err = base64.StdEncoding.DecodeString(strings.TrimSpace(psbtInput))
	if err != nil {
		// Fall back to hex
		psbtBytes, err = hex.DecodeString(strings.TrimSpace(psbtInput))
		if err != nil {
			return fmt.Errorf("failed to decode PSBT (tried base64 and hex): %w", err)
		}
	}

	// Parse PSBT
	p, err := psbt.NewFromRawBytes(bytes.NewReader(psbtBytes), false)
	if err != nil {
		return fmt.Errorf("failed to parse PSBT: %w", err)
	}

	var output string

	// Global transaction
	output += fmt.Sprintf("\n%s\n",
		sectionStyle.Render("Global:"),
	)
	tx := p.UnsignedTx
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("Version:"),
		valueStyle.Render(fmt.Sprintf("%d", tx.Version)),
	)
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("LockTime:"),
		valueStyle.Render(fmt.Sprintf("%d", tx.LockTime)),
	)
	if tx.TxHash().String() != "" {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("TxId:"),
			valueStyle.Render(tx.TxHash().String()),
		)
	}

	// Inputs
	output += fmt.Sprintf("\n%s\n",
		sectionStyle.Render(fmt.Sprintf("Inputs (%d):", len(tx.TxIn))),
	)
	for i, txIn := range tx.TxIn {
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render(fmt.Sprintf("[%d]:", i)),
		)
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  PreviousOutPoint:"),
			valueStyle.Render(txIn.PreviousOutPoint.String()),
		)
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  Sequence:"),
			valueStyle.Render(fmt.Sprintf("%d", txIn.Sequence)),
		)

		// PSBT input specific data
		if i < len(p.Inputs) {
			in := p.Inputs[i]
			if in.RedeemScript != nil {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("  RedeemScript:"),
					valueStyle.Render(hex.EncodeToString(in.RedeemScript)),
				)
			}
			if in.WitnessScript != nil {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("  WitnessScript:"),
					valueStyle.Render(hex.EncodeToString(in.WitnessScript)),
				)
			}
			if len(in.Bip32Derivation) > 0 {
				output += fmt.Sprintf("%s\n",
					subLabelStyle.Render("  Bip32Derivation:"),
				)
				for j, der := range in.Bip32Derivation {
					fpBytes := make([]byte, 4)
					binary.BigEndian.PutUint32(fpBytes, der.MasterKeyFingerprint)
					output += fmt.Sprintf("%s%s\n",
						subLabelStyle.Render(fmt.Sprintf("    [%d] MasterFingerprint:", j)),
						valueStyle.Render(hex.EncodeToString(fpBytes)),
					)
					output += fmt.Sprintf("%s%s\n",
						subLabelStyle.Render(fmt.Sprintf("    [%d] Path:", j)),
						valueStyle.Render(formatBip32Path(der.Bip32Path)),
					)
					output += fmt.Sprintf("%s%s\n",
						subLabelStyle.Render(fmt.Sprintf("    [%d] PubKey:", j)),
						valueStyle.Render(hex.EncodeToString(der.PubKey)),
					)
				}
			}
			if in.NonWitnessUtxo != nil {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("  NonWitnessUtxo:"),
					valueStyle.Render("present"),
				)
			}
			if in.WitnessUtxo != nil {
				output += fmt.Sprintf("%s\n",
					subLabelStyle.Render("  WitnessUtxo:"),
				)
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("    Value:"),
					valueStyle.Render(fmt.Sprintf("%d sats", in.WitnessUtxo.Value)),
				)
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("    PkScript:"),
					valueStyle.Render(hex.EncodeToString(in.WitnessUtxo.PkScript)),
				)
			}

			// Decode ARK PSBT fields
			output += formatArkPsbtFields(p, i)
		}
	}

	// Outputs
	output += fmt.Sprintf("\n%s\n",
		sectionStyle.Render(fmt.Sprintf("Outputs (%d):", len(tx.TxOut))),
	)
	for i, txOut := range tx.TxOut {
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render(fmt.Sprintf("[%d]:", i)),
		)
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  Value:"),
			valueStyle.Render(fmt.Sprintf("%d sats", txOut.Value)),
		)
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  PkScript:"),
			valueStyle.Render(hex.EncodeToString(txOut.PkScript)),
		)

		disasm, err := txscript.DisasmString(txOut.PkScript)
		if err == nil {
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("  Script ASM:"),
				valueStyle.Render(disasm),
			)
		}

		// PSBT output specific data
		if i < len(p.Outputs) {
			out := p.Outputs[i]
			if out.RedeemScript != nil {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("  RedeemScript:"),
					valueStyle.Render(hex.EncodeToString(out.RedeemScript)),
				)
			}
			if out.WitnessScript != nil {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("  WitnessScript:"),
					valueStyle.Render(hex.EncodeToString(out.WitnessScript)),
				)
			}
			if len(out.Bip32Derivation) > 0 {
				output += fmt.Sprintf("%s\n",
					subLabelStyle.Render("  Bip32Derivation:"),
				)
				for j, der := range out.Bip32Derivation {
					fpBytes := make([]byte, 4)
					binary.BigEndian.PutUint32(fpBytes, der.MasterKeyFingerprint)
					output += fmt.Sprintf("%s%s\n",
						subLabelStyle.Render(fmt.Sprintf("    [%d] MasterFingerprint:", j)),
						valueStyle.Render(hex.EncodeToString(fpBytes)),
					)
					output += fmt.Sprintf("%s%s\n",
						subLabelStyle.Render(fmt.Sprintf("    [%d] Path:", j)),
						valueStyle.Render(formatBip32Path(der.Bip32Path)),
					)
					output += fmt.Sprintf("%s%s\n",
						subLabelStyle.Render(fmt.Sprintf("    [%d] PubKey:", j)),
						valueStyle.Render(hex.EncodeToString(der.PubKey)),
					)
				}
			}
		}
	}

	fmt.Print(output)
	return nil
}

func formatBip32Path(path []uint32) string {
	if len(path) == 0 {
		return "<empty>"
	}
	pathStr := "m"
	for _, p := range path {
		if p >= 0x80000000 {
			pathStr += fmt.Sprintf("/%d\"", p-0x80000000)
		} else {
			pathStr += fmt.Sprintf("/%d", p)
		}
	}
	return pathStr
}

func formatArkPsbtFields(p *psbt.Packet, inputIndex int) string {
	var output string
	hasAnyFields := false

	// Condition Witness Field
	conditionWitnesses, err := txutils.GetArkPsbtFields(p, inputIndex, txutils.ConditionWitnessField)
	if err == nil && len(conditionWitnesses) > 0 {
		if !hasAnyFields {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render("  ARK PSBT Fields:"),
			)
			hasAnyFields = true
		}
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("    ConditionWitness:"),
		)
		for j, witness := range conditionWitnesses {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render(fmt.Sprintf("      [%d]:", j)),
			)
			for k, item := range witness {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render(fmt.Sprintf("        [%d]:", k)),
					valueStyle.Render(hex.EncodeToString(item)),
				)
			}
		}
	}

	// Cosigner Public Key Field
	cosignerKeys, err := txutils.GetArkPsbtFields(p, inputIndex, txutils.CosignerPublicKeyField)
	if err == nil && len(cosignerKeys) > 0 {
		if !hasAnyFields {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render("  ARK PSBT Fields:"),
			)
			hasAnyFields = true
		}
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("    CosignerPublicKey:"),
		)
		for j, cosignerKey := range cosignerKeys {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render(fmt.Sprintf("      [%d]:", j)),
			)
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("        Index:"),
				valueStyle.Render(fmt.Sprintf("%d", cosignerKey.Index)),
			)
			if cosignerKey.PublicKey != nil {
				pubKeyBytes := schnorr.SerializePubKey(cosignerKey.PublicKey)
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render("        PublicKey:"),
					valueStyle.Render(hex.EncodeToString(pubKeyBytes)),
				)
			}
		}
	}

	// VTXO Taproot Tree Field
	vtxoTaprootTrees, err := txutils.GetArkPsbtFields(p, inputIndex, txutils.VtxoTaprootTreeField)
	if err == nil && len(vtxoTaprootTrees) > 0 {
		if !hasAnyFields {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render("  ARK PSBT Fields:"),
			)
			hasAnyFields = true
		}
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("    VtxoTaprootTree:"),
		)
		for j, tree := range vtxoTaprootTrees {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render(fmt.Sprintf("      [%d]:", j)),
			)
			for k, scriptHex := range tree {
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render(fmt.Sprintf("        [%d]:", k)),
					valueStyle.Render(scriptHex),
				)
			}
		}
	}

	// VTXO Tree Expiry Field
	vtxoTreeExpiries, err := txutils.GetArkPsbtFields(p, inputIndex, txutils.VtxoTreeExpiryField)
	if err == nil && len(vtxoTreeExpiries) > 0 {
		if !hasAnyFields {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render("  ARK PSBT Fields:"),
			)
			hasAnyFields = true
		}
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("    VtxoTreeExpiry:"),
		)
		for j, expiry := range vtxoTreeExpiries {
			output += fmt.Sprintf("%s\n",
				subLabelStyle.Render(fmt.Sprintf("      [%d]:", j)),
			)
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("        Type:"),
				valueStyle.Render(formatRelativeLocktimeType(expiry.Type)),
			)
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("        Value:"),
				valueStyle.Render(fmt.Sprintf("%d", expiry.Value)),
			)
		}
	}

	return output
}
