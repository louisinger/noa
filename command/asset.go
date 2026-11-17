package command

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/arkade-os/arkd/pkg/ark-lib/asset"
	"github.com/arkade-os/arkd/pkg/ark-lib/extension"
	"github.com/btcsuite/btcd/txscript/v2"
)

// RunAssetPacketDecode decodes an asset packet from hex (raw packet or OP_RETURN script)
func RunAssetPacketDecode(input string) error {
	input = cleanHexInput(input)

	scriptBytes, err := hex.DecodeString(input)
	if err != nil {
		return fmt.Errorf("failed to decode hex input: %w", err)
	}

	var packet asset.Packet

	if len(scriptBytes) > 0 && scriptBytes[0] == txscript.OP_RETURN {
		packet, err = decodeAssetPacketScript(scriptBytes)
		if err != nil {
			// A raw packet's group count can also begin with byte 0x6a.
			if rawPacket, rawErr := asset.NewPacketFromBytes(scriptBytes); rawErr == nil {
				packet, err = rawPacket, nil
			}
		}
	} else {
		packet, err = asset.NewPacketFromBytes(scriptBytes)
	}
	if err != nil {
		return fmt.Errorf("failed to decode asset packet: %w", err)
	}

	return printAssetPacket(packet)
}

// decodeAssetPacketScript accepts current ARK extensions and the legacy
// OP_RETURN ARK + asset marker + packet encoding used by earlier Arkd versions.
func decodeAssetPacketScript(script []byte) (asset.Packet, error) {
	ext, extensionErr := extension.NewExtensionFromBytes(script)
	if extensionErr == nil {
		if packet := ext.GetAssetPacket(); packet != nil {
			return packet, nil
		}
		return nil, fmt.Errorf("extension contains no asset packet")
	}

	tokenizer := txscript.MakeScriptTokenizer(0, script)
	if !tokenizer.Next() || tokenizer.Opcode() != txscript.OP_RETURN {
		return nil, extensionErr
	}
	var payload []byte
	for tokenizer.Next() {
		if len(tokenizer.Data()) == 0 {
			return nil, extensionErr
		}
		payload = append(payload, tokenizer.Data()...)
	}
	if err := tokenizer.Err(); err != nil {
		return nil, err
	}
	legacyPrefix := []byte{'A', 'R', 'K', asset.PacketType}
	if bytes.HasPrefix(payload, legacyPrefix) {
		if packet, err := asset.NewPacketFromBytes(payload[len(legacyPrefix):]); err == nil {
			return packet, nil
		}
	}
	return nil, extensionErr
}

// printAssetPacket outputs a formatted asset packet
func printAssetPacket(packet asset.Packet) error {
	var output string

	output += fmt.Sprintf("\n%s\n",
		sectionStyle.Render(fmt.Sprintf("Asset Packet (%d groups):", len(packet))),
	)

	for i, group := range packet {
		output += formatAssetGroup(i, &group)
	}

	fmt.Print(output)
	return nil
}

// formatAssetGroup formats a single asset group
func formatAssetGroup(index int, group *asset.AssetGroup) string {
	var output string

	output += fmt.Sprintf("\n%s\n",
		subLabelStyle.Render(fmt.Sprintf("[Group %d]:", index)),
	)

	// Asset ID
	if group.AssetId != nil {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  AssetId:"),
			valueStyle.Render(fmt.Sprintf("%s:%d", group.AssetId.Txid.String(), group.AssetId.Index)),
		)
	} else {
		output += fmt.Sprintf("%s%s\n",
			subLabelStyle.Render("  AssetId:"),
			valueStyle.Render("<new issuance>"),
		)
	}

	// Type
	groupType := "transfer"
	if group.IsIssuance() {
		groupType = "issuance"
	} else if group.IsReissuance() {
		groupType = "reissuance"
	}
	output += fmt.Sprintf("%s%s\n",
		subLabelStyle.Render("  Type:"),
		valueStyle.Render(groupType),
	)

	// Control Asset
	if group.ControlAsset != nil {
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("  ControlAsset:"),
		)
		switch group.ControlAsset.Type {
		case asset.AssetRefByID:
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("    ById:"),
				valueStyle.Render(fmt.Sprintf("%s:%d", group.ControlAsset.AssetId.Txid.String(), group.ControlAsset.AssetId.Index)),
			)
		case asset.AssetRefByGroup:
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render("    ByGroup:"),
				valueStyle.Render(fmt.Sprintf("%d", group.ControlAsset.GroupIndex)),
			)
		}
	}

	// Metadata
	if len(group.Metadata) > 0 {
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("  Metadata:"),
		)
		for _, md := range group.Metadata {
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render(fmt.Sprintf("    %s:", string(md.Key))),
				valueStyle.Render(string(md.Value)),
			)
		}
	}

	// Inputs
	if len(group.Inputs) > 0 {
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("  Inputs:"),
		)
		for j, in := range group.Inputs {
			switch in.Type {
			case asset.AssetInputTypeLocal:
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render(fmt.Sprintf("    [%d]:", j)),
					valueStyle.Render(fmt.Sprintf("local vin:%d amount:%d", in.Vin, in.Amount)),
				)
			case asset.AssetInputTypeIntent:
				output += fmt.Sprintf("%s%s\n",
					subLabelStyle.Render(fmt.Sprintf("    [%d]:", j)),
					valueStyle.Render(fmt.Sprintf("intent txid:%s vout:%d amount:%d", in.Txid.String(), in.Vin, in.Amount)),
				)
			}
		}
	}

	// Outputs
	if len(group.Outputs) > 0 {
		output += fmt.Sprintf("%s\n",
			subLabelStyle.Render("  Outputs:"),
		)
		for j, out := range group.Outputs {
			output += fmt.Sprintf("%s%s\n",
				subLabelStyle.Render(fmt.Sprintf("    [%d]:", j)),
				valueStyle.Render(fmt.Sprintf("vout:%d amount:%d", out.Vout, out.Amount)),
			)
		}
	}

	return output
}

// cleanHexInput removes common prefixes and whitespace from hex input
func cleanHexInput(input string) string {
	input = strings.Join(strings.Fields(input), "")
	// Remove 0x prefix if present
	if len(input) >= 2 && input[:2] == "0x" {
		input = input[2:]
	}
	return input
}
