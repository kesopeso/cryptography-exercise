package assert

import "testing"

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
