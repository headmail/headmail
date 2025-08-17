package service

import (
	"log"
	"strings"
	"testing"
)

func TestInjectTracking_TableDriven(t *testing.T) {
	s := &DeliveryService{trackingHost: "https://tracking.example.com"}
	deliveryID := "del-123"

	cases := []struct {
		name     string
		in       string
		expected []string // substrings expected to appear
		not      []string // substrings that must NOT appear
	}{
		{
			name: "rewrite anchor only",
			in: `<html><body>
				<a href="https://example.com/page">link</a>
				<img src="https://example.com/img" />
				<link href="https://example.com/link" />
				<a href="mailto:foo@example.com">mail</a>
				</body></html>`,
			expected: []string{
				`<a href="https://tracking.example.com/r/del-123/c?u=https%3A%2F%2Fexample.com%2Fpage">`,
				`<img src="https://example.com/img"`,
				`<link href="https://example.com/link"`,
				`<a href="mailto:foo@example.com">`,
			},
		},
		{
			name: "skip anchors and javascript",
			in: `<html><body>
				<a href="#top">top</a>
				<a href="javascript:void(0)">js</a>
				</body></html>`,
			expected: []string{},
			not:      []string{"/r/" + deliveryID + "/c?u="},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := s.injectTracking(deliveryID, tc.in)
			log.Printf("%s", out)
			for _, want := range tc.expected {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q; output=%s", want, out)
				}
			}
			for _, nw := range tc.not {
				if strings.Contains(out, nw) && len(nw) > 0 {
					t.Fatalf("did not expect output to contain %q; output=%s", nw, out)
				}
			}
		})
	}
}

func TestAppendPixel_TableDriven(t *testing.T) {
	s := &DeliveryService{trackingHost: "https://tracking.example.com"}
	cases := []struct {
		name     string
		in       string
		expected string // substring that must exist after processing
	}{
		{
			name:     "with body",
			in:       `<html><body><p>Hello</p></body></html>`,
			expected: "https://tracking.example.com/r/del-xyz/o",
		},
		{
			name:     "no body",
			in:       `<div>Hello</div>`,
			expected: "https://tracking.example.com/r/del-xyz/o",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := s.appendPixel(tc.in, "del-xyz")
			if !strings.Contains(out, tc.expected) {
				t.Fatalf("expected output to contain %q; got: %s", tc.expected, out)
			}
		})
	}
}
