package main

import (
	"strings"
	"testing"
)

func TestDemoBathhouseImages_ReturnRelativePaths(t *testing.T) {
	images := demoBathhouseImages("http://localhost:3000", "banya-couple")
	if len(images) == 0 {
		t.Fatal("expected non-empty images")
	}

	for _, image := range images {
		if !strings.HasPrefix(image, "/demo/bathhouses/") {
			t.Fatalf("expected relative demo asset path, got %q", image)
		}
		if strings.Contains(image, "localhost:3000") {
			t.Fatalf("expected helper to avoid baked localhost frontend URL, got %q", image)
		}
	}
}

func TestDemoReviewImages_ReturnRelativePaths(t *testing.T) {
	images := demoReviewImages("http://localhost:3000", "review-premium")
	if len(images) != 2 {
		t.Fatalf("expected 2 demo review images, got %d", len(images))
	}

	for _, image := range images {
		if !strings.HasPrefix(image, "/demo/bathhouses/") {
			t.Fatalf("expected relative demo asset path, got %q", image)
		}
		if strings.Contains(image, "localhost:3000") {
			t.Fatalf("expected helper to avoid baked localhost frontend URL, got %q", image)
		}
	}
}
