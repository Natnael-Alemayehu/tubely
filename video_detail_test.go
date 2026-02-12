package main

import (
	"testing"
)

func TestVideoDetail(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "horizontal",
			filename: "./samples/boots-video-horizontal.mp4",
			expected: "16:9",
		},
		{
			name:     "vertival",
			filename: "./samples/boots-video-vertical.mp4",
			expected: "9:16",
		},
	}

	for _, tt := range cases {
		got, err := getVideoAspectRatio(tt.filename)
		if err != nil {
			t.Fatalf("ERROR OCCURED: %v", err)
			return
		}

		if got != tt.expected {
			t.Fatalf("GOT: %v, EXPECTED: %v", got, tt.expected)
		}
	}

}
