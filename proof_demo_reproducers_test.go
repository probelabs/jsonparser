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
	// Control (safe path): a well-formed \uXXXX escape decodes correctly and does
	// NOT trigger the defect — the impact is confined to the truncated-input case.
	if res, err := Unescape([]byte(`\u00`+`41`), nil); err != nil || string(res) != "A" {
		t.Errorf("KI-5 control: expected \"A\" for a well-formed \\u0041 escape, got %q err=%v", res, err)
	}
	// Direct-locus witness: decodeSingleUnicodeEscape (escape.go:88) is the mutated
	// function. With only 4 bytes it must reject (ok=false); the weakened `< 4`
	// guard instead indexes in[4]/in[5] out of bounds.
	if _, ok := decodeSingleUnicodeEscape([]byte(`\u00`)); ok {
		t.Errorf("KI-5: decodeSingleUnicodeEscape accepted a 4-byte truncated escape instead of rejecting it")
	}
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
	// Control (safe path): a well-formed 2-byte escape unescapes correctly and does
	// NOT trigger the defect — the impact is confined to the truncated 1-byte input.
	if inLen, outLen := unescapeToUTF8Config(DefaultConfig, []byte(`\n`), make([]byte, 4)); inLen != 2 || outLen != 1 {
		t.Errorf("KI-6 control: expected (2,1) for a well-formed \\n escape, got (%d,%d)", inLen, outLen)
	}
	// Direct-locus witness: unescapeToUTF8Config (escape.go:194) is the mutated
	// function. A lone backslash (1 byte) must return inLen=-1; the weakened `< 1`
	// guard instead indexes in[1] out of bounds.
	if inLen, _ := unescapeToUTF8Config(DefaultConfig, []byte(`\`), make([]byte, 4)); inLen != -1 {
		t.Errorf("KI-6: unescapeToUTF8Config accepted a lone trailing backslash (inLen=%d), expected -1", inLen)
	}
	_, err := Unescape([]byte(`\`), nil)
	if err != MalformedStringEscapeError {
		t.Errorf("KI-6: expected MalformedStringEscapeError for a lone backslash, got %v", err)
	}
}

// Reproduces: KI-7
func TestDemoKI7_NineteenDigitOverflow(t *testing.T) {
	// Control (safe path): an in-range integer parses correctly and does NOT
	// trigger the defect — the impact is confined to 19-digit values above MaxInt64.
	if cv, ok, overflow := parseInt([]byte("42")); !ok || overflow || cv != 42 {
		t.Errorf("KI-7 control: expected parseInt(\"42\")=(42,true,false), got (%d,%v,%v)", cv, ok, overflow)
	}
	// Direct-locus witness: parseInt (bytes.go:27) is the mutated function. A 19-digit
	// value above math.MaxInt64 must set overflow=true; the widened `< 20` fast-path
	// threshold instead wraps it silently with plain int64 arithmetic.
	if _, _, overflow := parseInt([]byte("9999999999999999999")); !overflow {
		t.Errorf("KI-7: parseInt did not flag 19-digit overflow (overflow=false) — value wrapped silently")
	}
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
	// Control (safe path): a well-formed high+low surrogate PAIR decodes to the
	// astral code point and does NOT trigger the defect — the impact is confined to
	// a LONE high surrogate at end-of-buffer.
	if r, n := decodeUnicodeEscape([]byte(`\uD8` + `34` + `\uDD` + `1E`)); n != 12 || r != rune(0x1D11E) {
		t.Errorf("KI-8 control: expected (U+1D11E, 12) for a valid surrogate pair, got (%U, %d)", r, n)
	}
	// Direct-locus witness: decodeUnicodeEscape (escape.go:146) is the mutated
	// function. A lone high surrogate (6 bytes) must consume 6 and substitute U+FFFD;
	// the weakened `len(in) < 6` guard instead indexes in[6]/in[7] out of bounds.
	if _, n := decodeUnicodeEscape([]byte(`\uD8` + `34`)); n != 6 {
		t.Errorf("KI-8: decodeUnicodeEscape consumed %d bytes for a lone high surrogate, expected 6 (U+FFFD)", n)
	}
	res, err := Unescape([]byte(`\uD834`), nil)
	if err == nil && string(res) != "�" {
		t.Errorf("KI-8: expected U+FFFD or an error for a lone high surrogate, got %q err=%v", res, err)
	}
}

// Reproduces: KI-9
func TestDemoKI9_InvalidHexDigitG(t *testing.T) {
	// Control (safe path): genuine hex digits decode correctly and a clearly
	// out-of-range byte is still rejected — the impact is confined to the single
	// off-by-one byte 'G' (71), one past 'F'.
	if h2I('A') != 10 || h2I('F') != 15 || h2I('Z') != badHex {
		t.Errorf("KI-9 control: expected h2I('A')=10, h2I('F')=15, h2I('Z')=badHex; got %d, %d, %d", h2I('A'), h2I('F'), h2I('Z'))
	}
	// Direct-locus witness: h2I (escape.go:71) is the mutated function. Byte 'G' (71)
	// is not a hex digit and must map to badHex; the widened `c <= 71` bound instead
	// decodes it as 16.
	if h2I('G') != badHex {
		t.Errorf("KI-9: h2I('G')=%d, expected badHex(%d) — 'G' wrongly accepted as a hex digit", h2I('G'), badHex)
	}
	res, err := Unescape([]byte(`\uG000`), nil)
	if err != MalformedStringEscapeError {
		t.Errorf("KI-9: expected MalformedStringEscapeError for %q, got %q err=%v", `\uG000`, res, err)
	}
}

// Reproduces: KI-10
func TestDemoKI10_EmptyArrayEach(t *testing.T) {
	// Control (safe path): a NON-empty array iterates correctly and does NOT trigger
	// the defect — the impact is confined to the empty-array `[]` sentinel case.
	ctrl := 0
	if _, err := ArrayEach([]byte(`[1,2]`), func(value []byte, dataType ValueType, offset int, e error) {
		ctrl++
	}); err != nil || ctrl != 2 {
		t.Errorf("KI-10 control: expected 2 callbacks and nil error for [1,2], got %d callbacks err=%v", ctrl, err)
	}
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
	// Control (safe path): the exact "true"/"false" literals parse correctly and do
	// NOT trigger the defect — the impact is confined to tokens with trailing bytes.
	if tv, err := ParseBoolean([]byte("true")); err != nil || tv != true {
		t.Errorf("KI-11 control: expected ParseBoolean(\"true\")=(true,nil), got (%v,%v)", tv, err)
	}
	if fv, err := ParseBoolean([]byte("false")); err != nil || fv != false {
		t.Errorf("KI-11 control: expected ParseBoolean(\"false\")=(false,nil), got (%v,%v)", fv, err)
	}
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
