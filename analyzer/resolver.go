package analyzer

import (
	"fmt"
	"nUML/models"
	"nUML/utils"
)

// HierarchyResolver is responsible for resolving inheritance and method overrides.
// HierarchyResolver chịu trách nhiệm giải quyết việc kế thừa và ghi đè phương thức.
type HierarchyResolver struct{}

// NewHierarchyResolver creates a new instance of HierarchyResolver.
// NewHierarchyResolver tạo một phiên bản mới của HierarchyResolver.
func NewHierarchyResolver() *HierarchyResolver {
	return &HierarchyResolver{}
}

// Resolve handles method inheritance and auto-overriding from abstract classes and interfaces.
// Resolve xử lý việc kế thừa phương thức và tự động ghi đè từ các lớp trừu tượng và giao diện.
func (hr *HierarchyResolver) Resolve(classes map[string]*models.ClassModel) {
	// Helper to find class by Name (since relationships use Name, not ID)
	// Trình trợ giúp để tìm lớp theo Tên (vì các mối quan hệ sử dụng Tên, không phải ID)
	nameToClass := make(map[string]*models.ClassModel)
	for _, cls := range classes {
		nameToClass[cls.Name] = cls
	}

	for _, cls := range classes {
		if cls.Type == models.Interface || cls.Type == models.Enum {
			continue
		}

		// 1. Check Extends (Abstract Class)
		// 1. Kiểm tra Extends (Lớp trừu tượng)
		if cls.Extends != "" {
			parent, ok := nameToClass[cls.Extends]
			if ok && parent.Type == models.Abstract {
				// Inherit abstract methods
				// Kế thừa các phương thức trừu tượng
				for _, pm := range parent.Methods {
					if pm.IsAbstract {
						// Check if cls already has it
						// Kiểm tra xem cls đã có nó chưa
						hasIt := false
						for _, cm := range cls.Methods {
							if cm.Name == pm.Name { // Simple name check for now
								hasIt = true
								break
							}
						}
						if !hasIt {
							// Add stub
							// Thêm stub
							newM := pm
							newM.IsAbstract = false // Concrete implementation
							newM.IsOverride = true
							cls.Methods = append(cls.Methods, newM)
							utils.LogVerbose(fmt.Sprintf("Auto-Override: %s inherits %s from %s", cls.Name, pm.Name, parent.Name))
						}
					}
				}
			}
		}

		// 2. Check Implements (Interfaces)
		// 2. Kiểm tra Implements (Giao diện)
		for _, impl := range cls.Implements {
			iface, ok := nameToClass[impl]
			if ok && iface.Type == models.Interface {
				for _, im := range iface.Methods {
					// Check if cls already has it
					// Kiểm tra xem cls đã có nó chưa
					hasIt := false
					for _, cm := range cls.Methods {
						if cm.Name == im.Name {
							hasIt = true
							break
						}
					}

					// If Interface method is DEFAULT, we don't *need* to override, but user said "smart create".
					// Nếu phương thức Giao diện là DEFAULT, chúng ta không *cần* phải ghi đè, nhưng người dùng đã nói "tạo thông minh".
					// Actually, if it's default, we usually DON'T force override unless requested.
					// Thực ra, nếu là default, chúng ta thường KHÔNG ép buộc ghi đè trừ khi được yêu cầu.
					// But if it's abstract (no default), we MUST.
					// Nhưng nếu nó là trừu tượng (không có default), chúng ta PHẢI ghi đè.
					// The parser sets Visibility="default" for default methods.
					// Trình phân tích cú pháp đặt Visibility="default" cho các phương thức mặc định.
					isDefault := im.Visibility == "default"

					if !hasIt && !isDefault {
						// Add stub
						// Thêm stub
						newM := im
						newM.IsAbstract = false
						newM.IsOverride = true
						// Ensure public visibility for interface impl
						// Đảm bảo phạm vi truy cập public cho việc triển khai giao diện
						newM.Visibility = "public"
						cls.Methods = append(cls.Methods, newM)
						utils.LogVerbose(fmt.Sprintf("Auto-Implements: %s implements %s from %s", cls.Name, im.Name, iface.Name))
					}
				}
			}
		}
	}
}
