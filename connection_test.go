package siridb

import (
	"testing"
)

func TestConnection(t *testing.T) {
	conn := NewConnection("::1", 9000)
	want := "[::1]:9000"

	if conn.ToString() != want {
		t.Errorf("conn.ToString() == %v, want %v", conn.ToString(), want)
	}

	conn = NewConnection("localhost", 5050)
	want = "localhost:5050"

	if conn.ToString() != want {
		t.Errorf("conn.ToString() == %v, want %v", conn.ToString(), want)
	}
}
