//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package converter

import (
	"strings"
	"testing"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

func TestDetectDocumentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected types.DocumentType
	}{
		{"HTML file", "test.html", types.TypeHTML},
		{"HTM file", "test.htm", types.TypeHTML},
		{"Markdown file", "test.md", types.TypeMarkdown},
		{"RST file", "test.rst", types.TypeReStructuredText},
		{"Unknown file", "test.txt", types.TypeUnknown},
		{"No extension", "test", types.TypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDocumentType(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvertHTML(t *testing.T) {
	tests := []struct {
		name             string
		html             []byte
		expectedTitle    string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			"Basic HTML with title",
			[]byte(`
                <!DOCTYPE html>
                <html>
                <head><title>Test Title</title></head>
                <body>
                    <h1>Heading</h1>
                    <p>This is a paragraph.</p>
                </body>
                </html>
            `),
			"Test Title",
			[]string{"# Test Title", "## Heading", "This is a paragraph"},
			[]string{},
		},
		{
			"HTML with entity in title",
			[]byte(`
                <!DOCTYPE html>
                <html>
                <head><title>Test &#8212; Title</title></head>
                <body>
                    <p>Content with &#8212; dash</p>
                </body>
                </html>
            `),
			"Test — Title",
			[]string{"# Test — Title", "Content with — dash"},
			[]string{"&#8212;"},
		},
		{
			"HTML with multiple heading levels",
			[]byte(`
                <!DOCTYPE html>
                <html>
                <head><title>Page Title</title></head>
                <body>
                    <h1>Section 1</h1>
                    <p>Content 1</p>
                    <h2>Section 1.1</h2>
                    <p>Content 1.1</p>
                    <h3>Section 1.1.1</h3>
                    <p>Content 1.1.1</p>
                </body>
                </html>
            `),
			"Page Title",
			[]string{
				"# Page Title",
				"## Section 1",
				"### Section 1.1",
				"#### Section 1.1.1",
			},
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, title, err := convertHTML(tt.html)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if title != tt.expectedTitle {
				t.Errorf("expected title '%s', got '%s'", tt.expectedTitle, title)
			}

			for _, expected := range tt.shouldContain {
				if !strings.Contains(markdown, expected) {
					t.Errorf("markdown should contain '%s', got:\n%s", expected, markdown)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(markdown, notExpected) {
					t.Errorf("markdown should not contain '%s', got:\n%s", notExpected, markdown)
				}
			}
		})
	}
}

func TestProcessMarkdown(t *testing.T) {
	markdown := []byte(`---
title: Frontmatter Title
---

# Main Title

This is content.
`)

	result, title, err := processMarkdown(markdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if title != "Main Title" {
		t.Errorf("expected title 'Main Title', got '%s'", title)
	}

	if result != string(markdown) {
		t.Error("markdown should be unchanged")
	}
}

func TestExtractMarkdownTitle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			"Simple title",
			"# Test Title\n\nContent",
			"Test Title",
		},
		{
			"Title with frontmatter",
			"---\ntitle: FM Title\n---\n\n# Main Title\n\nContent",
			"Main Title",
		},
		{
			"No title",
			"Content without title",
			"",
		},
		{
			"H2 not extracted",
			"## Second Level\n\nContent",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMarkdownTitle(tt.content)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExtractRSTTitle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			"Title with equals",
			"Main Title\n==========\n\nContent",
			"Main Title",
		},
		{
			"Title with dashes",
			"Subtitle\n--------\n\nContent",
			"Subtitle",
		},
		{
			"No title",
			"Just content without underline",
			"",
		},
		{
			"Title with directive in backticks",
			"`Add Named Restore Point Dialog`:index:\n==========================================\n\nContent",
			"Add Named Restore Point Dialog",
		},
		{
			"Title with overline and underline with directive",
			"*************************\n`Coding Standards`:index:\n*************************\n\nContent",
			"Coding Standards",
		},
		{
			"Title with directive without backticks",
			"Configuration :ref:\n===================\n\nContent",
			"Configuration",
		},
		{
			"Title after RST anchor without underscore",
			".. cloud_azure_database:\n\nAzure Database Cloud Deployment\n================================\n\nContent",
			"Azure Database Cloud Deployment",
		},
		{
			"Title after RST anchor with underscore",
			".. _cloud_azure_database:\n\nAzure Database Cloud Deployment\n================================\n\nContent",
			"Azure Database Cloud Deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRSTTitle(tt.content)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"HTML supported", "test.html", true},
		{"Markdown supported", "test.md", true},
		{"RST supported", "test.rst", true},
		{"TXT not supported", "test.txt", false},
		{"Unknown not supported", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupported(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetSupportedExtensions(t *testing.T) {
	exts := GetSupportedExtensions()

	expected := []string{".html", ".htm", ".md", ".rst"}

	if len(exts) != len(expected) {
		t.Errorf("expected %d extensions, got %d", len(expected), len(exts))
	}

	for _, exp := range expected {
		found := false
		for _, ext := range exts {
			if ext == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected extension %s not found", exp)
		}
	}
}

func TestConvertRSTHeadings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			"Basic headings",
			`Main Title
==========

Subtitle
--------

Content here.`,
			[]string{"# Main Title", "## Subtitle"},
		},
		{
			"Heading with overline and underline",
			`.. _coding_standards:

*************************
` + "`Coding Standards`:index:" + `
*************************

Sub heading 1
*************`,
			[]string{"# Coding Standards", "## Sub heading 1"},
		},
		{
			"Arbitrary punctuation order",
			`First Heading
~~~~~~~~~~~~~

Second Heading
**************

Third Heading
~~~~~~~~~~~~~`,
			[]string{"# First Heading", "## Second Heading", "# Third Heading"},
		},
		{
			"Mixed overline and underline",
			`=====
Title
=====

Subtitle
--------`,
			[]string{"# Title", "## Subtitle"},
		},
		{
			"Anchor without underscore before heading",
			`.. cloud_azure_database:

Azure Database Cloud Deployment
================================

Some content here.`,
			[]string{"# Azure Database Cloud Deployment"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRSTHeadings(tt.input)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected '%s' in output, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestCleanHeadingText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Directive with backticks",
			"`Coding Standards`:index:",
			"Coding Standards",
		},
		{
			"Directive without backticks",
			"Introduction :ref:",
			"Introduction",
		},
		{
			"Multiple directives",
			"`Installation`:index: :ref:",
			"Installation",
		},
		{
			"No directive",
			"Plain Text",
			"Plain Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanHeadingText(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConvertRSTImages(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Image with alt text",
			`.. image:: images/screenshot.png
   :alt: Screenshot of the application`,
			"![Screenshot of the application](images/screenshot.png)",
		},
		{
			"Image without alt text",
			`.. image:: path/to/image.jpg`,
			"![](path/to/image.jpg)",
		},
		{
			"Figure directive",
			`.. figure:: diagrams/architecture.svg
   :alt: System architecture diagram
   :width: 500px`,
			"![System architecture diagram](diagrams/architecture.svg)",
		},
		{
			"Multiple images",
			`Some text

.. image:: image1.png
   :alt: First image

More text

.. image:: image2.png
   :alt: Second image`,
			"![First image](image1.png)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRSTImages(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected output to contain '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConvertRSTHeadingsRemovesAnchors(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldNotContain string
	}{
		{
			"Anchor without underscore removed",
			`.. cloud_azure_database:

Azure Database Cloud Deployment
================================`,
			".. cloud_azure_database:",
		},
		{
			"Anchor with underscore removed",
			`.. _cloud_azure_database:

Azure Database Cloud Deployment
================================`,
			".. _cloud_azure_database:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRSTHeadings(tt.input)
			if strings.Contains(result, tt.shouldNotContain) {
				t.Errorf("output should not contain '%s', but got:\n%s", tt.shouldNotContain, result)
			}
		})
	}
}
