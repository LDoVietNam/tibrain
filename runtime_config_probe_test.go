package main

import "testing"

func TestRuntimeConfigProbe(t *testing.T) {
	value := resolveMCPHubKeyFromFiles()
	t.Logf("configured=%t length=%d", value != "", len(value))
}