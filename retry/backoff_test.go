package retry

import (
	"fmt"
	"testing"
	"time"
)

func TestConstBackoff(t *testing.T) {
	z := time.Duration(5 * time.Second)
	boff := ConstantBackoff(z)
	for i := uint(0); i < 1000; i++ {
		if d := boff(i); d != z {
			t.Fatalf("violated: actual %v != expected %v", d, z)
		}
	}
}

func TestExponentialRandomBackoff(t *testing.T) {
	w := 100 * time.Millisecond
	eboff := ExpontentialRandomBackoff(w, 5)
	examples := []struct {
		n     uint
		lower time.Duration
		upper time.Duration
	}{
		{0, time.Duration(0), time.Duration(0)},
		{1, time.Duration(0), w * 2},
		{2, time.Duration(0), w * 4},
		{3, time.Duration(0), w * 8},
		{4, time.Duration(0), w * 16},
		{5, time.Duration(0), w * 32},
		{6, time.Duration(0), w * 32},
		{7, time.Duration(0), w * 32},
		{8, time.Duration(0), w * 32},
	}
	for _, ex := range examples {
		t.Run(fmt.Sprintf("eboff-%d", ex.n), func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				d := eboff(ex.n)
				if d > ex.upper || d < ex.lower {
					t.Fatalf("violated: %v < EBOff(%d)=%v < %v", ex.lower, ex.n, d, ex.upper)
				}
			}
		})
	}
}
