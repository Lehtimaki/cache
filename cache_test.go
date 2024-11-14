package cache

import (
	"testing"
)

func TestPutGetKey(t *testing.T) {
	kv := New()

	kv.Put("test", []byte(`hello`))

	v, ok := kv.Get("test")
	if !ok {
		t.Fatalf("unexpected missing key")
	}

	if string(v) != "hello" {
		t.Fatalf("expected %v, got: %v", "hello", string(v))
	}
}

func TestDelKey(t *testing.T) {
	kv := New()

	kv.Put("test", []byte(`hello`))
	if _, ok := kv.Get("test"); !ok {
		t.Fatalf("unexpected missing key")
	}

	kv.Del("test")
	if _, ok := kv.Get("test"); ok {
		t.Fatalf("expected key to be deleted")
	}
}
