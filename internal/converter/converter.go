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
	"bufio"
	"errors"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

var (
	// ErrUnsupportedFormat is returned when a file format is not supported
	ErrUnsupportedFormat = errors.New("unsupported document format")
)

// DetectDocumentType detects the document type from file extension
func DetectDocumentType(filename string) types.DocumentType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html", ".htm":
		return types.TypeHTML
	case ".md":
		return types.TypeMarkdown
	case ".rst":
		return types.TypeReStructuredText
	case ".sgml", ".sgm", ".xml":
		return types.TypeSGML
	default:
		return types.TypeUnknown
	}
}

// Convert converts a document to markdown based on its type
func Convert(content []byte, docType types.DocumentType) (markdown string, title string, err error) {
	switch docType {
	case types.TypeHTML:
		return convertHTML(content)
	case types.TypeMarkdown:
		return processMarkdown(content)
	case types.TypeReStructuredText:
		return convertRST(content)
	case types.TypeSGML:
		return convertSGML(content)
	default:
		return "", "", ErrUnsupportedFormat
	}
}

// convertHTML converts HTML to Markdown and extracts the title
func convertHTML(content []byte) (string, string, error) {
	converter := md.NewConverter("", true, nil)

	// Add custom rule to shift heading levels down by one
	// (since we use the <title> as H1, all other headings should be shifted)
	converter.AddRules(md.Rule{
		Filter: []string{"h1", "h2", "h3", "h4", "h5", "h6"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			// Shift each heading level down by one
			level := 2 // h1 becomes h2 (##)
			switch selec.Nodes[0].Data {
			case "h1":
				level = 2
			case "h2":
				level = 3
			case "h3":
				level = 4
			case "h4":
				level = 5
			case "h5":
				level = 6
			case "h6":
				level = 6 // h6 stays at max level (######)
			}

			result := strings.Repeat("#", level) + " " + content
			return &result
		},
	})

	markdown, err := converter.ConvertBytes(content)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert HTML: %w", err)
	}

	// Extract title from HTML
	title := extractHTMLTitle(content)

	// Prepend title as H1 heading if we have one
	markdownStr := string(markdown)
	if title != "" {
		// The html-to-markdown library includes the title as plain text at the start
		// We need to replace it with a proper markdown heading
		markdownStr = strings.TrimSpace(markdownStr)

		// Check if the markdown starts with the title (without HTML entities decoded)
		// The library decodes entities in the output but we extract title from raw HTML
		if strings.HasPrefix(markdownStr, title) {
			// Remove the plain title and replace with heading
			markdownStr = strings.TrimPrefix(markdownStr, title)
			markdownStr = strings.TrimSpace(markdownStr)
		}

		// Add title as H1 heading
		markdownStr = "# " + title + "\n\n" + markdownStr
	}

	return markdownStr, title, nil
}

// extractHTMLTitle extracts the title from HTML <title> tag
func extractHTMLTitle(content []byte) string {
	titleRe := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := titleRe.FindSubmatch(content)
	if len(matches) > 1 {
		// Decode HTML entities (e.g., &#8212; -> —)
		title := html.UnescapeString(string(matches[1]))
		return strings.TrimSpace(title)
	}
	return ""
}

// processMarkdown processes Markdown and extracts the title
func processMarkdown(content []byte) (string, string, error) {
	// Markdown is already in the target format
	markdown := string(content)

	// Extract title from first # heading
	title := extractMarkdownTitle(markdown)

	return markdown, title, nil
}

// extractMarkdownTitle extracts the title from the first # heading
func extractMarkdownTitle(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inMetadata := false
	metadataDelimiterCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Skip YAML front matter
		if line == "---" {
			metadataDelimiterCount++
			if metadataDelimiterCount == 1 {
				inMetadata = true
				continue
			} else if metadataDelimiterCount == 2 {
				inMetadata = false
				continue
			}
		}

		if inMetadata {
			continue
		}

		// Look for first # heading
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}

	return ""
}

// convertRST converts reStructuredText to Markdown
func convertRST(content []byte) (string, string, error) {
	// Basic RST to Markdown conversion
	// This is a simplified implementation
	text := string(content)
	title := extractRSTTitle(text)

	// Convert RST headings to Markdown
	markdown := convertRSTHeadings(text)

	// Convert RST images to Markdown
	markdown = convertRSTImages(markdown)

	return markdown, title, nil
}

// extractRSTTitle extracts the title from reStructuredText
func extractRSTTitle(content string) string {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines)-1; i++ {
		current := strings.TrimSpace(lines[i])
		next := strings.TrimSpace(lines[i+1])

		// Skip RST directives, anchors, and labels (.. name: or .. _name:)
		if strings.HasPrefix(current, "..") && strings.HasSuffix(current, ":") {
			continue
		}

		// Check for overline+underline pattern (heading with line above and below)
		if i+2 < len(lines) && isUnderline(current) {
			text := strings.TrimSpace(lines[i+1])
			underline := strings.TrimSpace(lines[i+2])

			// Make sure the text line is not a directive either
			if text != "" && current == underline && isUnderline(underline) &&
				!(strings.HasPrefix(text, "..") && strings.HasSuffix(text, ":")) {
				// This is a heading with overline and underline - likely the title
				return cleanHeadingText(text)
			}
		}

		// Check for underline-only pattern (=, -, ~, etc.)
		if current != "" && next != "" {
			char := next[0]
			if (char == '=' || char == '-' || char == '~' || char == '#' || char == '*') &&
				strings.Count(next, string(char)) == len(next) &&
				len(next) >= len(current) {
				return cleanHeadingText(current)
			}
		}
	}

	return ""
}

// convertRSTHeadings converts RST-style headings to Markdown
func convertRSTHeadings(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	// Track heading patterns in order of appearance
	headingPatterns := make(map[string]int)
	nextLevel := 1

	i := 0
	for i < len(lines) {
		// Skip RST directives, anchors, and labels (.. name: or .. _name:)
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "..") && strings.HasSuffix(trimmed, ":") {
			i++
			continue
		}

		current := lines[i]
		currentTrim := strings.TrimSpace(current)

		// Check for heading with overline and underline
		if i+2 < len(lines) && isUnderline(currentTrim) {
			text := strings.TrimSpace(lines[i+1])
			underline := strings.TrimSpace(lines[i+2])

			if text != "" && currentTrim == underline && isUnderline(underline) {
				// This is a heading with overline and underline
				pattern := string(currentTrim[0]) + "o" // 'o' for overline
				level := getOrAssignLevel(pattern, headingPatterns, &nextLevel)
				cleanText := cleanHeadingText(text)
				result = append(result, strings.Repeat("#", level)+" "+cleanText)
				i += 3
				continue
			}
		}

		// Check for heading with just underline
		if i+1 < len(lines) && currentTrim != "" {
			next := strings.TrimSpace(lines[i+1])
			if isUnderline(next) && len(next) >= len(currentTrim) {
				// This is a heading with just underline
				pattern := string(next[0]) + "u" // 'u' for underline only
				level := getOrAssignLevel(pattern, headingPatterns, &nextLevel)
				cleanText := cleanHeadingText(currentTrim)
				result = append(result, strings.Repeat("#", level)+" "+cleanText)
				i += 2
				continue
			}
		}

		result = append(result, current)
		i++
	}

	return strings.Join(result, "\n")
}

// isUnderline checks if a line is a valid RST underline (all same punctuation)
func isUnderline(line string) bool {
	if line == "" {
		return false
	}

	// Check if all characters are the same punctuation
	char := line[0]
	if !isPunctuation(char) {
		return false
	}

	for _, c := range line {
		if byte(c) != char {
			return false
		}
	}

	return true
}

// isPunctuation checks if a character is a valid RST heading punctuation
func isPunctuation(c byte) bool {
	// Common RST heading characters
	punctuation := "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	return strings.ContainsRune(punctuation, rune(c))
}

// getOrAssignLevel gets or assigns a heading level for a pattern
func getOrAssignLevel(pattern string, patterns map[string]int, nextLevel *int) int {
	if level, exists := patterns[pattern]; exists {
		return level
	}

	level := *nextLevel
	patterns[pattern] = level
	*nextLevel++

	// Cap at level 6 (max Markdown heading level)
	if *nextLevel > 6 {
		*nextLevel = 6
	}

	return level
}

// cleanHeadingText removes RST directives and extra formatting from heading text
func cleanHeadingText(text string) string {
	// Remove inline directives like :index:, :ref:, etc.
	// Pattern: `text`:directive:
	re := regexp.MustCompile("`([^`]+)`:[a-zA-Z]+:")
	text = re.ReplaceAllString(text, "$1")

	// Remove just the directive part if no backticks
	// Pattern: :directive:
	re2 := regexp.MustCompile(":[a-zA-Z]+:")
	text = re2.ReplaceAllString(text, "")

	return strings.TrimSpace(text)
}

// convertRSTImages converts RST image directives to Markdown format
func convertRSTImages(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check for image or figure directive
		if strings.HasPrefix(trimmed, ".. image::") || strings.HasPrefix(trimmed, ".. figure::") {
			// Extract image path
			parts := strings.SplitN(trimmed, "::", 2)
			if len(parts) == 2 {
				imagePath := strings.TrimSpace(parts[1])
				altText := ""

				// Look ahead for :alt: option
				j := i + 1
				for j < len(lines) {
					nextLine := strings.TrimSpace(lines[j])

					// Stop if we hit a non-indented line or empty line after options
					if nextLine == "" {
						break
					}
					if !strings.HasPrefix(lines[j], "   ") && !strings.HasPrefix(lines[j], "\t") {
						break
					}

					// Extract alt text
					if strings.HasPrefix(nextLine, ":alt:") {
						altParts := strings.SplitN(nextLine, ":alt:", 2)
						if len(altParts) == 2 {
							altText = strings.TrimSpace(altParts[1])
						}
					}
					j++
				}

				// Convert to Markdown format
				markdownImage := fmt.Sprintf("![%s](%s)", altText, imagePath)
				result = append(result, markdownImage)
				result = append(result, "")

				// Skip the directive and its options
				i = j
				continue
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

// convertSGML converts SGML/DocBook to Markdown and extracts the title
func convertSGML(content []byte) (string, string, error) {
	text := string(content)

	// Extract title from SGML
	title := extractSGMLTitle(text)

	// Convert SGML tags to Markdown
	markdown := convertSGMLTags(text)

	// Prepend title as H1 heading if we have one and it's not already in the content
	if title != "" {
		markdown = strings.TrimSpace(markdown)
		// Check if markdown already starts with the title as a heading
		expectedStart := "# " + title
		if !strings.HasPrefix(markdown, expectedStart) {
			markdown = "# " + title + "\n\n" + markdown
		}
	}

	return markdown, title, nil
}

// extractSGMLTitle extracts the title from SGML/DocBook documents
func extractSGMLTitle(content string) string {
	// Try refentrytitle first (PostgreSQL-style reference pages)
	// This is more specific than generic <title> tags
	refTitleRe := regexp.MustCompile(`(?i)<refentrytitle[^>]*>([^<]+)</refentrytitle>`)
	matches := refTitleRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		return html.UnescapeString(strings.TrimSpace(matches[1]))
	}

	// Try to extract from <title> tags
	titleRe := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches = titleRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		return html.UnescapeString(strings.TrimSpace(matches[1]))
	}

	return ""
}

// convertSGMLTags converts SGML/DocBook tags to Markdown
func convertSGMLTags(content string) string {
	result := content

	// Remove SGML comments using simple string operations to avoid regex issues
	for {
		start := strings.Index(result, "<!--")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "-->")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+3:]
	}

	// Remove DOCTYPE declarations
	doctypeRe := regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`)
	result = doctypeRe.ReplaceAllString(result, "")

	// Remove XML declarations
	xmlDeclRe := regexp.MustCompile(`<\?xml[^?]*\?>`)
	result = xmlDeclRe.ReplaceAllString(result, "")

	// Convert headings
	result = convertSGMLHeadings(result)

	// Convert itemized lists BEFORE para conversion
	// Handle listitem with nested para specially - consume the opening para tag
	listItemParaRe := regexp.MustCompile(`(?i)<listitem[^>]*>\s*<para[^>]*>`)
	result = listItemParaRe.ReplaceAllString(result, "\n- ")
	// Handle remaining listitem tags without para
	itemRe := regexp.MustCompile(`(?i)<listitem[^>]*>`)
	result = itemRe.ReplaceAllString(result, "\n- ")
	// Handle closing para inside listitem - just remove it
	listItemEndParaRe := regexp.MustCompile(`(?i)</para>\s*</listitem>`)
	result = listItemEndParaRe.ReplaceAllString(result, "")
	itemEndRe := regexp.MustCompile(`(?i)</listitem>`)
	result = itemEndRe.ReplaceAllString(result, "")

	// Remove list container tags
	listRe := regexp.MustCompile(`(?i)</?(?:itemizedlist|orderedlist|variablelist|simplelist)[^>]*>`)
	result = listRe.ReplaceAllString(result, "\n")

	// Convert paragraph tags to proper spacing
	paraRe := regexp.MustCompile(`(?i)<para[^>]*>`)
	result = paraRe.ReplaceAllString(result, "\n\n")
	paraEndRe := regexp.MustCompile(`(?i)</para>`)
	result = paraEndRe.ReplaceAllString(result, "\n\n")

	// Convert emphasis to italic
	emphRe := regexp.MustCompile(`(?i)<emphasis[^>]*>([^<]*)</emphasis>`)
	result = emphRe.ReplaceAllString(result, "*$1*")

	// Convert code-like elements to inline code
	codeElements := []string{"literal", "command", "filename", "function", "type",
		"varname", "option", "parameter", "constant", "replaceable"}
	for _, elem := range codeElements {
		re := regexp.MustCompile(`(?i)<` + elem + `[^>]*>([^<]*)</` + elem + `>`)
		result = re.ReplaceAllString(result, "`$1`")
	}

	// Convert programlisting to code blocks
	progRe := regexp.MustCompile(`(?is)<programlisting[^>]*>(.*?)</programlisting>`)
	result = progRe.ReplaceAllStringFunc(result, func(match string) string {
		inner := progRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			code := strings.TrimSpace(inner[1])
			return "\n```\n" + code + "\n```\n"
		}
		return match
	})

	// Convert screen to code blocks (similar to programlisting)
	screenRe := regexp.MustCompile(`(?is)<screen[^>]*>(.*?)</screen>`)
	result = screenRe.ReplaceAllStringFunc(result, func(match string) string {
		inner := screenRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			code := strings.TrimSpace(inner[1])
			return "\n```\n" + code + "\n```\n"
		}
		return match
	})

	// Convert links
	linkRe := regexp.MustCompile(`(?i)<ulink[^>]*url="([^"]*)"[^>]*>([^<]*)</ulink>`)
	result = linkRe.ReplaceAllString(result, "[$2]($1)")

	// Convert xref links (just use the linkend as text)
	xrefRe := regexp.MustCompile(`(?i)<xref[^>]*linkend="([^"]*)"[^>]*/>`)
	result = xrefRe.ReplaceAllString(result, "`$1`")

	// Remove remaining tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	result = tagRe.ReplaceAllString(result, "")

	// Decode HTML entities
	result = html.UnescapeString(result)

	// Clean up excessive whitespace
	multiNewlineRe := regexp.MustCompile(`\n{3,}`)
	result = multiNewlineRe.ReplaceAllString(result, "\n\n")

	// Trim leading/trailing whitespace from each line
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	result = strings.Join(lines, "\n")

	return strings.TrimSpace(result)
}

// convertSGMLHeadings converts SGML/DocBook section tags to Markdown headings
func convertSGMLHeadings(content string) string {
	result := content

	// Map of SGML heading tags to Markdown levels
	headingMappings := []struct {
		tag   string
		level int
	}{
		{"chapter", 1},
		{"appendix", 1},
		{"article", 1},
		{"book", 1},
		{"sect1", 2},
		{"refsect1", 2},
		{"refsynopsisdiv", 2},
		{"sect2", 3},
		{"refsect2", 3},
		{"sect3", 4},
		{"refsect3", 4},
		{"sect4", 5},
		{"sect5", 6},
		{"section", 2}, // Generic section
	}

	for _, mapping := range headingMappings {
		// Match opening tag with nested title
		pattern := `(?is)<` + mapping.tag + `[^>]*>\s*<title[^>]*>([^<]*)</title>`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			inner := re.FindStringSubmatch(match)
			if len(inner) > 1 {
				title := html.UnescapeString(strings.TrimSpace(inner[1]))
				return "\n" + strings.Repeat("#", mapping.level) + " " + title + "\n"
			}
			return match
		})

		// Remove closing tags
		closeRe := regexp.MustCompile(`(?i)</` + mapping.tag + `>`)
		result = closeRe.ReplaceAllString(result, "\n")
	}

	// Handle refentry specially (PostgreSQL man pages)
	refentryRe := regexp.MustCompile(`(?is)<refentry[^>]*>`)
	result = refentryRe.ReplaceAllString(result, "")
	refentryEndRe := regexp.MustCompile(`(?i)</refentry>`)
	result = refentryEndRe.ReplaceAllString(result, "")

	// Handle refnamediv (name and purpose)
	refnamedivRe := regexp.MustCompile(`(?is)<refnamediv[^>]*>.*?<refname[^>]*>([^<]*)</refname>.*?<refpurpose[^>]*>([^<]*)</refpurpose>.*?</refnamediv>`)
	result = refnamedivRe.ReplaceAllStringFunc(result, func(match string) string {
		inner := refnamedivRe.FindStringSubmatch(match)
		if len(inner) > 2 {
			name := html.UnescapeString(strings.TrimSpace(inner[1]))
			purpose := html.UnescapeString(strings.TrimSpace(inner[2]))
			return "\n## " + name + "\n\n" + purpose + "\n"
		}
		return match
	})

	return result
}

// IsSupported returns true if the file type is supported
func IsSupported(filename string) bool {
	docType := DetectDocumentType(filename)
	return docType != types.TypeUnknown
}

// GetSupportedExtensions returns a list of supported file extensions
func GetSupportedExtensions() []string {
	return []string{".html", ".htm", ".md", ".rst", ".sgml", ".sgm", ".xml"}
}

// ReadAll reads all content from a reader
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
