package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Slaviors-Group/dawg/engine/internal/artifact"
	"github.com/Slaviors-Group/dawg/engine/internal/dawgenv"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/Slaviors-Group/dawg/engine/internal/packager"
	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
	"github.com/spf13/cobra"
)

const editorEventPrefix = "DAWG_EDITOR_EVENT\t"

type editorResult struct {
	Status string `json:"status"`
}

func newEditorCommand() *cobra.Command {
	var interactive bool
	command := &cobra.Command{Use: "editor <artifact-directory>", Short: "Open the Chromium review editor", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if !interactive {
			return fmt.Errorf("editor: --interactive is required")
		}
		return runEditor(command, arguments[0])
	}}
	command.Flags().BoolVar(&interactive, "interactive", false, "Run the interactive Chromium review editor")
	return command
}

func runEditor(command *cobra.Command, artifactPath string) error {
	source, err := artifact.Validate(artifactPath)
	if err != nil {
		return fmt.Errorf("editor: validate artifact: %w", err)
	}
	temporary, err := os.MkdirTemp("", "dawg-editor-*")
	if err != nil {
		return fmt.Errorf("editor: create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporary)
	if err := packager.Unpack(artifactPath, temporary); err != nil {
		return fmt.Errorf("editor: unpack artifact: %w", err)
	}

	draftPath, err := editorDraftPath(artifactPath)
	if err != nil {
		return err
	}
	review, hasDraft, err := readEditorDraft(draftPath, source)
	if err != nil {
		return err
	}
	if !hasDraft && source.Review != nil {
		review = artifact.NewReview(source, append([]manifest.ReviewFlag(nil), source.Review.Flags...), time.Now().UTC())
		hasDraft = true
	}
	nextRevision := 1
	if source.Review != nil {
		nextRevision = source.Review.Revision + 1
	}
	arguments := []string{
		dawgenv.ResolveScript("editor-browser.cjs"),
		"--rrweb-input", filepath.Join(temporary, "traces", "rrweb.jsonl"),
		"--default-artifact-title", defaultReviewArtifactTitle(source.Title, nextRevision),
	}
	if hasDraft {
		initialReview := filepath.Join(temporary, "review.json")
		contents, marshalErr := json.Marshal(review)
		if marshalErr != nil {
			return fmt.Errorf("editor: serialize review: %w", marshalErr)
		}
		if writeErr := os.WriteFile(initialReview, append(contents, '\n'), 0o600); writeErr != nil {
			return fmt.Errorf("editor: write review: %w", writeErr)
		}
		arguments = append(arguments, "--review-file", initialReview)
	} else {
		lineage := artifact.NewReview(source, []manifest.ReviewFlag{{ID: "00000000000000000000000000000000", Title: "placeholder", Category: "note", Severity: "info"}}, time.Now().UTC())
		arguments = append(arguments, "--root-artifact-id", lineage.RootArtifactID, "--parent-artifact-id", lineage.ParentArtifactID, "--revision", fmt.Sprintf("%d", lineage.Revision))
	}
	if colorScheme := editorColorScheme(artifactPath); colorScheme != "" {
		arguments = append(arguments, "--color-scheme", colorScheme)
	}

	ctx := command.Context()
	process := exec.CommandContext(ctx, dawgenv.ResolveNode(), arguments...)
	process.Stdin = os.Stdin
	process.Stderr = os.Stderr
	procutil.HideWindow(process)
	if browsersDirectory := dawgenv.ResolveBrowsersDir(); browsersDirectory != "" {
		process.Env = append(os.Environ(), "PLAYWRIGHT_BROWSERS_PATH="+browsersDirectory)
	}
	if chromiumPath := dawgenv.ResolveChromiumExecutable(); chromiumPath != "" {
		if process.Env == nil {
			process.Env = os.Environ()
		}
		process.Env = append(process.Env, "DAWG_CHROMIUM_EXECUTABLE_PATH="+chromiumPath)
	}
	stdout, err := process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("editor: open browser stdout: %w", err)
	}
	if err := process.Start(); err != nil {
		return fmt.Errorf("editor: start browser: %w", err)
	}

	forwardErr := forwardEditorEvents(stdout, command.ErrOrStderr(), artifactPath, source, draftPath)
	waitErr := process.Wait()
	if forwardErr != nil {
		return forwardErr
	}
	if waitErr != nil {
		return fmt.Errorf("editor: browser exited: %w", waitErr)
	}
	return json.NewEncoder(command.OutOrStdout()).Encode(editorResult{Status: "completed"})
}

func editorColorScheme(artifactPath string) string {
	evidence, err := packager.ReadDiagnostics(artifactPath)
	if err != nil {
		return ""
	}
	appearance, ok := evidence.Device["appearance"].(map[string]any)
	if !ok {
		return ""
	}
	colorScheme, _ := appearance["colorScheme"].(string)
	if colorScheme == "dark" || colorScheme == "light" {
		return colorScheme
	}
	return ""
}

func forwardEditorEvents(reader io.Reader, stderr io.Writer, artifactPath string, source manifest.Manifest, draftPath string) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 16<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		var event map[string]json.RawMessage
		if err := json.Unmarshal(line, &event); err != nil {
			return fmt.Errorf("editor: invalid browser event: %w", err)
		}
		eventType, err := editorEventType(event)
		if err != nil {
			return err
		}
		switch eventType {
		case "draftSaved":
			if err := saveEditorDraft(draftPath, source, event["flags"]); err != nil {
				return err
			}
		case "artifactSaved":
			metadata, err := publishEditorReview(artifactPath, source, event["review"], event["artifactTitle"])
			if err != nil {
				event["type"] = json.RawMessage(`"validationError"`)
				message, _ := json.Marshal(err.Error())
				event["message"] = message
			} else {
				_ = os.Remove(draftPath)
				path, _ := json.Marshal(metadata.Path)
				artifactID, _ := json.Marshal(metadata.ID)
				instanceID, _ := json.Marshal(metadata.InstanceID)
				artifactTitle, _ := json.Marshal(metadata.Title)
				event["artifactPath"] = path
				event["artifactId"] = artifactID
				event["instanceId"] = instanceID
				event["artifactTitle"] = artifactTitle
			}
		}
		contents, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("editor: serialize browser event: %w", err)
		}
		if _, err := fmt.Fprintf(stderr, "%s%s\n", editorEventPrefix, contents); err != nil {
			return fmt.Errorf("editor: forward browser event: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("editor: read browser events: %w", err)
	}
	return nil
}

func editorEventType(event map[string]json.RawMessage) (string, error) {
	var protocol, eventType string
	if err := json.Unmarshal(event["protocol"], &protocol); err != nil || protocol != "dawg.editor.v1" {
		return "", fmt.Errorf("editor: invalid browser event protocol")
	}
	if err := json.Unmarshal(event["type"], &eventType); err != nil || eventType == "" {
		return "", fmt.Errorf("editor: invalid browser event type")
	}
	return eventType, nil
}

func editorDraftPath(artifactPath string) (string, error) {
	absolutePath, err := filepath.Abs(artifactPath)
	if err != nil {
		return "", fmt.Errorf("editor: resolve artifact path: %w", err)
	}
	digest := sha256.Sum256([]byte(absolutePath))
	return defaultDawgDir("drafts", hex.EncodeToString(digest[:])+".json"), nil
}

func readEditorDraft(path string, source manifest.Manifest) (manifest.Review, bool, error) {
	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return manifest.Review{}, false, nil
	}
	if err != nil {
		return manifest.Review{}, false, fmt.Errorf("editor: read draft: %w", err)
	}
	var review manifest.Review
	if err := json.Unmarshal(contents, &review); err != nil {
		return manifest.Review{}, false, fmt.Errorf("editor: decode draft: %w", err)
	}
	if err := validateReviewLineage(source, review); err != nil {
		return manifest.Review{}, false, fmt.Errorf("editor: invalid draft: %w", err)
	}
	return review, true, nil
}

func saveEditorDraft(path string, source manifest.Manifest, rawFlags json.RawMessage) error {
	var flags []manifest.ReviewFlag
	if err := json.Unmarshal(rawFlags, &flags); err != nil {
		return fmt.Errorf("editor: decode draft flags: %w", err)
	}
	if len(flags) == 0 {
		return fmt.Errorf("editor: drafts require at least one review flag")
	}
	review := artifact.NewReview(source, flags, time.Now().UTC())
	if err := review.Validate(-1); err != nil {
		return fmt.Errorf("editor: validate draft: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("editor: create draft directory: %w", err)
	}
	contents, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return fmt.Errorf("editor: serialize draft: %w", err)
	}
	return os.WriteFile(path, append(contents, '\n'), 0o600)
}

func publishEditorReview(artifactPath string, source manifest.Manifest, rawReview, rawArtifactTitle json.RawMessage) (artifactMetadata, error) {
	var review manifest.Review
	if err := json.Unmarshal(rawReview, &review); err != nil {
		return artifactMetadata{}, fmt.Errorf("editor: decode review: %w", err)
	}
	if err := validateReviewLineage(source, review); err != nil {
		return artifactMetadata{}, fmt.Errorf("editor: validate review: %w", err)
	}
	artifactTitle, err := editorArtifactTitle(rawArtifactTitle, source.Title, review.Revision)
	if err != nil {
		return artifactMetadata{}, err
	}
	return publishReviewArtifact(artifactPath, source, review, artifactTitle)
}

func editorArtifactTitle(raw json.RawMessage, sourceTitle string, revision int) (string, error) {
	if len(raw) == 0 {
		return defaultReviewArtifactTitle(sourceTitle, revision), nil
	}
	var requested *string
	if err := json.Unmarshal(raw, &requested); err != nil || requested == nil {
		return "", fmt.Errorf("editor: artifactTitle must be a string")
	}
	title := strings.TrimSpace(*requested)
	if title == "" {
		return defaultReviewArtifactTitle(sourceTitle, revision), nil
	}
	if utf8.RuneCountInString(title) > 120 {
		return "", fmt.Errorf("editor: artifactTitle must be at most 120 characters")
	}
	return title, nil
}

func validateReviewLineage(source manifest.Manifest, review manifest.Review) error {
	expected := artifact.NewReview(source, append([]manifest.ReviewFlag(nil), review.Flags...), review.ReviewedAt)
	if review.FormatVersion != expected.FormatVersion || review.Kind != expected.Kind || review.RootArtifactID != expected.RootArtifactID || review.ParentArtifactID != expected.ParentArtifactID || review.Revision != expected.Revision {
		return fmt.Errorf("review lineage does not match source artifact")
	}
	return review.Validate(-1)
}
