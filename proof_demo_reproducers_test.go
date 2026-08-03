// proof-demo showcase reproducers.
//
// This file lives ONLY on the `proof-demo` branch. Each test below is the
// reproducer for one intentionally re-introduced defect catalogued under
// proof/known-issues (KI-5 through KI-16). The reproducer marker on the line
// directly above each test binds that test to its known-issue so the
// proof-portal public dashboard can render the exact reproducer code next to
// the finding.
//
// Every test drives the real public API and ASSERTS the CORRECT behavior:
//   - KI-5..KI-14 are OPEN defects — while the bug is present on this branch
//     these tests are EXPECTED to fail (that is the point of the showcase).
//     The panic-prone cases (KI-5, KI-6, KI-8) recover so a present bug fails
//     gracefully with t.Errorf instead of crashing the whole suite.
//   - KI-15 and KI-16 are FIXED (the current code is already correct); their
//     reproducers are GREEN and pass today.
//
// DO NOT "fix" the failing open-bug tests — they document defects that must
// stay reproducible for the demo.
package jsonparser

import "testing"

// Reproduces: KI-5
func TestDemoKI5_TruncatedUnicodeEscapeOOB(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("KI-5: panic on truncated \\u escape %q: %v", `\u00`, r)
		}
	}()
	_, err := Unescape([]byte(`\u00`), nil)
	if err != MalformedStringEscapeError {
		t.Errorf("KI-5: expected MalformedStringEscapeError for %q, got %v", `\u00`, err)
	}
}

// Reproduces: KI-6
func TestDemoKI6_LoneTrailingBackslash(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("KI-6: panic on lone trailing backslash: %v", r)
		}
	}()
	_, err := Unescape([]byte(`\`), nil)
	if err != MalformedStringEscapeError {
		t.Errorf("KI-6: expected MalformedStringEscapeError for a lone backslash, got %v", err)
	}
}

// Reproduces: KI-7
func TestDemoKI7_NineteenDigitOverflow(t *testing.T) {
	v, err := ParseInt([]byte("9999999999999999999"))
	if err != OverflowIntegerError {
		t.Errorf("KI-7: expected OverflowIntegerError for 19-digit overflow, got v=%d err=%v", v, err)
	}
}

// Reproduces: KI-8
func TestDemoKI8_LoneHighSurrogateOOB(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("KI-8: panic on lone high surrogate %q: %v", `\uD834`, r)
		}
	}()
	res, err := Unescape([]byte(`\uD834`), nil)
	if err == nil && string(res) != "�" {
		t.Errorf("KI-8: expected U+FFFD or an error for a lone high surrogate, got %q err=%v", res, err)
	}
}

// Reproduces: KI-9
func TestDemoKI9_InvalidHexDigitG(t *testing.T) {
	res, err := Unescape([]byte(`\uG000`), nil)
	if err != MalformedStringEscapeError {
		t.Errorf("KI-9: expected MalformedStringEscapeError for %q, got %q err=%v", `\uG000`, res, err)
	}
}

// Reproduces: KI-10
func TestDemoKI10_EmptyArrayEach(t *testing.T) {
	count := 0
	_, err := ArrayEach([]byte(`[]`), func(value []byte, dataType ValueType, offset int, e error) {
		count++
	})
	if err != nil {
		t.Errorf("KI-10: expected nil error iterating an empty array, got %v", err)
	}
	if count != 0 {
		t.Errorf("KI-10: expected 0 callbacks for an empty array, got %d", count)
	}
}

// Reproduces: KI-11
func TestDemoKI11_BooleanTrailingGarbage(t *testing.T) {
	v, err := ParseBoolean([]byte("trueish"))
	if err != MalformedValueError {
		t.Errorf("KI-11: expected MalformedValueError for \"trueish\", got v=%v err=%v", v, err)
	}
}

// Reproduces: KI-12
func TestDemoKI12_EmptyInputErrorType(t *testing.T) {
	_, err := ArrayEach([]byte(""), func(value []byte, dataType ValueType, offset int, e error) {})
	if err != MalformedObjectError {
		t.Errorf("KI-12: expected MalformedObjectError for empty input, got %v", err)
	}
}

// Reproduces: KI-13
func TestDemoKI13_TabWhitespace(t *testing.T) {
	val, dataType, _, err := Get([]byte("{\"a\":\t123}"), "a")
	if err != nil {
		t.Errorf("KI-13: expected to read a value preceded by a tab, got err=%v", err)
	} else if string(val) != "123" {
		t.Errorf("KI-13: expected value 123 after a tab, got %q (type %v)", val, dataType)
	}
}

// Reproduces: KI-14
func TestDemoKI14_EscapedSlash(t *testing.T) {
	res, err := Unescape([]byte(`\/`), nil)
	if err != nil || string(res) != "/" {
		t.Errorf("KI-14: expected \"/\" for the \\/ escape, got %q err=%v", res, err)
	}
}

// Reproduces: KI-15
func TestDemoKI15_LoneLowSurrogateReplacement(t *testing.T) {
	res, err := Unescape([]byte(`\uDC00`), nil)
	if err != nil || string(res) != "�" {
		t.Errorf("KI-15: expected U+FFFD for a lone low surrogate, got %q err=%v", res, err)
	}
}

// Reproduces: KI-16
func TestDemoKI16_DeleteNoPath(t *testing.T) {
	res := Delete([]byte(`{"a":1}`), []string{}...)
	if len(res) != 0 {
		t.Errorf("KI-16: expected an empty result for a keyless Delete, got %q", res)
	}
}
