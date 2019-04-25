package db

import (
	"fmt"
	"os"
	"testing"
)

func compareStrSlice(a, b []*Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i].Key != b[i].Key || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

func TestList(t *testing.T) {
	c, _ := NewClient("./test-data")
	defer c.Close()
	defer os.RemoveAll("./test-data")

	for i := 10; i < 30; i++ {
		c.Set(fmt.Sprintf("key%v", i), fmt.Sprintf("val%v", i))
	}
	_, total, _ := c.List("", 0, 0)
	if want := uint(20); total != want {
		t.Errorf("total want %v, but got %v", want, total)
	}

	vals, total, _ := c.List("", 0, 5)
	if want := uint(20); total != want {
		t.Errorf("total want %v, but got %v", want, total)
	}

	if want := []*Item{
		&Item{Key: "key10", Value: "val10"},
		&Item{Key: "key11", Value: "val11"},
		&Item{Key: "key12", Value: "val12"},
		&Item{Key: "key13", Value: "val13"},
		&Item{Key: "key14", Value: "val14"},
	}; !compareStrSlice(vals, want) {
		t.Errorf("want %v, but got %v", want, vals)
	}

	vals, total, _ = c.List("key2", 2, 5)
	if want := uint(10); total != want {
		t.Errorf("total want %v, but got %v", want, total)
	}

	if want := []*Item{
		&Item{Key: "key22", Value: "val22"},
		&Item{Key: "key23", Value: "val23"},
		&Item{Key: "key24", Value: "val24"},
		&Item{Key: "key25", Value: "val25"},
		&Item{Key: "key26", Value: "val26"},
	}; !compareStrSlice(vals, want) {
		t.Errorf("want %v, but got %v", want, vals)
	}
}
