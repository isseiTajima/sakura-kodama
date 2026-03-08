package ws

import (
	"reflect"
	"testing"
)

func TestEventIncludesTimestampAndProfileFields(t *testing.T) {
	typ := reflect.TypeOf(Event{})

	timestampField, ok := typ.FieldByName("Timestamp")
	if !ok {
		t.Fatal("Event struct must include Timestamp field")
	}
	if tag := timestampField.Tag.Get("json"); tag != "timestamp" {
		t.Fatalf("Timestamp field must have json tag \"timestamp\", got %q", tag)
	}

	profileField, ok := typ.FieldByName("Profile")
	if !ok {
		t.Fatal("Event struct must include Profile field")
	}
	if tag := profileField.Tag.Get("json"); tag != "profile" {
		t.Fatalf("Profile field must have json tag \"profile\", got %q", tag)
	}
}
