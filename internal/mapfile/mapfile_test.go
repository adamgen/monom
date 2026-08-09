package mapfile

import (
	"errors"
	"strings"
	"testing"
)

func parse(t *testing.T, src string) *Node {
	t.Helper()
	n, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse(%s): %v", src, err)
	}
	return n
}

func parseErr(t *testing.T, src, wantSubstr string) {
	t.Helper()
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatalf("Parse(%s): expected error containing %q, got nil", src, wantSubstr)
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("Parse(%s): error %q does not contain %q", src, err, wantSubstr)
	}
}

// --- node model ---

func TestStringValueIsCommandShorthand(t *testing.T) {
	n := parse(t, `{"deploy": "ops/full_deploy.sh"}`)
	cmd := n.Children["deploy"]
	if !cmd.IsCommand() || cmd.Run != "ops/full_deploy.sh" {
		t.Fatalf("expected command node with target, got %+v", cmd)
	}
}

func TestObjectWithRunKeyIsCommandNode(t *testing.T) {
	n := parse(t, `{"deploy": {"run": "ops/full_deploy.sh", "$note": "x"}}`)
	cmd := n.Children["deploy"]
	if !cmd.IsCommand() || cmd.Run != "ops/full_deploy.sh" {
		t.Fatalf("expected command node with target, got %+v", cmd)
	}
}

func TestObjectWithoutRunKeyIsCategory(t *testing.T) {
	n := parse(t, `{"db": {"migrate": "s/m.sh"}}`)
	cat := n.Children["db"]
	if cat.IsCommand() {
		t.Fatalf("expected category node, got command %+v", cat)
	}
	if !cat.Children["migrate"].IsCommand() {
		t.Fatalf("expected nested command node")
	}
}

func TestEmptyCategoryIsAllowedAndYieldsNoCommands(t *testing.T) {
	n := parse(t, `{"db": {}}`)
	if got := Complete(n); len(got) != 0 {
		t.Fatalf("expected no commands, got %v", got)
	}
}

func TestNoteIsIgnoredEverywhere(t *testing.T) {
	n := parse(t, `{"$note": "top", "db": {"$note": "cat", "seed": "t/s.sh"}}`)
	if got := Complete(n); len(got) != 1 || got[0] != "db/seed" {
		t.Fatalf("expected [db/seed], got %v", got)
	}
}

// --- parse rejections ---

func TestRejectsDuplicateKeys(t *testing.T) {
	parseErr(t, `{"db": {"a": "x.sh", "a": "y.sh"}}`, "duplicate key")
}

func TestRejectsRunPlusChildren(t *testing.T) {
	parseErr(t, `{"db": {"run": "x.sh", "seed": "y.sh"}}`, "command or a category, not both")
}

func TestRejectsUnknownMetadataKey(t *testing.T) {
	parseErr(t, `{"$version": "1"}`, "only $note is recognized")
}

func TestRejectsReservedNamesAsChildren(t *testing.T) {
	for _, name := range []string{"complete", "pre-run", "post-run"} {
		parseErr(t, `{"db": {"`+name+`": "x.sh"}}`, "reserved")
	}
}

func TestRejectsTopLevelRun(t *testing.T) {
	parseErr(t, `{"run": "x.sh"}`, "top level must be a category")
}

func TestRejectsNonStringNonObjectValues(t *testing.T) {
	parseErr(t, `{"db": 3}`, "must be a string target or an object")
	parseErr(t, `{"db": ["x.sh"]}`, "must be a string target or an object")
	parseErr(t, `{"db": null}`, "must be a string target or an object")
}

func TestRejectsNonObjectTopLevel(t *testing.T) {
	parseErr(t, `["db"]`, "top level must be a JSON object")
}

func TestRejectsInvalidNames(t *testing.T) {
	parseErr(t, `{"db migrate": "x.sh"}`, "whitespace")
	parseErr(t, `{"db/migrate": "x.sh"}`, "slash")
	parseErr(t, `{"": "x.sh"}`, "empty")
}

func TestRejectsInvalidTargets(t *testing.T) {
	parseErr(t, `{"a": ""}`, "target is empty")
	parseErr(t, `{"a": "has space.sh"}`, "whitespace")
	parseErr(t, `{"a": "/abs/path.sh"}`, "absolute")
	parseErr(t, `{"a": "../outside.sh"}`, "escapes the project root")
	parseErr(t, `{"a": "x/../../outside.sh"}`, "escapes the project root")
}

func TestAcceptsDotDotThatStaysInsideRoot(t *testing.T) {
	parse(t, `{"a": "scripts/../tools/x.sh"}`)
}

// --- Complete ---

func TestCompleteIsSortedAndSlashJoined(t *testing.T) {
	n := parse(t, `{
		"release": "t/r.sh",
		"db": {"seed": "t/s.sh", "migrate": "t/m.sh"},
		"deploy": {"run": "o/d.sh"}
	}`)
	got := Complete(n)
	want := []string{"db/migrate", "db/seed", "deploy", "release"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

// --- Resolve ---

func testTree(t *testing.T) *Node {
	t.Helper()
	return parse(t, `{
		"db": {"migrate": "scripts/legacy/run_migrations.sh"},
		"deploy": {"run": "ops/full_deploy.sh"}
	}`)
}

func TestResolveCommandPrintsTargetAsTokens(t *testing.T) {
	out, err := Resolve(testTree(t), []string{"db", "migrate"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if out != "scripts legacy run_migrations.sh" {
		t.Fatalf("expected target tokens, got %q", out)
	}
}

func TestResolveCategoryReturnsGroupError(t *testing.T) {
	out, err := Resolve(testTree(t), []string{"db"})
	var ge *GroupError
	if !errors.As(err, &ge) {
		t.Fatalf("expected GroupError, got out=%q err=%v", out, err)
	}
	if ge.Path != "db" {
		t.Fatalf("expected group path db, got %q", ge.Path)
	}
	if out != "" {
		t.Fatalf("group outcome must carry no output, got %q", out)
	}
}

func TestResolveZeroWordsIsTheRootCategory(t *testing.T) {
	_, err := Resolve(testTree(t), nil)
	var ge *GroupError
	if !errors.As(err, &ge) {
		t.Fatalf("expected GroupError for zero words, got %v", err)
	}
}

func TestResolveUnknownWordsIsSilentNoMatch(t *testing.T) {
	out, err := Resolve(testTree(t), []string{"nope"})
	if out != "" || err != nil {
		t.Fatalf("expected silent no-match, got out=%q err=%v", out, err)
	}
}

func TestResolveWordsPastACommandIsSilentNoMatch(t *testing.T) {
	// Trailing command arguments are deferred; the map must not claim them.
	out, err := Resolve(testTree(t), []string{"deploy", "--force"})
	if out != "" || err != nil {
		t.Fatalf("expected silent no-match, got out=%q err=%v", out, err)
	}
}
