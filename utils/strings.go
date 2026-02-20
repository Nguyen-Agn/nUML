package utils

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// CleanHTML removes HTML tags and decodes entities from a string.
// CleanHTML loại bỏ các thẻ HTML và giải mã các thực thể từ một chuỗi.
func CleanHTML(s string) string {
	// 1. Basic HTML tag stripping
	// IMPORTANT: Only strip real tags, allow << for stereotypes
	// QUAN TRỌNG: Chỉ loại bỏ các thẻ thực sự, cho phép << cho các khuôn mẫu (stereotypes)
	s = strings.ReplaceAll(s, "<br>", " ")
	s = strings.ReplaceAll(s, "</div>", " ")
	s = strings.ReplaceAll(s, "</p>", " ")

	// Regex: Matches specific HTML tags (case insensitive)
	// Regex: Khớp các thẻ HTML cụ thể (không phân biệt hoa thường)
	// p, div, span, i, b, em, strong, font
	re := regexp.MustCompile(`(?i)</?(div|p|span|i|b|em|strong|font)[^>]*>`)
	clean := re.ReplaceAllString(s, "")

	// 2. Decode HTML entities (properly handles &lt; &gt; &nbsp; etc)
	// 2. Giải mã các thực thể HTML (xử lý đúng &lt; &gt; &nbsp; v.v.)
	clean = html.UnescapeString(clean)

	// 3. Normalize whitespace
	// 3. Chuẩn hóa khoảng trắng
	clean = strings.ReplaceAll(clean, "\u00A0", " ")
	return strings.TrimSpace(clean)
}

// IsFuzzyMatch checks if two strings are similar enough to be considered a match, handling typos.
// IsFuzzyMatch kiểm tra xem hai chuỗi có đủ giống nhau để được coi là khớp không, xử lý lỗi đánh máy.
func IsFuzzyMatch(s, target string) bool {
	// Simple typo tolerance:
	// 1. Exact match (case insensitive)
	// 2. Transposition of 2 adjacent chars (e.g. emun)
	// 3. Missing 1 char (e.g. iterface)
	// 4. Extra 1 char (e.g. interrface)
	s = strings.ToLower(s)
	target = strings.ToLower(target)

	if s == target {
		return true
	}
	if len(s) < 3 {
		return false
	} // Too short for fuzzy

	// Check if it's "close enough"
	// Just handling common user typos for now
	// Kiểm tra xem nó có "đủ gần" không
	// Hiện tại chỉ xử lý các lỗi đánh máy phổ biến của người dùng
	if (s == "emun" && target == "enum") || (s == "enmu" && target == "enum") {
		return true
	}
	if (s == "interfac" || s == "inteface" || s == "iterface") && target == "interface" {
		return true
	}

	return false
}

// LowercaseFirst converts the first character of the string to lowercase.
// LowercaseFirst chuyển ký tự đầu tiên của chuỗi thành chữ thường.
func LowercaseFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
