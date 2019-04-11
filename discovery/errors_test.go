package discovery

import (
	"testing"
)

func TestErrUndiscoverableBypassesDiscovery(t *testing.T) {
	if !ErrUndiscoverable.BypassDiscovery() {
		t.Fatal("ErrUndiscoverable.BypassDiscovery() must be true")
	}
}
