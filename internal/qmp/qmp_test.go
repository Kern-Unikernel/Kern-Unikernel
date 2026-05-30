package qmp

import (
	"testing"
)

func TestConnectInvalidSocket(t *testing.T) {
	_, err := Connect("/tmp/nonexistent.sock")
	if err == nil {
		t.Error("should return error for nonexistent socket")
	}
}

func TestFindSocketNoPID(t *testing.T) {
	_, err := FindSocket(-1)
	if err == nil {
		t.Error("should return error for invalid PID")
	}
}
