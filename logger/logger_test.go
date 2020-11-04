package logger

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer

	logger := New(&buf)
	if logger == nil {
		t.Error("shouldnt be nil")
	} else {
		logger.Log("logging")
		if buf.String() != "logging\n" {
			t.Errorf("not correct '%s'", buf.String())
		}
	}
}

func TestSilent(t *testing.T) {
	var silentLogger Logger = Silent()
	silentLogger.Log("silence")
}
