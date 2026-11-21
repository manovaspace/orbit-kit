package inproc_test

import (
	"testing"

	messagebus "github.com/manovaspace/orbit-kit/messagebus"
	"github.com/manovaspace/orbit-kit/messagebus/bustest"
	"github.com/manovaspace/orbit-kit/messagebus/inproc"
)

func TestConformance(t *testing.T) {
	bustest.Conformance(t, func(_ *testing.T) messagebus.Bus {
		return inproc.New(messagebus.Config{Source: "conformance", Version: "1"})
	})
}
