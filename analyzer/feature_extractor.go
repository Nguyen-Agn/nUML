package analyzer

import (
	"nUML/models"
	"nUML/utils"
	"regexp"
	"strings"
)

// FeatureExtractor is responsible for extracting fields and methods from the diagram.
// FeatureExtractor chịu trách nhiệm trích xuất các trường và phương thức từ biểu đồ.
type FeatureExtractor struct{}

// NewFeatureExtractor creates a new instance of FeatureExtractor.
// NewFeatureExtractor tạo một phiên bản mới của FeatureExtractor.
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

// Extract identifies fields and methods inside swimlanes.
// Extract xác định các trường và phương thức bên trong swimlanes.
func (fe *FeatureExtractor) Extract(cells []models.MxCell, classes map[string]*models.ClassModel) {
	for _, cell := range cells {
		// Items inside swimlanes
		// Các mục bên trong swimlanes
		if cell.Parent != "" {
			parentClass, ok := classes[cell.Parent]
			if ok {
				// It's a field or method or separator
				// Nó là một trường hoặc phương thức hoặc dấu phân cách
				val := utils.CleanHTML(cell.Value)
				rawVal := cell.Value
				if val == "" {
					continue
				}

				// Identify if Method or Field based on parenthesis
				// Xác định xem là Phương thức hay Trường dựa trên dấu ngoặc đơn
				if strings.Contains(val, "(") && strings.Contains(val, ")") {
					// Method
					m := fe.parseMethod(rawVal) // Pass RAW for italics check
					if parentClass.Type == models.Interface {
						m.IsAbstract = true
						m.Visibility = "public" // Force public for interface
					}
					parentClass.Methods = append(parentClass.Methods, m)
				} else {
					// Field (if not just a line separator)
					// Trường (nếu không chỉ là dòng phân cách)
					if !strings.Contains(cell.Style, "line") {
						f := fe.parseField(val)
						parentClass.Fields = append(parentClass.Fields, f)
					}
				}
			}
		}
	}
}

// parseField parses a string into a Field struct.
// parseField phân tích một chuỗi thành cấu trúc Field.
func (fe *FeatureExtractor) parseField(val string) models.Field {
	f := models.Field{Original: val}

	// Modifiers
	// Các từ khóa sửa đổi
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
	// Các từ khóa sửa đổi nâng cao
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
	// Làm sạch
	cleanVal := val
	cleanVal = strings.TrimLeft(cleanVal, "+-# ")
	cleanVal = strings.ReplaceAll(cleanVal, "static", "")
	cleanVal = strings.ReplaceAll(cleanVal, "final", "")
	cleanVal = strings.TrimSpace(cleanVal)

	// name: Type = Value
	// First, check for initialization
	// Đầu tiên, kiểm tra khởi tạo
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
	// Làm sạch Tên
	f.Name = strings.ReplaceAll(f.Name, " ", "")
	reValid := regexp.MustCompile(`[^a-zA-Z0-9_$]`)
	f.Name = reValid.ReplaceAllString(f.Name, "")

	return f
}

// parseMethod parses a string into a Method struct.
// parseMethod phân tích một chuỗi thành cấu trúc Method.
func (fe *FeatureExtractor) parseMethod(rawVal string) models.Method {
	m := models.Method{Original: rawVal}
	// Check Abstract (Italics) in RAW HTML
	// Kiểm tra Trừu tượng (Italics) trong HTML thô
	if strings.Contains(rawVal, "<i>") || strings.Contains(rawVal, "<em>") {
		m.IsAbstract = true
	}

	val := utils.CleanHTML(rawVal)

	// Modifiers
	// Các từ khóa sửa đổi
	if strings.Contains(val, "default") {
		// Special case for Interface Default Methods
		// Trường hợp đặc biệt cho các phương thức mặc định của Giao diện
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
	// Kiểm tra Kiểu trả về: ": Type" SAU dấu ngoặc đóng
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
