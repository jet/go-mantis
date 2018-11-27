// Package version is a parsing library for versions that conform to a superset of the Semantic Version specification
// specified on https://semver.org.

package version // import "github.com/jet/go-mantis/version"

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ErrInvalidVersion is returned when the version could not be parsed
// or does not conform to the semantic version specification
var ErrInvalidVersion = errors.New("invalid version")

var versionRegexp = regexp.MustCompile(`^v?(?P<version>[0-9]+(?:[.][0-9]+)*)(?:-(?P<prerelease>[0-9A-Za-z-]+(?:[.][0-9A-Za-z-]+)*))?(?:[+](?P<build_metadata>[0-9A-Za-z.-]+))?$`)

// ParseSemanticVersion will parse a version string into Elements
// This is similar to ParseVersion but strictly enforces
// three components: Number = [Major, Minor, Patch]
func ParseSemanticVersion(s string) (sv Version, err error) {
	sv, err = ParseVersion(s)
	if err != nil {
		return
	}
	if len(sv.Number) != 3 {
		err = ErrInvalidVersion
		return
	}
	return
}

// ParseVersion will parse a version string into Elements
// Like `ParseSemanticVersion` except doesn't constrain the Number component;
// allows the `Number` part of the version to vary in length.
func ParseVersion(s string) (ver Version, err error) {
	m := versionRegexp.FindStringSubmatch(s)
	if len(m) == 0 {
		return Version{}, ErrInvalidVersion
	}
	for _, v := range strings.Split(m[1], ".") {
		c, e := ParseElement(v)
		if e != nil {
			return Version{}, e
		}
		nv, ok := c.(Number)
		if !ok {
			// the regexp match *should* make this impossible
			return Version{}, ErrInvalidVersion
		}
		ver.Number = append(ver.Number, nv)
	}
	if m[2] != "" {
		for _, v := range strings.Split(m[2], ".") {
			c, e := ParseElement(v)
			if e != nil {
				return Version{}, e
			}
			ver.PreRelease = append(ver.PreRelease, c)
		}
	}
	if m[3] != "" {
		for _, v := range strings.Split(m[3], ".") {
			c, e := ParseElement(v)
			if e != nil {
				return Version{}, e
			}
			ver.BuildMetadata = append(ver.BuildMetadata, c)
		}
	}
	return
}

type Numbers []Number

func (n Numbers) Compare(o Numbers) int {
	if len(o) > len(n) {
		return -o.Compare(n)
	}
	for i, a := range n {
		var b Number
		if len(o) > i {
			b = o[i]
		}
		if a != b {
			return a.Compare(b)
		}
	}
	return 0
}

// String
func (n Numbers) String() string {
	// Use semver semantics by default
	// which assumes 3 elements
	return fmt.Sprintf("%3s", n)
}

// Format Numbers as a dot-separated list
// Assuming a verion is [1,2]
//
// - %s     print version (1.2)
// - %4s    print version padded with zeros to 4 elements (1.2.0.0)
func (n Numbers) Format(s fmt.State, verb rune) {
	width := len(n)
	if w, ok := s.Width(); ok && w > width {
		width = w
	}
	for i := 0; i < width; i++ {
		if i > 0 {
			fmt.Fprint(s, ".")
		}
		if i < len(n) {
			fmt.Fprint(s, n[i].String())
		} else {
			fmt.Fprint(s, "0")
		}
	}
}

// Version encapsulates the 3 basic elements of a version
//
//              PreRelease
//                  |
//               \--+-/
//          1.0.0-rc.1+sha.5114f85
//         /-+-+-\    /---+-------\
//            |             |
//          Number     BuildMetadata
//
//
// To create this structure, it is recommended to use `ParseVersion` or `ParseSemanticVersion`
type Version struct {
	Number        Numbers
	PreRelease    Elements
	BuildMetadata Elements
}

// UnmarshalText implements encoding.TextUnmarshaler
func (v *Version) UnmarshalText(text []byte) error {
	ver, err := ParseVersion(string(text))
	if err != nil {
		return err
	}
	v.Number = ver.Number
	v.PreRelease = ver.PreRelease
	v.BuildMetadata = ver.BuildMetadata
	return nil
}

// MarshalText implements encoding.TextMarshaler
func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// Format formats the Version according to the fmt.Formatter interface.
//
//    %s    1.0.0-rc.1+sha.5114f85
//    %-s   1.0.0-rc.1 (hide build metadata)
//    %q    equivalent to %s enclosed in double-quotes
//    %-q   equivalent to %-s enclosed in double-quotes
func (v Version) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		if width, ok := s.Width(); ok {
			format := fmt.Sprintf("%%%ds", width)
			fmt.Fprintf(s, format, v.Number)
		} else {
			fmt.Fprintf(s, "%3s", v.Number)
		}
		if len(v.PreRelease) > 0 {
			fmt.Fprintf(s, "-%s", v.PreRelease)
		}
		if s.Flag('-') {
			return
		}
		if len(v.BuildMetadata) > 0 {
			fmt.Fprintf(s, "+%s", v.BuildMetadata)
		}
	case 'v':
		v.Format(s, 's')
	case 'q':
		fmt.Fprintf(s, `"`)
		v.Format(s, 's')
		fmt.Fprintf(s, `"`)
	default:
		v.Format(s, 'v')
	}
}

// Print Version assuming semver semantics
// Example: 1.0.0-rc.1+sha.5114f85
func (v Version) String() string {
	return fmt.Sprintf("%s", v)
}

// Compare two versions
// - negative if this version is < the given version
// - positive if this version is > the given version
// - zero if this version equals the given version
//
// When comparing the numeric portion of the version,
// the shorter one is padded with zeros so that both versions contain
// the same number of elements. For example: comparing 1.0 with 2.0.0
// will actually compare 1.0.0 with 2.0.0
//
// Otherwise, follows Semantic Versioning 2.0.0 precedence rules (https://semver.org/spec/v2.0.0.html)
//
// Build metadata is ignored in this comparison
func (v Version) Compare(o Version) int {
	if c := v.Number.Compare(o.Number); c != 0 {
		return c
	}
	if len(v.PreRelease) > 0 {
		if len(o.PreRelease) > 0 {
			return v.PreRelease.Compare(o.PreRelease)
		}
		// this version has a pre-release
		// but compared version does not
		// so this version has lower precedence
		return -1
	}
	if len(o.PreRelease) > 0 {
		// this version has no pre-release
		// but compared version does
		// so this version has higher precedence
		return 1
	}
	return 0
}

// Elements of a version
type Elements []Element

// String concatenates the elemnts separated by dots '.'
func (cs Elements) String() string {
	return fmt.Sprintf("%s", cs)
}

// Format Elements as a dot-separated list
func (cs Elements) Format(s fmt.State, verb rune) {
	for i, e := range cs {
		if i > 0 {
			fmt.Fprint(s, ".")
		}
		fmt.Fprint(s, e.String())
	}
}

// Compare two Element lists
func (cs Elements) Compare(os Elements) int {
	for i, x := range cs {
		if i < len(os) {
			if c := x.Compare(os[i]); c != 0 {
				return c
			}
		}
	}
	// If all preceding Elements are equal, then
	// higher count have higher precedence
	return compareUint64(uint64(len(cs)), uint64(len(os)))
}

// Element of a version component
type Element interface {
	// Compare two version Elements
	// If both Elements are numeric, they are compared numerically
	// If both Elements are non-numeric, then they are compared lexicographically
	// Otherwise, the numeric Element is always less than the non-numeric one
	Compare(e Element) int
	// String returns the element as a string
	String() string
}

// Number represents an Element of a numeric type for comparison purposes
type Number uint64

// Compare a number to another element
// Numbers are always lower precedence than TextElements
func (n Number) Compare(oe Element) int {
	if o, ok := oe.(Number); ok {
		return compareUint64(uint64(n), uint64(o))
	}
	// Numbers have lower precedence
	return -1
}

// String returns the number as a base-10 string
func (n Number) String() string {
	return strconv.FormatUint(uint64(n), 10)
}

// TextElement implements Element for lexicographical comparison
type TextElement string

// TextElement returns the element's string representation
func (e TextElement) String() string {
	return string(e)
}

// Compare a Text element with another text element
// A TextElement is always higher precedence than a non TextElement
func (e TextElement) Compare(oe Element) int {
	if o, ok := oe.(TextElement); ok {
		return strings.Compare(string(e), string(o))
	}
	// Text has a higher precedence
	return 1
}

// ParseElement a Element of a version string
func ParseElement(s string) (Element, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return TextElement(s), nil
	}
	if s == "0" {
		// There is only 1 valid way to represent zero
		return Number(0), nil
	}
	if strings.HasPrefix(s, "0") {
		// Zero-padded elements are not valid
		return TextElement(s), ErrInvalidVersion
	}
	return Number(i), nil
}

func compareUint64(a, b uint64) int {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}
