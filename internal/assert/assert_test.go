package assert

import (
	"errors"
	"testing"
)

func TestTrue_DoesNotPanicWhenConditionIsTrue(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	True(true, "should not panic")
}

func TestTrue_PanicsWhenConditionIsFalse(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		want := "condition failed"
		if msg != want {
			t.Errorf("panic message = %q, want %q", msg, want)
		}
	}()

	True(false, "condition failed")
}

func TestNoError_DoesNotPanicWhenNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	NoError(nil)
}

func TestNoError_PanicsWhenError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		want := "something went wrong"
		if msg != want {
			t.Errorf("panic message = %q, want %q", msg, want)
		}
	}()

	NoError(errors.New("something went wrong"))
}
