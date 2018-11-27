package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestVersionMarshaling(t *testing.T) {
	ver, err := ParseVersion("1.0.0-rc.1+sha.5114f85")
	if err != nil {
		t.Fatal("Setup Error", err)
	}
	bs, err := ver.MarshalText()
	if err != nil {
		t.Fatal("MarshalText Error", err)
	}
	ver2 := &Version{}
	if err = ver2.UnmarshalText(bs); err != nil {
		t.Fatal("UnmarshalText Error", err)
	}
	if cmp := ver.Compare(*ver2); cmp != 0 {
		t.Fatal("Compare Error", err)
	}
}

type marshalTestJSON struct {
	Pointer1 *Version `json:"ptr1"`
	Pointer2 *Version `json:"ptr2,omitempty"`
	Value    Version  `json:"value"`
}

func TestMarshalUnMarshalJSON(t *testing.T) {
	v1, _ := ParseVersion("1.0")
	v2, _ := ParseVersion("1.2.3-rc.1")
	v3, _ := ParseVersion("1.2.3-rc.2+sha.0f1e2d")
	tests := []struct {
		v *marshalTestJSON
		e []byte
	}{
		{v: &marshalTestJSON{Pointer1: &v1, Pointer2: &v2, Value: v3}, e: []byte(`{"ptr1":"1.0.0","ptr2":"1.2.3-rc.1","value":"1.2.3-rc.2+sha.0f1e2d"}`)},
		{v: &marshalTestJSON{Pointer1: &v1, Pointer2: nil, Value: v3}, e: []byte(`{"ptr1":"1.0.0","value":"1.2.3-rc.2+sha.0f1e2d"}`)},
		{v: &marshalTestJSON{Pointer1: nil, Pointer2: &v2, Value: v3}, e: []byte(`{"ptr1":null,"ptr2":"1.2.3-rc.1","value":"1.2.3-rc.2+sha.0f1e2d"}`)},
		{v: &marshalTestJSON{Pointer1: &v1, Pointer2: &v2}, e: []byte(`{"ptr1":"1.0.0","ptr2":"1.2.3-rc.1","value":"0.0.0"}`)},
		{v: &marshalTestJSON{}, e: []byte(`{"ptr1":null,"value":"0.0.0"}`)},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			bs, err := json.Marshal(test.v)
			t.Logf("%s", string(bs))
			if err != nil {
				t.Fatal("unexpected marshal error")
			}
			if !bytes.Equal(test.e, bs) {
				t.Fatalf("unexpected json body. Expected\n%s\nActual\n%s", string(test.e), string(bs))
			}
			var v marshalTestJSON
			if err = json.Unmarshal(bs, &v); err != nil {
				t.Fatal("unexpected unmarshal error")
			}
			if v.Value.Compare(test.v.Value) != 0 {
				t.Fatalf("unexpected unmarshal Value. Expected\n%v\nActual\n%v", test.v.Value, v.Value)
			}

			if v.Pointer1 != nil && test.v.Pointer1 != nil {
				if v.Pointer1.Compare(*test.v.Pointer1) != 0 {
					t.Fatalf("unexpected unmarshal Pointer1. Expected\n%v\nActual\n%v", test.v.Pointer1, v.Pointer1)
				}
			} else if v.Pointer1 != test.v.Pointer1 { // fail unless both are nil
				t.Fatalf("unexpected nil `ptr1`")
			}
			if v.Pointer2 != nil && test.v.Pointer2 != nil {
				if v.Pointer2.Compare(*test.v.Pointer2) != 0 {
					t.Fatalf("unexpected unmarshal Pointer1. Expected\n%v\nActual\n%v", test.v.Pointer2, v.Pointer2)
				}
			} else if v.Pointer2 != test.v.Pointer2 { // fail unless both are nil
				t.Fatalf("unexpected nil `ptr2`")
			}
		})
	}
}

func TestVersionFormatter(t *testing.T) {
	vstr := "1.0.0-rc.1+sha.5114f85"
	ver, err := ParseVersion(vstr)
	if err != nil {
		t.Fatal("Setup Error", err)
	}
	if f := fmt.Sprint(ver); f != vstr {
		t.Fatalf("String() => '%s', expected '%s'", f, vstr)
	}
	tests := []struct {
		Format   string
		Expected string
	}{
		{"%s", "1.0.0-rc.1+sha.5114f85"},
		{"%-s", "1.0.0-rc.1"},
		{"%+4s", "1.0.0.0-rc.1+sha.5114f85"},
		{"%-q", `"1.0.0-rc.1"`},
		{"%q", `"1.0.0-rc.1+sha.5114f85"`},
	}
	for _, test := range tests {
		t.Run(test.Format, func(t *testing.T) {
			f := fmt.Sprintf(test.Format, ver)
			if f != test.Expected {
				t.Errorf("got '%s', expected '%s", f, test.Expected)
			}
			t.Logf(f)
		})
	}
}

func TestNumbersFormatter(t *testing.T) {
	var n0 Numbers

	n1 := Numbers{Number(1), Number(2)}
	n1ex := "1.2.0"

	if f := n1.String(); f != n1ex {
		t.Errorf("got '%s', expected '%s", f, n1ex)
	}

	var e1 Elements
	e1ex := "1.2"
	for _, n := range n1 {
		e1 = append(e1, n)
	}
	if f := e1.String(); f != e1ex {
		t.Errorf("got '%s', expected '%s", f, e1ex)
	}

	tests := []struct {
		Numbers  Numbers
		Format   string
		Expected string
	}{
		{n0, "%s", ""},
		{n0, "%3s", "0.0.0"},
		{n1, "%s", "1.2"},
		{n1, "%3s", "1.2.0"},
		{n1, "%4s", "1.2.0.0"},
	}
	for _, test := range tests {
		t.Run(test.Format, func(t *testing.T) {
			f := fmt.Sprintf(test.Format, test.Numbers)
			if f != test.Expected {
				t.Errorf("got '%s', expected '%s", f, test.Expected)
			}
			t.Logf(f)
		})
	}
}

func TestVersionStringer(t *testing.T) {
	tests := []string{
		"1.0.0",
		"1.0.0-rc.1",
		"1.0.0-rc.1+sha.5114f85",
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("ex.%d", i), func(t *testing.T) {
			v1, err := ParseVersion(test)
			if err != nil {
				t.Fatalf("parse failure on '%s'", test)
			}
			if v1.String() != test {
				t.Fatalf("String() => '%s' but should be '%s'", v1.String(), test)
			}
		})
	}
}

func TestInvalidSemanticVersions(t *testing.T) {
	tests := []string{
		"",
		"1.2",
		"1.2.3.4",
		"1.-1.3",
		"1.2.01",
		"1.2.00",
		"1.2.3-rc.01",
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("ex.%d", i), func(t *testing.T) {
			_, err := ParseSemanticVersion(test)
			if err != ErrInvalidVersion {
				t.Fatalf("expected parse failure on '%s'", test)
			}
		})
	}
}

func TestInvalidVersions(t *testing.T) {
	tests := []string{
		"",
		"1.2.a",
		"1.-1.3",
		"1.2.01",
		"1.2.00",
		"1.2.3-rc.01",
		"1.2.3-rc.1+01",
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("ex.%d", i), func(t *testing.T) {
			_, err := ParseVersion(test)
			if err != ErrInvalidVersion {
				t.Fatalf("expected parse failure on '%s'", test)
			}
		})
	}
}

func TestEqualityRules(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		// equal if all version components match
		{"1.0.0", "1.0.0", 0},
		// equal if all version components match (zero padded)
		{"1", "1.0.0", 0},
		{"1.0", "1.0.0", 0},
		{"1", "1.0.0", 0},
		{"1.0.0", "1.0", 0},
		// equality is not affected by build Metadata
		{"1.0+foo", "1.0.0+bar", 0},
		// equal if pre-release information matches (ignoring build metadata)
		{"1.0-foo", "1.0.0-foo", 0},
		{"1.0-foo.1", "1.0.0-foo.1", 0},
		{"1.0-foo.a", "1.0.0-foo.a", 0},
		{"1.0-foo.a+zzz", "1.0.0-foo.a+aaa", 0},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("ex.%d", i), func(t *testing.T) {
			v1, err := ParseVersion(test.v1)
			if err != nil {
				t.Fatal(err)
			}
			v2, err := ParseVersion(test.v2)
			if err != nil {
				t.Fatal(err)
			}
			if c := v1.Compare(v2); c != test.expected {
				t.Fatalf("%v compare %v expected %d but was %d", test.v1, test.v2, test.expected, c)
			}
		})
	}
}

func TestPrecedenceRules(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		// Adapted from https://github.com/rbarrois/python-semanticversion/blob/master/tests/test_spec.py
		// SPEC:
		// Precedence is determined by the first difference when comparing from
		// left to right as follows: Major, minor, and patch versions are always
		// compared numerically.
		// Example: 1.0.0 < 2.0.0 < 2.1.0 < 2.1.1
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "2.1.0", -1},
		{"2.1.0", "2.1.1", -1},

		// SPEC:
		// When major, minor, and patch are equal, a pre-release version has
		// lower precedence than a normal version.
		// Example: 1.0.0-alpha < 1.0.0
		{"1.0.0-alpha", "1.0.0", -1},

		// SPEC:
		// Precedence for two pre-release versions with the same major, minor,
		// and patch version MUST be determined by comparing each dot separated
		// identifier from left to right until a difference is found as follows:
		// identifiers consisting of only digits are compared numerically
		{"1.0.0-1", "1.0.0-2", -1},

		// and identifiers with letters or hyphens are compared lexically in
		// ASCII sort order.
		{"1.0.0-aa", "1.0.0-ab", -1},

		// Numeric identifiers always have lower precedence than
		// non-numeric identifiers.
		{"1.0.0-9", "1.0.0-a", -1},

		// A larger set of pre-release fields has a higher precedence than a
		// smaller set, if all of the preceding identifiers are equal.
		{"1.0.0-a.b.c", "1.0.0-a.b.c.0", -1},

		// Example: 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-beta.2", -1},
		{"1.0.0-beta.2", "1.0.0-beta.11", -1},
		{"1.0.0-beta.11", "1.0.0-rc.1", -1},
		{"1.0.0-rc.1", "1.0.0", -1},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("ex.%d", i), func(t *testing.T) {
			v1, err := ParseSemanticVersion(test.v1)
			if err != nil {
				t.Fatal(err)
			}
			v2, err := ParseSemanticVersion(test.v2)
			if err != nil {
				t.Fatal(err)
			}
			if c := v1.Compare(v2); c != test.expected {
				t.Fatalf("%v compare %v expected %d but was %d", test.v1, test.v2, test.expected, c)
			}
			if c := v2.Compare(v1); c != -test.expected {
				t.Fatalf("%v compare %v expected %d but was %d", test.v2, test.v1, -test.expected, c)
			}
		})
	}
}

func ExampleVersion_Manual() {
	ver := Version{
		Number:        Numbers{Number(1), Number(2), Number(3)},
		PreRelease:    Elements{TextElement("rc"), Number(1)},
		BuildMetadata: Elements{TextElement("sha"), TextElement("a0b1c2d")},
	}
	fmt.Println(ver)
	// Output: 1.2.3-rc.1+sha.a0b1c2d
}

func ExampleParseSemanticVersion() {
	semver, _ := ParseSemanticVersion("1.2.3-rc.1+sha.a0b1c2d")
	fmt.Println(semver)
	// Output: 1.2.3-rc.1+sha.a0b1c2d
}

func ExampleParseVersion() {
	semver, _ := ParseVersion("1.0.0.1")
	fmt.Println(semver)
	// Output: 1.0.0.1
}
