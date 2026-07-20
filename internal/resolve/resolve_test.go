package resolve

import (
	"reflect"
	"testing"
)

var table = `
# legacy commands mapped during the monom migration
api/db/migrate.sh    db migrate
api/db/seed.sh       db seed
api/auth/login.sh    auth login
standalone.sh        deploy
`

func TestRunMultiWordKey(t *testing.T) {
	got, err := Run(table, []string{"db", "migrate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "api/db/migrate.sh" {
		t.Errorf("got %q, want %q", got, "api/db/migrate.sh")
	}
}

func TestRunSingleWordKey(t *testing.T) {
	got, err := Run(table, []string{"deploy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "standalone.sh" {
		t.Errorf("got %q, want %q", got, "standalone.sh")
	}
}

func TestRunNoWordsReturnsError(t *testing.T) {
	_, err := Run(table, nil)
	if err == nil {
		t.Fatal("expected error for empty words")
	}
	if err.Error() != "no command words given" {
		t.Errorf("got error %q, want %q", err.Error(), "no command words given")
	}
}

func TestRunMissPassesWordsThrough(t *testing.T) {
	got, err := Run(table, []string{"db", "drop"})
	if err != nil {
		t.Fatalf("a miss must not be an error, got: %v", err)
	}
	if got != "db drop" {
		t.Errorf("miss must pass the words through: got %q, want %q", got, "db drop")
	}
}

func TestRunKeyPrefixDoesNotMatch(t *testing.T) {
	if got, _ := Run(table, []string{"db"}); got != "db" {
		t.Errorf("key prefix must pass through, not match a longer key: got %q", got)
	}
	if got, _ := Run(table, []string{"db", "migrate", "now"}); got != "db migrate now" {
		t.Errorf("longer words must pass through, not match a shorter key: got %q", got)
	}
}

func TestRunFirstMatchWins(t *testing.T) {
	dup := "first.sh   db migrate\nsecond.sh  db migrate"
	got, err := Run(dup, []string{"db", "migrate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "first.sh" {
		t.Errorf("got %q, want first entry %q", got, "first.sh")
	}
}

func TestRunIgnoresWhitespaceVariations(t *testing.T) {
	misaligned := "  api/db/migrate.sh \t db  \t migrate  "
	got, err := Run(misaligned, []string{"db", "migrate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "api/db/migrate.sh" {
		t.Errorf("got %q, want %q", got, "api/db/migrate.sh")
	}
}

func TestRunSkipsValueOnlyLines(t *testing.T) {
	bare := "orphan.sh\nreal.sh  orphan.sh"
	got, err := Run(bare, []string{"orphan.sh"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "real.sh" {
		t.Errorf("bare value line must be skipped, not matched: got %q", got)
	}
}

func TestRunEmptyTablePassesWordsThrough(t *testing.T) {
	got, err := Run("", []string{"db", "migrate"})
	if err != nil {
		t.Fatalf("an empty table must not be an error, got: %v", err)
	}
	if got != "db migrate" {
		t.Errorf("empty table must pass the words through: got %q", got)
	}
}

func TestCompleteListsKeysSlashDelimitedInTableOrder(t *testing.T) {
	got := Complete(table)
	want := []string{"db/migrate", "db/seed", "auth/login", "deploy"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompleteSkipsCommentsBlankAndValueOnlyLines(t *testing.T) {
	got := Complete("\n# a comment\norphan.sh\napi/db/migrate.sh  db migrate\n")
	want := []string{"db/migrate"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompleteEmptyTableReturnsNothing(t *testing.T) {
	if got := Complete(""); len(got) != 0 {
		t.Errorf("empty table must yield no keys, got %v", got)
	}
}
