package dispatcher

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestPayloadUsesSnakeCaseFieldNames(t *testing.T) {
	payload, err := json.Marshal(Payload{
		EventID:          uuid.New(),
		EventType:        "SessionRevoked",
		UserID:           uuid.New(),
		CentralSessionID: ptrUUID(uuid.New()),
		Reason:           "USER_LOGOUT",
	})
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"event_id", "event_type", "user_id", "central_session_id", "reason"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("expected JSON field %q in %s", field, payload)
		}
	}
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
