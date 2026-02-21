package command

import (
	"encoding/hex"
	"fmt"

	"github.com/arkade-os/arkd/pkg/ark-lib/asset"
)

// RunAssetPacketDecode decodes an asset packet from hex (raw packet or OP_RETURN script)
func RunAssetPacketDecode(input string) error {
	input = cleanHexInput(input)

	scriptBytes, err := hex.DecodeString(input)
	if err != nil {
		return fmt.Errorf("failed to decode hex input: %w", err)
	}

	var packet asset.Packet

	// Try to parse as full OP_RETURN script first
	packet, err = asset.NewPacketFromScript(scriptBytes)
	if err != nil {
		return fmt.Errorf("failed to decode asset packet: %w", err)
	}

	return printAssetPacket(packet)
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
	// Remove 0x prefix if present
	if len(input) >= 2 && input[:2] == "0x" {
		input = input[2:]
	}
	return input
}
