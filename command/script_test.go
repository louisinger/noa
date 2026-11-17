package command

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Covenant supplied as a regression case for Arkade script decoding.
const arkadeCovenantHex = "00204c9c7c70cd5ab959852e878559b3ff876dbddb41bf353e69d9b5f0d50c86d6b4d4519dd5539dcd009d0020a064714c32fb8acf854954e989d3a6ba8dfd9dab0b85d6e54a22d1e9b2dab6ce000000537a547af2637651bb5152bb68527991630051bb5152bb687c7777519d002008faa95febe70e8ff9890eea03ab4bc7837bccdbddc99670059b386bdefcfd5400ef7791690020003112383e8e6e16b6811f40ce23fafe3833fae41872bd80d53a7d3bca70112500ef779169002008d811ac7202ae56edf51838561ffb24874974d7540eb5b0f7d3c20ecbfea6e800ef77916900200eb865588af0a3d60efb737586b5bd9dc0440ea487c2c52d84dd496ec47a4e2600ef779169002056124b9b25ef8cabfd9b6bccd244f5f0ec3ff3f894dfb8cdc6825d30b4407bfb00ef7791690057707c51f7696e000000537a547af2637651bb5152bb68527991630051bb5152bb687c7777519d6e515276009c630053795579ef779169670053795579ef69789d6878009c6376009c67006863707ce877916967707ce869767651ec7c00ec94707c949d00ec52799d686d6d6d51"

func TestRunScriptArkadeCovenant(t *testing.T) {
	output, err := captureRunScript(t, arkadeCovenantHex)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"arkade asm:",
		"4c9c7c70cd5ab959852e878559b3ff876dbddb41bf353e69d9b5f0d50c86d6b4",
		"OP_INSPECTNUMINPUTS 1 OP_NUMEQUALVERIFY OP_INSPECTNUMOUTPUTS 3 OP_NUMEQUALVERIFY OP_PUSHCURRENTINPUTINDEX 0 OP_NUMEQUALVERIFY",
		"OP_INSPECTINASSETLOOKUP",
		"OP_INSPECTOUTASSETLOOKUP",
		"OP_PUT",
		"OP_TUNNEL",
		"OP_FINDASSETGROUPBYASSETID",
		"OP_INSPECTASSETGROUPSUM",
		"OP_2DROP OP_2DROP OP_2DROP 1",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q: %s", want, output)
		}
	}
	_, arkadeOutput, ok := strings.Cut(output, "arkade asm:")
	if !ok {
		t.Fatal("missing Arkade ASM")
	}
	if strings.Contains(arkadeOutput, "OP_UNKNOWN") || strings.Contains(output, "Closure:") {
		t.Fatalf("unexpected opcode or closure output: %s", output)
	}
}

func TestRunScriptMultisigClosure(t *testing.T) {
	const pubkey = "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	for _, tc := range []struct{ prefix, closure string }{
		{"", "MultisigClosure"},
		{"029000b275", "CSVMultisigClosure"},
		{"029000b175", "CLTVMultisigClosure"},
		{"5169", "ConditionMultisigClosure"},
		{"5169029000b275", "ConditionCSVMultisigClosure"},
	} {
		t.Run(tc.closure, func(t *testing.T) {
			output, err := captureRunScript(t, tc.prefix+"20"+pubkey+"ac")
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{pubkey + " OP_CHECKSIG", "Closure:", tc.closure, "PubKeys:"} {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q: %s", want, output)
				}
			}
			if strings.Contains(output, "arkade asm:") {
				t.Fatalf("standard closure should retain its original output: %s", output)
			}
		})
	}
}

func TestRunScriptAmbiguousOpcode(t *testing.T) {
	output, err := captureRunScript(t, "b3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "asm: OP_NOP4") || !strings.Contains(output, "arkade asm: OP_MERKLEBRANCHROOT") {
		t.Fatalf("both opcode interpretations must be preserved: %s", output)
	}
}

func TestDisasmArkadeScriptPushData(t *testing.T) {
	// All four push encodings contain bytes that are Arkade opcodes outside a push.
	script, err := hex.DecodeString("02f7bb4c02f2ef4d0200ecd54e02000000e8cd00bb4f5160f7")
	if err != nil {
		t.Fatal(err)
	}
	got, err := disasmArkadeScript(script)
	if err != nil {
		t.Fatal(err)
	}
	const want = "f7bb f2ef ecd5 e8cd 0 OP_PUT -1 1 16 OP_TUNNEL"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRunScriptMalformed(t *testing.T) {
	for _, input := range []string{"zz", "0", "02ff", "4c", "4d01", "4e010000", "51f74c02ff"} {
		t.Run(input, func(t *testing.T) {
			output, err := captureRunScript(t, input)
			if err == nil {
				t.Fatal("expected malformed script error")
			}
			if output != "" {
				t.Fatalf("unexpected success output: %s", output)
			}
		})
	}
}

func captureRunScript(t *testing.T, scriptHex string) (string, error) {
	t.Helper()
	return captureOutput(t, func() error { return RunScript(scriptHex) })
}

func captureOutput(t *testing.T, run func() error) (string, error) {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), "stdout"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	original := os.Stdout
	os.Stdout = f
	defer func() { os.Stdout = original }()
	runErr := run()
	output, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(output), runErr
}
