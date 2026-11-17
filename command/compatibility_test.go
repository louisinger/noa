package command

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/arkade-os/arkd/pkg/ark-lib/asset"
	"github.com/arkade-os/arkd/pkg/ark-lib/txutils"
	"github.com/btcsuite/btcd/psbt/v2"
	"github.com/btcsuite/btcd/wire/v2"
)

// Asset issuance fixture from the pre-upgrade Arkd library, with a supply of
// 21,000,000 and a control asset referencing group zero.
const legacyAssetScript = "6a1241524b0001020200000001010000c0de810a"
const rawAssetPacket = "01020200000001010000c0de810a"
const extensionAssetScript = "6a1341524b000e01020200000001010000c0de810a"

func TestAssetPacketFormats(t *testing.T) {
	for _, input := range []string{legacyAssetScript, rawAssetPacket, extensionAssetScript} {
		t.Run(input, func(t *testing.T) {
			output, err := captureOutput(t, func() error { return RunAssetPacketDecode(input) })
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"Asset Packet (1 groups)", "issuance", "ByGroup: 0", "vout:0 amount:21000000"} {
				if !strings.Contains(output, want) {
					t.Errorf("missing %q in %s", want, output)
				}
			}
		})
	}
}

func TestInvalidAssetPackets(t *testing.T) {
	// Empty, truncated, unknown packet type, and a valid extension without assets.
	for _, input := range []string{"", "6a", "6a1341524b000d01", "6a0641524b020100", "6a1241524b0001020200000001010000c0de81"} {
		t.Run(input, func(t *testing.T) {
			_, err := captureOutput(t, func() error { return RunAssetPacketDecode(input) })
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestRawAssetPacketWithOPReturnCount(t *testing.T) {
	raw, err := hex.DecodeString(rawAssetPacket)
	if err != nil {
		t.Fatal(err)
	}
	one, err := asset.NewPacketFromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	packet := make(asset.Packet, 0x6a)
	for i := range packet {
		packet[i] = one[0]
	}
	serialized, err := packet.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	output, err := captureOutput(t, func() error { return RunAssetPacketDecode(hex.EncodeToString(serialized)) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "Asset Packet (106 groups)") {
		t.Fatalf("raw group count mistaken for OP_RETURN: %s", output)
	}
}

func TestPsbtEncodingsAndAssetOutputs(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{Index: 1}, nil, nil))
	for _, scriptHex := range []string{legacyAssetScript, extensionAssetScript} {
		pkScript, err := hex.DecodeString(scriptHex)
		if err != nil {
			t.Fatal(err)
		}
		tx.AddTxOut(wire.NewTxOut(0, pkScript))
	}
	packet, err := psbt.NewFromUnsignedTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err := txutils.SetArkPsbtField(packet, 0, txutils.ConditionWitnessField, wire.TxWitness{{0x01, 0x02}}); err != nil {
		t.Fatal(err)
	}
	b64, err := packet.B64Encode()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := packet.Serialize(&buf); err != nil {
		t.Fatal(err)
	}
	var previous string
	for _, input := range []string{b64, hex.EncodeToString(buf.Bytes())} {
		output, err := captureOutput(t, func() error { return RunPsbtDecode(input) })
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"Inputs (1)", "Outputs (2)", "ConditionWitness:", "0102"} {
			if !strings.Contains(output, want) {
				t.Errorf("missing %q in %s", want, output)
			}
		}
		if strings.Count(output, "Asset Packet (1 groups)") != 2 {
			t.Fatalf("expected both legacy and current asset outputs: %s", output)
		}
		if previous != "" && output != previous {
			t.Fatal("base64 and hex must decode identically")
		}
		previous = output
	}
}

func TestExistingCommands(t *testing.T) {
	const multisig = "2079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798ac"
	const legacyTree = "01c022" + multisig
	const tree = "00c022" + multisig
	for _, tc := range []struct {
		name string
		run  func() error
		want string
	}{
		{"address", func() error {
			return RunAddress("tark1qqellv77udfmr20tun8dvju5vgudpf9vxe8jwhthrkn26fz96pawqfdy8nk05rsmrf8h94j26905e7n6sng8y059z8ykn2j5xcuw4xt846qj6x")
		}, "0233ffb3dee353b1a9ebe4ced64b946238d0a4ac364f275d771da6ad2445d07ae0"},
		{"note", func() error { return RunNoteFromTxid(strings.Repeat("0", 63) + "1") }, "512052d3ecc5ca62828934a0c28eff1facd2892ddd9d485c28979c98416ea4a23122"},
		{"taptree encode", func() error { return RunTaptreeEncode([]string{multisig}) }, tree},
		{"taptree decode", func() error { return RunTaptreeDecode(tree) }, "512089b13f1de2d5bc700695813283363c8c3464dd9597994c072ca5e4df022c3947"},
		{"legacy taptree decode", func() error { return RunTaptreeDecode(legacyTree) }, "512089b13f1de2d5bc700695813283363c8c3464dd9597994c072ca5e4df022c3947"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := captureOutput(t, tc.run)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output, tc.want) {
				t.Fatalf("missing %q in %s", tc.want, output)
			}
		})
	}
}
