package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSEOAuditScraping(t *testing.T) {
	mockHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Best Espresso Machines 2026 | Comprehensive Guide</title>
	<meta name="description" content="Discover top-rated espresso machines, user reviews, pricing breakdown, and expert brewing advice to help you select the ideal home coffee setup.">
	<link rel="canonical" href="https://example.com/espresso-guide">
</head>
<body>
	<h1>Ultimate Espresso Machine Guide</h1>
	<h2>Top 5 Picks for Beginners</h2>
	<p>Here is our list of machines.</p>
	<img src="machine.jpg" alt="Stainless steel espresso machine">
	<img src="shot.jpg" alt="Rich crema espresso shot">
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	svc := NewSEOAuditService()
	result, err := svc.AuditURL(server.URL)
	if err != nil {
		t.Fatalf("AuditURL failed: %v", err)
	}

	expectedTitle := "Best Espresso Machines 2026 | Comprehensive Guide"
	if result.Title != expectedTitle {
		t.Errorf("expected Title %q, got %q", expectedTitle, result.Title)
	}

	expectedH1 := "Ultimate Espresso Machine Guide"
	if result.H1Count != 1 {
		t.Errorf("expected 1 H1 tag, got %d", result.H1Count)
	}
	if result.H1Primary != expectedH1 {
		t.Errorf("expected H1Primary %q, got %q", expectedH1, result.H1Primary)
	}

	if result.Description == "" {
		t.Errorf("expected non-empty Description")
	}

	if result.ImagesTotal != 2 || result.ImagesWithoutAlt != 0 {
		t.Errorf("expected 2 images with 0 missing alt, got total %d, missing %d", result.ImagesTotal, result.ImagesWithoutAlt)
	}

	if result.Score < 80 {
		t.Errorf("expected high score >= 80 for optimal page, got %d", result.Score)
	}
}

func TestSEOAuditMultipleH1(t *testing.T) {
	mockHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Short Title</title>
</head>
<body>
	<h1>First Main Heading</h1>
	<h1>Second Main Heading</h1>
	<h1>Third Heading</h1>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	svc := NewSEOAuditService()
	result, err := svc.AuditURL(server.URL)
	if err != nil {
		t.Fatalf("AuditURL failed: %v", err)
	}

	if result.H1Count != 3 {
		t.Errorf("expected 3 H1 tags, got %d", result.H1Count)
	}

	hasMultipleH1Issue := false
	for _, issue := range result.Issues {
		if len(issue) > 0 && (issue[:8] == "Multiple" || issue[:7] == "Missing") {
			hasMultipleH1Issue = true
			break
		}
	}

	if !hasMultipleH1Issue {
		t.Errorf("expected issue regarding multiple H1 tags, got issues: %v", result.Issues)
	}
}

func TestSEOAuditMissingTags(t *testing.T) {
	mockHTML := `<!DOCTYPE html>
<html>
<head>
</head>
<body>
	<p>Empty page with no headers or meta</p>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	svc := NewSEOAuditService()
	result, err := svc.AuditURL(server.URL)
	if err != nil {
		t.Fatalf("AuditURL failed: %v", err)
	}

	if result.Title != "" {
		t.Errorf("expected empty Title, got %q", result.Title)
	}
	if result.Description != "" {
		t.Errorf("expected empty Description, got %q", result.Description)
	}
	if result.H1Count != 0 {
		t.Errorf("expected 0 H1 tags, got %d", result.H1Count)
	}
	if result.Score >= 50 {
		t.Errorf("expected low score < 50 for page missing all meta, got %d", result.Score)
	}
}
