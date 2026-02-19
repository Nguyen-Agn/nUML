package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func ParseXML(inputFile string) ([]MxCell, error) {
	byteValue, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var mxFile MxFile
	if err := xml.Unmarshal(byteValue, &mxFile); err != nil {
		return nil, fmt.Errorf("error parsing XML: %v", err)
	}

	return mxFile.Diagram.MxGraphModel.Root.MxCells, nil
}

func parseClassNameAndType(raw string) (string, ClassType) {
	cType := Class

	// Check tags in RAW string for Abstract (Italics)
	if strings.Contains(raw, "<i>") || strings.Contains(raw, "<em>") {
		cType = Abstract
	}

	clean := cleanHTML(raw)
	// fmt.Printf("DEBUG: Raw='%s', Clean='%s'\n", raw, clean)

	// Robust Stereotype Parsing
	// 1. Look for stereotypes like <<Enum>>, «Interface»
	reStereo := regexp.MustCompile(`(<<|«)\s*(\w+)\s*(>>|»)`)
	match := reStereo.FindStringSubmatch(clean)

	if len(match) > 2 {
		tag := strings.ToLower(match[2])
		if tag == "interface" || isFuzzyMatch(tag, "interface") {
			cType = Interface
		} else if tag == "enum" || isFuzzyMatch(tag, "enum") {
			cType = Enum
		} else if tag == "record" || isFuzzyMatch(tag, "record") {
			cType = Record
		}
		// Remove stereotype from name
		clean = reStereo.ReplaceAllString(clean, "")
	} else {
		// Fallback: Fuzzy keyword detection
		lowerClean := strings.ToLower(clean)
		reWords := regexp.MustCompile(`\W+`)
		words := reWords.Split(lowerClean, -1)

		for _, w := range words {
			if w == "interface" || isFuzzyMatch(w, "interface") {
				cType = Interface
			}
			if w == "enum" || isFuzzyMatch(w, "enum") {
				cType = Enum
			}
			if w == "record" || isFuzzyMatch(w, "record") {
				cType = Record
			}
			if w == "abstract" && cType != Abstract {
				cType = Abstract
			}
		}
	}

	// Extraction of Name: Aggressive Cleaning
	// Remove keywords if they are floating around
	reKeywords := regexp.MustCompile(`(?i)\b(interface|enum|record|abstract|class)\b`)
	clean = reKeywords.ReplaceAllString(clean, "")

	// Strip invalid chars
	reValid := regexp.MustCompile(`[^a-zA-Z0-9_$]`)
	name := reValid.ReplaceAllString(clean, "")

	// fmt.Printf("DEBUG: Final Name='%s', Type='%s'\n", name, cType)
	return name, cType
}

func parseField(val string) Field {
	f := Field{Original: val}

	// Modifiers
	if strings.HasPrefix(val, "+") {
		f.Visibility = "public"
	} else if strings.HasPrefix(val, "-") {
		f.Visibility = "private"
	} else if strings.HasPrefix(val, "#") {
		f.Visibility = "protected"
	} else {
		f.Visibility = "private"
	} // Default

	// Advanced Modifiers
	if strings.Contains(val, "static") {
		f.IsStatic = true
	}
	if strings.ToUpper(val) == val && len(val) > 0 { // Simple heuristic
		f.IsFinal = true
	}
	if strings.Contains(val, "final") {
		f.IsFinal = true
	}

	// Cleanup
	cleanVal := val
	cleanVal = strings.TrimLeft(cleanVal, "+-# ")
	cleanVal = strings.ReplaceAll(cleanVal, "static", "")
	cleanVal = strings.ReplaceAll(cleanVal, "final", "")
	cleanVal = strings.TrimSpace(cleanVal)

	// name: Type = Value
	// First, check for initialization
	initParts := strings.SplitN(cleanVal, "=", 2)
	if len(initParts) == 2 {
		f.InitialValue = strings.TrimSpace(initParts[1])
		cleanVal = strings.TrimSpace(initParts[0])
	}

	parts := strings.Split(cleanVal, ":")
	if len(parts) >= 2 {
		f.Name = strings.TrimSpace(parts[0])
		f.Type = strings.TrimSpace(parts[1])
	} else {
		partsSpace := strings.Fields(cleanVal)
		if len(partsSpace) >= 2 {
			f.Type = partsSpace[0]
			f.Name = partsSpace[1]
		} else {
			f.Name = cleanVal
			f.Type = "String" // Default
		}
	}

	// Sanitize Name
	f.Name = strings.ReplaceAll(f.Name, " ", "")
	reValid := regexp.MustCompile(`[^a-zA-Z0-9_$]`)
	f.Name = reValid.ReplaceAllString(f.Name, "")

	return f
}

func parseMethod(rawVal string) Method {
	m := Method{Original: rawVal}
	// Check Abstract (Italics) in RAW HTML
	if strings.Contains(rawVal, "<i>") || strings.Contains(rawVal, "<em>") {
		m.IsAbstract = true
	}

	val := cleanHTML(rawVal)

	// Modifiers
	if strings.Contains(val, "default") {
		// Special case for Interface Default Methods
		m.Visibility = "default"
	} else if strings.HasPrefix(val, "+") {
		m.Visibility = "public"
	} else if strings.HasPrefix(val, "-") {
		m.Visibility = "private"
	} else if strings.HasPrefix(val, "#") {
		m.Visibility = "protected"
	} else {
		m.Visibility = "public"
	}

	if strings.Contains(rawVal, "static") { // Check in raw too just in case literal word
		m.IsStatic = true
	} else if strings.Contains(val, "static") {
		m.IsStatic = true
	}

	if strings.Contains(val, "abstract") {
		m.IsAbstract = true
	}

	cleanVal := val
	cleanVal = strings.ReplaceAll(cleanVal, "\\", "")
	cleanVal = strings.TrimLeft(cleanVal, "+-# ")
	cleanVal = strings.ReplaceAll(cleanVal, "static", "")
	cleanVal = strings.ReplaceAll(cleanVal, "abstract", "")
	// Don't strip "default" here, or handle it carefully
	if m.Visibility == "default" {
		cleanVal = strings.ReplaceAll(cleanVal, "default", "")
	}
	cleanVal = strings.TrimSpace(cleanVal)

	// Check for Return Type: ": Type" AFTER the closing parenthesis
	lastParen := strings.LastIndex(cleanVal, ")")

	if lastParen != -1 {
		afterParen := cleanVal[lastParen+1:]
		if strings.Contains(afterParen, ":") {
			parts := strings.SplitN(afterParen, ":", 2)
			if len(parts) > 1 {
				m.ReturnType = strings.TrimSpace(parts[1])
			}
		}

		if m.ReturnType == "" {
			m.ReturnType = "void"
		}

		lhs := cleanVal[:lastParen+1]
		parenStart := strings.Index(lhs, "(")

		if parenStart != -1 {
			m.Name = strings.TrimSpace(lhs[:parenStart])
			m.Parameters = lhs[parenStart+1 : lastParen]
		} else {
			m.Name = lhs
		}
	} else {
		m.Name = cleanVal
		m.ReturnType = "void"
	}

	m.Name = strings.ReplaceAll(m.Name, " ", "")
	reValid := regexp.MustCompile(`[^a-zA-Z0-9_$]`)
	m.Name = reValid.ReplaceAllString(m.Name, "")

	return m
}
