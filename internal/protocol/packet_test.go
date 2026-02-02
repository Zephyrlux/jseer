package protocol

import "testing"

func TestBuildAndParse(t *testing.T) {
	body := []byte{1, 2, 3}
	pkt := BuildResponse(1001, 42, 0, body)
	length, _, cmdID, userID, _, parsedBody, err := ParsePacket(pkt)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if length != len(pkt) {
		t.Fatalf("length mismatch: %d vs %d", length, len(pkt))
	}
	if cmdID != 1001 || userID != 42 {
		t.Fatalf("header mismatch cmd=%d user=%d", cmdID, userID)
	}
	if len(parsedBody) != len(body) {
		t.Fatalf("body len mismatch")
	}
}
