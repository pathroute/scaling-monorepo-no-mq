package db

import (
	"errors"
	"testing"
)

type row struct{ Name string }

func TestInsertGetList(t *testing.T) {
	tb := NewTable[row]()
	if err := tb.Insert("b", row{"bee"}); err != nil {
		t.Fatal(err)
	}
	tb.Insert("a", row{"ay"})
	if err := tb.Insert("a", row{"dup"}); !errors.Is(err, ErrExists) {
		t.Fatalf("dup insert: %v", err)
	}
	if v, err := tb.Get("a"); err != nil || v.Name != "ay" {
		t.Fatalf("get: %v %v", v, err)
	}
	if got := tb.List(nil); len(got) != 2 || got[0].Name != "ay" {
		t.Fatalf("list not ordered by id: %v", got)
	}
	if got := tb.List(func(r row) bool { return r.Name == "bee" }); len(got) != 1 {
		t.Fatalf("filtered list: %v", got)
	}
}

func TestTxRollsBackOnError(t *testing.T) {
	tb := NewTable[row]()
	tb.Insert("a", row{"ay"})
	boom := errors.New("boom")
	err := tb.Tx(func(tx *Tx[row]) error {
		tx.Update("a", row{"changed"})
		tx.Insert("b", row{"bee"})
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatal(err)
	}
	if v, _ := tb.Get("a"); v.Name != "ay" {
		t.Fatalf("update leaked: %v", v)
	}
	if _, err := tb.Get("b"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("insert leaked: %v", err)
	}
}

func TestTxReadsOwnWrites(t *testing.T) {
	tb := NewTable[row]()
	tb.Insert("a", row{"ay"})
	err := tb.Tx(func(tx *Tx[row]) error {
		if err := tx.Delete("a"); err != nil {
			return err
		}
		if _, err := tx.Get("a"); !errors.Is(err, ErrNotFound) {
			t.Errorf("deleted row still visible: %v", err)
		}
		return tx.Insert("a", row{"again"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := tb.Get("a"); v.Name != "again" {
		t.Fatalf("got %v", v)
	}
}

func TestNewIDUnique(t *testing.T) {
	if NewID() == NewID() {
		t.Fatal("ids collide")
	}
}
