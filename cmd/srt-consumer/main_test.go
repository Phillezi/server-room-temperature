package main

import (
	"bytes"
	"io"
	"testing"
)

func TestFrameReader_MultipleFrames(t *testing.T) {
	input := "@1234\n@5678\n@-999\n"
	fr := NewFrameReader(bytes.NewBufferString(input))

	expected := []string{"@1234\n", "@5678\n", "@-999\n"}
	var frame [32]byte

	for i, want := range expected {
		n, err := fr.ReadFrame(frame[:])
		if err != nil {
			t.Fatalf("frame %d: unexpected error: %v", i, err)
		}
		got := string(frame[:n])
		if got != want {
			t.Fatalf("frame %d: got %q, want %q", i, got, want)
		}
	}

	_, err := fr.ReadFrame(frame[:])
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestFrameReader_LeadingGarbageAndCorruptedFrames(t *testing.T) {
	// Leading noise, then a valid frame, then an oversized/garbage frame, then a valid frame
	input := "garbagedata@1234\n@verylongcorruptedandunparseableframeoverflowingbuffer12345678901234567890\n@9999\n"
	fr := NewFrameReader(bytes.NewBufferString(input))

	var frame [32]byte

	n, err := fr.ReadFrame(frame[:])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(frame[:n]); got != "@1234\n" {
		t.Fatalf("got %q, want %q", got, "@1234\n")
	}

	n, err = fr.ReadFrame(frame[:])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(frame[:n]); got != "@9999\n" {
		t.Fatalf("got %q, want %q", got, "@9999\n")
	}
}
