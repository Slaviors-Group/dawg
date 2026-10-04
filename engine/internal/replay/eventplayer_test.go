package replay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEventPlayerStreamsInteractiveEventsAndKeepsDiagnostics(t *testing.T) {
	nodeBinary, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to execute replay scripts")
	}
	directory := t.TempDir()
	scriptPath := filepath.Join(directory, "events.cjs")
	script := `process.stdout.write('{"protocol":"dawg.replay.v1","sequence":1,"type":"ready"}\n');
process.stdout.write('{"protocol":"dawg.replay.v1","sequence":2,"type":"state"}\n');
process.stderr.write('replay diagnostic\n');`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	var events []string
	player := EventPlayer{
		NodeBinary:  nodeBinary,
		ScriptPath:  scriptPath,
		Interactive: true,
		ReplayEventSink: func(event json.RawMessage) error {
			events = append(events, string(event))
			return nil
		},
	}
	outcome, err := player.Replay(context.Background(), directory)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two streamed events, got %#v", events)
	}
	if outcome.Output != "replay diagnostic\n" {
		t.Fatalf("unexpected replay diagnostics: %q", outcome.Output)
	}
}

func TestEventPlayerDrainsOutputBeforeWaiting(t *testing.T) {
	nodeBinary, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to execute replay scripts")
	}
	directory := t.TempDir()
	scriptPath := filepath.Join(directory, "buffered-events.cjs")
	script := `const fs = require("node:fs");
for (let sequence = 0; sequence < 256; sequence++) {
  fs.writeSync(1, JSON.stringify({sequence, payload: "x".repeat(1024)}) + "\n");
}
fs.writeSync(2, "final diagnostic\n");`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count := 0
	player := EventPlayer{
		NodeBinary:  nodeBinary,
		ScriptPath:  scriptPath,
		Interactive: true,
		ReplayEventSink: func(event json.RawMessage) error {
			var got struct {
				Sequence int `json:"sequence"`
			}
			if err := json.Unmarshal(event, &got); err != nil {
				return err
			}
			if got.Sequence != count {
				return fmt.Errorf("unexpected sequence: got %d, want %d", got.Sequence, count)
			}
			count++
			// Keep events buffered when the child exits to exercise pipe draining.
			time.Sleep(time.Millisecond)
			return nil
		},
	}
	outcome, err := player.Replay(ctx, directory)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if count != 256 || outcome.Output != "final diagnostic\n" {
		t.Fatalf("incomplete output: events=%d, diagnostics=%q", count, outcome.Output)
	}
}

func TestEventPlayerStopsChildWhenEventReaderFails(t *testing.T) {
	nodeBinary, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to execute replay scripts")
	}
	for _, scenario := range []struct {
		name      string
		event     string
		sinkError error
		wantError string
	}{
		{name: "invalid-event", event: "not-json", wantError: "invalid browser event"},
		{name: "sink-error", event: `{}`, sinkError: errors.New("sink unavailable"), wantError: "forward browser event: sink unavailable"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			directory := t.TempDir()
			scriptPath := filepath.Join(directory, "reader-error.cjs")
			event, err := json.Marshal(scenario.event + "\n")
			if err != nil {
				t.Fatal(err)
			}
			script := "process.stdout.write(" + string(event) + "); setInterval(() => {}, 1000);"
			if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			player := EventPlayer{
				NodeBinary:      nodeBinary,
				ScriptPath:      scriptPath,
				Interactive:     true,
				ReplayEventSink: func(json.RawMessage) error { return scenario.sinkError },
			}
			_, err = player.Replay(ctx, directory)
			if err == nil || !strings.Contains(err.Error(), scenario.wantError) {
				t.Fatalf("expected %q, got %v", scenario.wantError, err)
			}
			if ctx.Err() != nil {
				t.Fatalf("reader failure did not stop child before deadline: %v", ctx.Err())
			}
		})
	}
}

func TestEventPlayerPassesReplayArguments(t *testing.T) {
	nodeBinary, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required to execute replay scripts")
	}

	for _, scenario := range []struct {
		name        string
		interactive bool
		colorScheme string
	}{
		{name: "noninteractive"},
		{name: "interactive", interactive: true},
		{name: "dark-color-scheme", colorScheme: "dark"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
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
				Interactive: scenario.interactive,
				ColorScheme: scenario.colorScheme,
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
			if scenario.colorScheme != "" {
				want = append(want, "--color-scheme", scenario.colorScheme)
			}
			if scenario.interactive {
				want = append(want, "--interactive")
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unexpected script arguments:\n got: %#v\nwant: %#v", got, want)
			}
		})
	}
}
