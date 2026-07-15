package config

import (
	"os"
	"sync"

	"github.com/metacubex/mihomo/constant"
)

type OverrideSlot int

const (
	OverrideSlotPersist OverrideSlot = iota
	OverrideSlotSession
)

const defaultPersistOverride = `{}`
const defaultSessionOverride = `{}`

var sessionOverride = defaultSessionOverride
var overrideMu sync.RWMutex

func overridePersistPath() string {
	return constant.Path.Resolve("override.json")
}

func ReadOverride(slot OverrideSlot) string {
	overrideMu.RLock()
	defer overrideMu.RUnlock()

	switch slot {
	case OverrideSlotPersist:
		buf, err := os.ReadFile(overridePersistPath())
		if err != nil {
			return defaultPersistOverride
		}

		return string(buf)
	case OverrideSlotSession:
		return sessionOverride
	}

	return ""
}

func WriteOverride(slot OverrideSlot, content string) {
	overrideMu.Lock()
	defer overrideMu.Unlock()

	switch slot {
	case OverrideSlotPersist:
		_ = os.WriteFile(overridePersistPath(), []byte(content), 0600)
	case OverrideSlotSession:
		sessionOverride = content
	}
}

func ClearOverride(slot OverrideSlot) {
	overrideMu.Lock()
	defer overrideMu.Unlock()

	switch slot {
	case OverrideSlotPersist:
		_ = os.Remove(overridePersistPath())
	case OverrideSlotSession:
		sessionOverride = defaultSessionOverride
	}
}
