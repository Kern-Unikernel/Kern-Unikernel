package registry

import (
	"testing"
)

func TestPushImageInvalidReference(t *testing.T) {
	err := PushImage("/tmp/test.uki", "invalid!!reference", Credentials{})
	if err == nil {
		t.Error("should return error for invalid reference")
	}
}

func TestPullImageInvalidReference(t *testing.T) {
	err := PullImage("invalid!!reference", "/tmp/test", Credentials{})
	if err == nil {
		t.Error("should return error for invalid reference")
	}
}

func TestPullImageNonExistent(t *testing.T) {
	err := PullImage("ghcr.io/nonexistent/nonexistent:latest", "/tmp/test", Credentials{})
	if err == nil {
		t.Error("should return error for nonexistent image")
	}
}
