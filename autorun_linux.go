//go:build linux

package main

func self() string {
	// Linux stub: return empty path
	return ""
}

func AutoRunEnabled() (bool, error) {
	// Linux stub: always return false
	return false, nil
}

func AutoRunDisable() error {
	// Linux stub: do nothing
	return nil
}

func AutoRunEnable() error {
	// Linux stub: do nothing
	return nil
}
