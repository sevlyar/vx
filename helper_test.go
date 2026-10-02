package vx

import "testing"

// wantNoError and wantError check an already-computed result; the call
// producing err runs once, outside the subtest, which is fine since Check
// and the schema constructors are pure.

func wantNoError(err error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		if err != nil {
			t.Errorf("error = %v, want nil", err)
		}
	}
}

func wantError(err error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		if err == nil {
			t.Error("error = nil, want error")
		}
	}
}

func wantPanic(f func()) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("did not panic")
			}
		}()
		f()
	}
}

func wantNoPanic(f func()) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panicked with %v, want no panic", r)
			}
		}()
		f()
	}
}
