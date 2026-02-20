package analyzer

import (
	"fmt"
	"nUML/models"
	"nUML/utils"
	"regexp"
	"strings"
)

// ClassExtractor is responsible for identifying classes from the diagram.
// ClassExtractor chịu trách nhiệm xác định các lớp từ biểu đồ.
type ClassExtractor struct{}

// NewClassExtractor creates a new instance of ClassExtractor.
// NewClassExtractor tạo một phiên bản mới của ClassExtractor.
func NewClassExtractor() *ClassExtractor {
	return &ClassExtractor{}
}

// Extract identifies classes (swimlanes) from the list of cells.
// Extract xác định các lớp (swimlanes) từ danh sách các ô.
func (ce *ClassExtractor) Extract(cells []models.MxCell) map[string]*models.ClassModel {
	classes := make(map[string]*models.ClassModel)

	for _, cell := range cells {
		if strings.Contains(cell.Style, "swimlane") {
			rawName := utils.CleanHTML(cell.Value)
			name, classType := ce.parseClassNameAndType(cell.Value) // Pass RAW for abstract detection

			classes[cell.ID] = &models.ClassModel{
				ID:      cell.ID,
				Name:    name,
				RawName: rawName,
				Type:    classType,
			}
			utils.LogVerbose(fmt.Sprintf("Found %s: %s", classType, name))
		}
	}
	return classes
}

// parseClassNameAndType parses the class name and determined its type (Class, Interface, etc.).
// parseClassNameAndType phân tích tên lớp và xác định loại của nó (Lớp, Giao diện, v.v.).
func (ce *ClassExtractor) parseClassNameAndType(raw string) (string, models.ClassType) {
	cType := models.Class

	// Check tags in RAW string for Abstract (Italics)
	// Kiểm tra các thẻ trong chuỗi RAW cho Trừu tượng (Italics)
	if strings.Contains(raw, "<i>") || strings.Contains(raw, "<em>") {
		cType = models.Abstract
	}

	clean := utils.CleanHTML(raw)

	// Robust Stereotype Parsing
	// Phân tích cú pháp khuôn mẫu (Stereotype) mạnh mẽ
	// 1. Look for stereotypes like <<Enum>>, «Interface»
	// 1. Tìm các khuôn mẫu như <<Enum>>, «Interface»
	reStereo := regexp.MustCompile(`(<<|«)\s*(\w+)\s*(>>|»)`)
	match := reStereo.FindStringSubmatch(clean)

	if len(match) > 2 {
		tag := strings.ToLower(match[2])
		if tag == "interface" || utils.IsFuzzyMatch(tag, "interface") {
			cType = models.Interface
		} else if tag == "enum" || utils.IsFuzzyMatch(tag, "enum") {
			cType = models.Enum
		} else if tag == "record" || utils.IsFuzzyMatch(tag, "record") {
			cType = models.Record
		}
		// Remove stereotype from name
		// Loại bỏ khuôn mẫu khỏi tên
		clean = reStereo.ReplaceAllString(clean, "")
	} else {
		// Fallback: Fuzzy keyword detection
		// Dự phòng: Phát hiện từ khóa mờ
		lowerClean := strings.ToLower(clean)
		reWords := regexp.MustCompile(`\W+`)
		words := reWords.Split(lowerClean, -1)

		for _, w := range words {
			if w == "interface" || utils.IsFuzzyMatch(w, "interface") {
				cType = models.Interface
			}
			if w == "enum" || utils.IsFuzzyMatch(w, "enum") {
				cType = models.Enum
			}
			if w == "record" || utils.IsFuzzyMatch(w, "record") {
				cType = models.Record
			}
			if w == "abstract" && cType != models.Abstract {
				cType = models.Abstract
			}
		}
	}

	// Extraction of Name: Aggressive Cleaning
	// Trích xuất Tên: Làm sạch tích cực
	// Remove keywords if they are floating around
	// Loại bỏ từ khóa nếu chúng trôi nổi xung quanh
	reKeywords := regexp.MustCompile(`(?i)\b(interface|enum|record|abstract|class)\b`)
	clean = reKeywords.ReplaceAllString(clean, "")

	// Strip invalid chars
	// Loại bỏ các ký tự không hợp lệ
	reValid := regexp.MustCompile(`[^a-zA-Z0-9_$]`)
	name := reValid.ReplaceAllString(clean, "")

	return name, cType
}
