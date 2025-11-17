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

    "github.com/PuerkitoBio/goquery"
    md "github.com/JohannesKaufmann/html-to-markdown"

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

// IsSupported returns true if the file type is supported
func IsSupported(filename string) bool {
    docType := DetectDocumentType(filename)
    return docType != types.TypeUnknown
}

// GetSupportedExtensions returns a list of supported file extensions
func GetSupportedExtensions() []string {
    return []string{".html", ".htm", ".md", ".rst"}
}

// ReadAll reads all content from a reader
func ReadAll(r io.Reader) ([]byte, error) {
    return io.ReadAll(r)
}
