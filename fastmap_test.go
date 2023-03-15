package fastmap

import (
	"strconv"
	"testing"
)

type animal struct {
	name string
}

func TestSetGetDelete(t *testing.T) {
	m := New(WithSize(8), WithLoadFactor(0.7))
	m.Set("a", 1)
	value, found := m.Get("a")
	if !found {
		t.Error("expected value to be found by key")
	}
	if value.(int) != 1 {
		t.Error("expected value = 1")
	}
	m.Delete("a")
	_, found = m.Get("a")
	if found {
		t.Error("expected value to not be found by key")
	}
}

func TestOverwrite(t *testing.T) {
	m := New()
	key := "1"
	cat := "cat"
	tiger := "tiger"

	m.Set(key, cat)
	m.Set(key, tiger)

	if m.Len() != 1 {
		t.Errorf("map should contain exactly one element but has %v items.", m.Len())
	}

	item, ok := m.Get(key) // Retrieve inserted element.
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
	if item != tiger {
		t.Error("wrong item returned.")
	}
}

func TestSet(t *testing.T) {
	m := New(WithSize(4))

	m.Set("4", "cat")
	m.Set("3", "cat")
	m.Set("2", "tiger")
	m.Set("1", "tiger")

	if m.Len() != 4 {
		t.Error("map should contain exactly 4 elements.")
	}
}

func TestSet2(t *testing.T) {
	h := New()
	for i := 1; i <= 10; i++ {
		h.Set(strconv.Itoa(i), i)
	}
	for i := 1; i <= 10; i++ {
		h.Delete(strconv.Itoa(i))
	}
	for i := 1; i <= 10; i++ {
		h.Set(strconv.Itoa(i), i)
	}
	for i := 1; i <= 10; i++ {
		id, ok := h.Get(strconv.Itoa(i))
		if !ok {
			t.Error("ok should be true for item stored within the map.")
		}
		if id != i {
			t.Error("item is not as expected.")
		}
	}
}

func TestGet(t *testing.T) {
	m := New()
	cat := "cat"
	key := "animal"

	_, ok := m.Get(key) // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}

	m.Set(key, cat)

	_, ok = m.Get("human") // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}

	value, ok := m.Get(key) // Retrieve inserted element.
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}

	if value != cat {
		t.Error("item was modified.")
	}
}

func TestDelete(t *testing.T) {
	m := New()
	cat := &animal{"cat"}
	tiger := &animal{"tiger"}

	m.Set("1", cat)
	m.Set("2", tiger)
	m.Delete("0")
	m.Delete("3")
	m.Delete("4")
	m.Delete("5")
	if m.Len() != 2 {
		t.Error("map should contain exactly two elements.")
	}
	m.Delete("1")
	m.Delete("2")
	m.Delete("1")

	if m.Len() != 0 {
		t.Error("map should be empty.")
	}

	_, ok := m.Get("1") // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}
}
