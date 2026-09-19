package replay

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEventPlayerPassesInteractiveArgument(t *testing.T) {
	nodeBinary, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to execute replay scripts")
	}

	for _, interactive := range []bool{false, true} {
		t.Run(map[bool]string{false: "noninteractive", true: "interactive"}[interactive], func(t *testing.T) {
			sessionDirectory := t.TempDir()
			argumentsPath := filepath.Join(sessionDirectory, "arguments.json")
			scriptPath := filepath.Join(sessionDirectory, "capture-arguments.cjs")
			if err := os.WriteFile(scriptPath, []byte(`require("node:fs").writeFileSync(process.env.DAWG_EVENTPLAYER_ARGUMENTS, JSON.stringify(process.argv.slice(2)));`), 0o600); err != nil {
				t.Fatalf("write node script: %v", err)
			}
			t.Setenv("DAWG_EVENTPLAYER_ARGUMENTS", argumentsPath)

			player := EventPlayer{
				NodeBinary:  nodeBinary,
				ScriptPath:  scriptPath,
				Interactive: interactive,
			}
			if _, err := player.Replay(context.Background(), sessionDirectory); err != nil {
				t.Fatalf("replay: %v", err)
			}

			contents, err := os.ReadFile(argumentsPath)
			if err != nil {
				t.Fatalf("read captured arguments: %v", err)
			}
			var got []string
			if err := json.Unmarshal(contents, &got); err != nil {
				t.Fatalf("decode captured arguments: %v", err)
			}
			want := []string{
				"--rrweb-input", filepath.Join(sessionDirectory, "traces", "rrweb.jsonl"),
				"--screenshot-output", filepath.Join(sessionDirectory, "outcome", "screenshot.png"),
			}
			if interactive {
				want = append(want, "--interactive")
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unexpected script arguments:\n got: %#v\nwant: %#v", got, want)
			}
		})
	}
}
