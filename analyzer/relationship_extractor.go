package analyzer

import (
	"fmt"
	"nUML/models"
	"nUML/utils"
	"strings"
)

// RelationshipExtractor is responsible for identifying relationships between classes.
// RelationshipExtractor chịu trách nhiệm xác định các mối quan hệ giữa các lớp.
type RelationshipExtractor struct{}

// NewRelationshipExtractor creates a new instance of RelationshipExtractor.
// NewRelationshipExtractor tạo một phiên bản mới của RelationshipExtractor.
func NewRelationshipExtractor() *RelationshipExtractor {
	return &RelationshipExtractor{}
}

// Extract identifies relationships (extends, implements) from edges.
// Extract xác định các mối quan hệ (kế thừa, triển khai) từ các cạnh.
func (re *RelationshipExtractor) Extract(cells []models.MxCell, classes map[string]*models.ClassModel) {
	for _, cell := range cells {
		// Relationships (Edges)
		// Các mối quan hệ (Cạnh)
		if cell.Edge == "1" && cell.Source != "" && cell.Target != "" {
			sourceClass, sourceOk := classes[cell.Source]
			targetClass, targetOk := classes[cell.Target]

			if !sourceOk || !targetOk {
				continue
			} // Skip if either end is not a recognized class
			style := cell.Style

			// Determine intended relationship from style
			// Xác định mối quan hệ dự kiến từ kiểu
			isImplements := strings.Contains(style, "dashed=1")
			isExtends := strings.Contains(style, "endArrow=") // Broad check for solid arrows

			// Logic Verification / Auto-Correction
			// Xác minh logic / Tự động sửa lỗi
			if isExtends && !isImplements {
				// User drew "Extends". Check validity.
				// Người dùng vẽ "Extends". Kiểm tra tính hợp lệ.
				switch targetClass.Type {
				case models.Interface:
					// Class extends Interface -> ERROR. Should be Implements.
					// Lớp kế thừa Giao diện -> LỖI. Nên là Triển khai.
					sourceClass.Implements = append(sourceClass.Implements, targetClass.Name)
					utils.LogVerbose(fmt.Sprintf("Auto-Correct: %s implements %s (was extends)", sourceClass.Name, targetClass.Name))
				case models.Enum:
					// Class extends Enum -> ERROR. Impossible in Java.
					// Lớp kế thừa Enum -> LỖI. Không thể trong Java.
					// Ignore it.
					// Bỏ qua nó.
					utils.LogVerbose(fmt.Sprintf("Auto-Correct: Ignoring %s extends Enum %s", sourceClass.Name, targetClass.Name))
				default:
					// Class extends Class -> OK
					// Lớp kế thừa Lớp -> OK
					sourceClass.Extends = targetClass.Name
					utils.LogVerbose(fmt.Sprintf("Relationship: %s extends %s", sourceClass.Name, targetClass.Name))
				}
			} else if isImplements {
				// User drew "Implements".
				// Người dùng vẽ "Implements".
				if targetClass.Type != models.Interface {
					// Implements non-interface?
					// Triển khai cái không phải giao diện?
					// Maybe they meant extends if it's a class?
					// Có lẽ họ có ý định kế thừa nếu đó là một lớp?
					// Let's stick to valid Java: Only interfaces can be implemented.
					// Hãy tuân thủ Java hợp lệ: Chỉ các giao diện mới có thể được triển khai.
					if targetClass.Type == models.Class || targetClass.Type == models.Abstract {
						sourceClass.Extends = targetClass.Name
						utils.LogVerbose(fmt.Sprintf("Auto-Correct: %s extends %s (was implements)", sourceClass.Name, targetClass.Name))
					} else {
						// E.g. Enum? Cannot implement enum.
						// Ví dụ: Enum? Không thể triển khai enum.
						sourceClass.Implements = append(sourceClass.Implements, targetClass.Name)
						utils.LogVerbose(fmt.Sprintf("Relationship: %s implements %s", sourceClass.Name, targetClass.Name))
					}
				} else {
					sourceClass.Implements = append(sourceClass.Implements, targetClass.Name)
					utils.LogVerbose(fmt.Sprintf("Relationship: %s implements %s", sourceClass.Name, targetClass.Name))
				}
			}
		}
	}
}
