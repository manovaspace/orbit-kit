package messagebus_test

import (
	"path/filepath"
	"testing"

	messagebus "github.com/manovaspace/orbit-kit/messagebus"
	"github.com/manovaspace/orbit-kit/messagebus/bustest"
)

func TestEnvelopeSchema(t *testing.T) {
	bustest.CheckEventSchema(t, messagebus.Envelope{}, filepath.Join("testdata", "envelope.schema.json"))
}
