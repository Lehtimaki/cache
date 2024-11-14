package cache

import (
	"context"
	"testing"
	"time"
)

func TestPutGetKey(t *testing.T) {
	kv := New(context.Background(), 0)

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
	kv := New(context.Background(), 0)

	kv.Put("test", []byte(`hello`))
	if _, ok := kv.Get("test"); !ok {
		t.Fatalf("unexpected missing key")
	}

	kv.Del("test")
	if _, ok := kv.Get("test"); ok {
		t.Fatalf("expected key to be deleted")
	}
}

func TestGarbageCollection(t *testing.T) {
	kv := New(context.Background(), time.Second)

	kv.Put("test", []byte(`hello`))

	if _, ok := kv.Get("test"); !ok {
		t.Fatalf("unexpected missing key")
	}

	time.Sleep((100 * time.Millisecond) + time.Second)

	if v, ok := kv.Get("test"); ok {
		t.Fatalf("expected key to have been garbage collected, found: %s", string(v))
	}
}
