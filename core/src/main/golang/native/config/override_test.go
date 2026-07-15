package config

import (
	"strings"
	"sync"
	"testing"

	"github.com/metacubex/mihomo/constant"
)

func useTemporaryOverrideHome(t *testing.T) {
	t.Helper()

	original := constant.Path.HomeDir()
	constant.SetHomeDir(t.TempDir())
	t.Cleanup(func() {
		ClearOverride(OverrideSlotPersist)
		ClearOverride(OverrideSlotSession)
		constant.SetHomeDir(original)
	})
}

func TestPersistOverrideRoundTrip(t *testing.T) {
	useTemporaryOverrideHome(t)

	if got := ReadOverride(OverrideSlotPersist); got != defaultPersistOverride {
		t.Fatalf("default persist override = %q, want %q", got, defaultPersistOverride)
	}

	want := `{"prepend-rules":["DOMAIN,example.com,DIRECT"]}`
	WriteOverride(OverrideSlotPersist, want)
	if got := ReadOverride(OverrideSlotPersist); got != want {
		t.Fatalf("persist override = %q, want %q", got, want)
	}

	ClearOverride(OverrideSlotPersist)
	if got := ReadOverride(OverrideSlotPersist); got != defaultPersistOverride {
		t.Fatalf("cleared persist override = %q, want %q", got, defaultPersistOverride)
	}
}

func TestSessionOverrideRoundTrip(t *testing.T) {
	useTemporaryOverrideHome(t)

	want := `{"mode":"direct"}`
	WriteOverride(OverrideSlotSession, want)
	if got := ReadOverride(OverrideSlotSession); got != want {
		t.Fatalf("session override = %q, want %q", got, want)
	}

	ClearOverride(OverrideSlotSession)
	if got := ReadOverride(OverrideSlotSession); got != defaultSessionOverride {
		t.Fatalf("cleared session override = %q, want %q", got, defaultSessionOverride)
	}
}

func TestPersistOverrideConcurrentReadWrite(t *testing.T) {
	useTemporaryOverrideHome(t)

	values := []string{
		`{"prepend-rules":["DOMAIN,one.example,DIRECT"]}` + strings.Repeat(" ", 4096),
		`{"prepend-rules":["DOMAIN,two.example,REJECT"]}` + strings.Repeat(" ", 4096),
	}
	valid := map[string]bool{
		defaultPersistOverride: true,
		values[0]:              true,
		values[1]:              true,
	}

	const iterations = 100
	errors := make(chan string, iterations*2)
	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		defer wait.Done()
		for i := 0; i < iterations; i++ {
			WriteOverride(OverrideSlotPersist, values[i%len(values)])
		}
	}()

	go func() {
		defer wait.Done()
		for i := 0; i < iterations; i++ {
			if got := ReadOverride(OverrideSlotPersist); !valid[got] {
				errors <- got
			}
		}
	}()

	wait.Wait()
	close(errors)
	for value := range errors {
		t.Fatalf("read partial override with length %d", len(value))
	}
}
