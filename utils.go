package main

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"
	"time"
)

// Global Log
var logBuffer strings.Builder
var targetPackage string
var VerboseMode bool
var OverwriteMode bool
var NoReportMode bool

func setupLogging() {
	logBuffer.Reset()
}

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

func logInfo(msg string) {
	fmt.Println(msg)
	logBuffer.WriteString("- " + msg + "\n")
}

func logVerbose(msg string) {
	if VerboseMode {
		parts := strings.SplitN(msg, ":", 2)
		if len(parts) == 2 {
			fmt.Println(Purple + "[VERBOSE] " + Cyan + parts[0] + ":" + Yellow + parts[1] + Reset)
		} else {
			fmt.Println(Purple + "[VERBOSE] " + Cyan + msg + Reset)
		}
	}
	logBuffer.WriteString("- [VERBOSE] " + msg + "\n")
}

func writeLog() {
	f, _ := os.Create("nUML_log.md")
	defer f.Close()
	f.WriteString("# nUML Generation Log\n\n")
	f.WriteString(fmt.Sprintf("Generated at: %s\n\n", time.Now().Format(time.RFC1123)))
	f.WriteString(logBuffer.String())
}

func isSwimlane(cell MxCell) bool {
	return strings.Contains(cell.Style, "swimlane")
}

func cleanHTML(s string) string {
	// 1. Basic HTML tag stripping
	// IMPORTANT: Only strip real tags, allow << for stereotypes
	s = strings.ReplaceAll(s, "<br>", " ")
	s = strings.ReplaceAll(s, "</div>", " ")
	s = strings.ReplaceAll(s, "</p>", " ")

	// Regex: Matches specific HTML tags (case insensitive)
	// p, div, span, i, b, em, strong, font
	re := regexp.MustCompile(`(?i)</?(div|p|span|i|b|em|strong|font)[^>]*>`)
	clean := re.ReplaceAllString(s, "")

	// 2. Decode HTML entities (properly handles &lt; &gt; &nbsp; etc)
	clean = html.UnescapeString(clean)

	// 3. Normalize whitespace
	clean = strings.ReplaceAll(clean, "\u00A0", " ")
	return strings.TrimSpace(clean)
}

func isFuzzyMatch(s, target string) bool {
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

	// Levenshtein-ish or simple heuristic
	// Check if it's "close enough"
	// Just handling common user typos for now
	if (s == "emun" && target == "enum") || (s == "enmu" && target == "enum") {
		return true
	}
	if (s == "interfac" || s == "inteface" || s == "iterface") && target == "interface" {
		return true
	}

	return false
}

func lowercaseFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func checkImports(cls *JavaClass) []string {
	imports := make(map[string]bool)

	// Check types
	checkType := func(t string) {
		if strings.Contains(t, "List") || strings.Contains(t, "Map") || strings.Contains(t, "Set") {
			imports["java.util.*"] = true
		}
		if strings.Contains(t, "LocalDate") || strings.Contains(t, "LocalTime") || strings.Contains(t, "Date") {
			imports["java.time.*"] = true
		}
	}

	for _, f := range cls.Fields {
		checkType(f.Type)
	}
	for _, m := range cls.Methods {
		checkType(m.ReturnType)
		checkType(m.Parameters)
	}

	var keys []string
	for k := range imports {
		keys = append(keys, k)
	}
	return keys
}
