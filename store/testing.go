package store

import (
	"fmt"
	"strings"
	"testing"
)

func TestStore(t *testing.T, cs string) (*Store, func(...string)) {
	t.Helper()

	config := NewConfig()

	s := New(config)
	if err := s.Open(cs); err != nil {
		t.Fatal(err)
	}

	return s, func(tables ...string) {
		if len(tables) > 0 {
			if _, err := s.db.Exec(fmt.Sprintf("TRUNCATE %s ", strings.Join(tables, ", "))); err != nil {
				t.Fatal(err)
			}
		}

		s.Close()
	}
}
