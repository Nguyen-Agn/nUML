package generator

import (
	"fmt"
	"nUML/models"
	"nUML/utils"
	"path/filepath"
	"regexp"
	"strings"
)

// JavaGenerator implements CodeGenerator for Java.
// JavaGenerator triển khai CodeGenerator cho Java.
type JavaGenerator struct {
	TargetPackage string // The target package name // Tên gói đích
}

// NewJavaGenerator creates a new instance of JavaGenerator.
// NewJavaGenerator tạo một phiên bản mới của JavaGenerator.
func NewJavaGenerator(targetPackage string) *JavaGenerator {
	return &JavaGenerator{
		TargetPackage: targetPackage,
	}
}

// Generate produces Java code for a ClassModel.
// Generate tạo code Java cho một ClassModel.
func (jg *JavaGenerator) Generate(cls *models.ClassModel) (*GeneratedArtifact, error) {
	fileName := cls.Name + ".java"
	if jg.TargetPackage != "" {
		fileName = filepath.Join(jg.TargetPackage, fileName)
	}

	utils.LogVerbose(fmt.Sprintf("Generating class: %s", cls.Name))

	var sb strings.Builder
	var attrList []string
	var constructorCount int
	var getterList []string
	var setterList []string
	var inheritedList []string
	var customMethodList []string

	// Package Decl
	// Khai báo Gói
	if jg.TargetPackage != "" {
		sb.WriteString("package " + jg.TargetPackage + ";\n\n")
	}

	// Imports
	// Nhập khẩu (Imports)
	imports := jg.checkImports(cls)
	for _, imp := range imports {
		sb.WriteString("import " + imp + ";\n")
	}
	if len(imports) > 0 {
		sb.WriteString("\n")
	}

	// 1. Declaration
	// 1. Khai báo
	access := "public"
	typeStr := "class"

	if cls.Type == models.Interface {
		typeStr = "interface"
	}
	if cls.Type == models.Enum {
		typeStr = "enum"
	}
	if cls.Type == models.Record {
		typeStr = "record"
	}
	if cls.Type == models.Abstract {
		typeStr = "abstract class"
	}

	if cls.Type == models.Record {
		// Record Syntax: public record Name(Type field1, Type field2) { ... }
		// Cú pháp Record: public record Name(Type field1, Type field2) { ... }
		// Collect fields first
		// Thu thập các trường trước
		var recordComponents []string
		for _, f := range cls.Fields {
			if !f.IsStatic {
				recordComponents = append(recordComponents, fmt.Sprintf("%s %s", f.Type, f.Name))
				attrList = append(attrList, f.Name)
			}
		}
		sb.WriteString(fmt.Sprintf("%s record %s(%s)", access, cls.Name, strings.Join(recordComponents, ", ")))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s %s", access, typeStr, cls.Name))
	}

	if cls.Extends != "" {
		sb.WriteString(fmt.Sprintf(" extends %s", cls.Extends))
	}

	if len(cls.Implements) > 0 {
		sb.WriteString(fmt.Sprintf(" implements %s", strings.Join(cls.Implements, ", ")))
	}

	sb.WriteString(" {\n\n")

	needGettersSetters := false

	// 2. Fields & Enum Constants
	// 2. Các trường & Hằng số Enum
	if cls.Type == models.Enum {
		var constants []string
		var fields []models.Field

		for _, field := range cls.Fields {
			// Heuristic for constants vs fields in Enum
			// Quy tắc heuristic cho hằng số vs trường trong Enum
			if strings.Contains(field.Original, ":") || strings.HasPrefix(field.Original, "-") || strings.HasPrefix(field.Original, "#") || strings.HasPrefix(field.Original, "+") {
				fields = append(fields, field)
			} else {
				// Also sanitize constant names
				// Cũng làm sạch tên hằng số
				cName := strings.ReplaceAll(field.Name, " ", "")
				reValid := regexp.MustCompile(`[^a-zA-Z0-9_$]`)
				cName = reValid.ReplaceAllString(cName, "")
				constants = append(constants, cName)
			}
		}

		if len(constants) > 0 {
			sb.WriteString("    " + strings.Join(constants, ", ") + ";\n\n")
			attrList = append(attrList, constants...)
		}

		for _, field := range fields {
			mod := field.Visibility
			if field.IsStatic {
				mod += " static"
			}
			if field.IsFinal {
				mod += " final"
			}
			sb.WriteString(fmt.Sprintf("    %s %s %s;\n", mod, field.Type, field.Name))
			attrList = append(attrList, field.Name)
		}
		sb.WriteString("\n")

	} else if cls.Type == models.Record {
		// Records don't list component fields inside body, only static/other fields
		// Bản ghi không liệt kê các trường thành phần bên trong thân, chỉ các trường tĩnh/khác
		for _, field := range cls.Fields {
			if field.IsStatic {
				mod := field.Visibility
				mod += " static"
				if field.IsFinal {
					mod += " final"
				}
				sb.WriteString(fmt.Sprintf("    %s %s %s;\n", mod, field.Type, field.Name))
				attrList = append(attrList, field.Name)
			}
		}
		sb.WriteString("\n")
	} else if cls.Type != models.Interface {
		for _, field := range cls.Fields {
			// ... (Normal Class Field logic) ...
			// ... (Logic trường lớp thông thường) ...
			if strings.Contains(strings.ToLower(field.Original), "getters/setters") {
				needGettersSetters = true
				continue
			}
			if strings.Contains(field.Name, "(") {
				continue
			}

			attrList = append(attrList, field.Name)

			mod := field.Visibility
			if field.IsStatic {
				mod += " static"
			}
			if field.IsFinal {
				mod += " final"
			}

			// Initialization
			// Khởi tạo
			initStr := ""
			if field.IsStatic && field.IsFinal {
				if field.InitialValue != "" {
					initStr = fmt.Sprintf(" = %s", field.InitialValue)
				} else {
					switch field.Type {
					case "int", "long", "short", "byte":
						initStr = " = 0"
					case "double", "float":
						initStr = " = 0.0"
					case "boolean":
						initStr = " = false"
					case "char":
						initStr = " = '\\u0000'"
					case "String":
						initStr = " = \"\""
					default:
						initStr = " = null"
					}
				}
			} else if field.InitialValue != "" {
				initStr = fmt.Sprintf(" = %s", field.InitialValue)
			}

			sb.WriteString(fmt.Sprintf("    %s %s %s%s;\n", mod, field.Type, field.Name, initStr))
		}
		sb.WriteString("\n")
	}

	// 3. Methods
	// 3. Các phương thức
	for _, method := range cls.Methods {
		// Skip placeholders
		// Bỏ qua phần giữ chỗ
		if strings.Contains(strings.ToLower(method.Original), "getters/setters") {
			needGettersSetters = true
			continue
		}

		isConstructor := method.Name == cls.Name
		if isConstructor {
			constructorCount++
		}

		if method.IsOverride {
			sb.WriteString("    @Override\n")
			inheritedList = append(inheritedList, method.Name)
		} else if !isConstructor {
			// Heuristic for getters/setters/custom
			// Heuristic cho getters/setters/tùy chỉnh
			lowerName := strings.ToLower(method.Name)
			if strings.HasPrefix(lowerName, "get") {
				getterList = append(getterList, utils.LowercaseFirst(strings.TrimPrefix(method.Name, "get")))
			} else if strings.HasPrefix(lowerName, "set") {
				setterList = append(setterList, utils.LowercaseFirst(strings.TrimPrefix(method.Name, "set")))
			} else {
				customMethodList = append(customMethodList, method.Name)
			}
		}

		mod := method.Visibility
		if cls.Type == models.Interface {
			mod = ""
		} else {
			if method.IsStatic {
				mod += " static"
			}
			if method.IsAbstract && cls.Type != models.Interface {
				mod += " abstract"
			}
		}

		// Params
		// Các tham số
		rawParams := strings.Split(method.Parameters, ",")
		var javaParams []string
		var paramNames []string

		for _, p := range rawParams {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			parts := strings.Split(p, ":")
			if len(parts) == 2 {
				pName := strings.TrimSpace(parts[0])
				pType := strings.TrimSpace(parts[1])
				javaParams = append(javaParams, fmt.Sprintf("%s %s", pType, pName))
				paramNames = append(paramNames, pName)
			} else {
				javaParams = append(javaParams, p)
			}
		}
		paramStr := strings.Join(javaParams, ", ")

		if isConstructor {
			sb.WriteString(fmt.Sprintf("    %s %s(%s)", mod, method.Name, paramStr))
		} else {
			if mod == "default" {
				sb.WriteString(fmt.Sprintf("    default %s %s(%s)", method.ReturnType, method.Name, paramStr))
			} else if mod != "" {
				sb.WriteString(fmt.Sprintf("    %s %s %s(%s)", strings.TrimSpace(mod), method.ReturnType, method.Name, paramStr))
			} else {
				sb.WriteString(fmt.Sprintf("    %s %s(%s)", method.ReturnType, method.Name, paramStr))
			}
		}

		if cls.Type == models.Interface || method.IsAbstract {
			sb.WriteString(";\n\n")
		} else {
			sb.WriteString(" {\n")

			// Auto-Body
			// Tự động tạo thân hàm
			if isConstructor {
				for _, pName := range paramNames {
					sb.WriteString(fmt.Sprintf("        this.%s = %s;\n", pName, pName))
				}
			} else if strings.HasPrefix(strings.ToLower(method.Name), "get") {
				fieldName := utils.LowercaseFirst(strings.TrimPrefix(method.Name, "get"))
				sb.WriteString(fmt.Sprintf("        return %s;\n", fieldName))
			} else if strings.HasPrefix(strings.ToLower(method.Name), "set") {
				fieldName := utils.LowercaseFirst(strings.TrimPrefix(method.Name, "set"))
				if len(paramNames) > 0 {
					sb.WriteString(fmt.Sprintf("        this.%s = %s;\n", fieldName, paramNames[0]))
				}
			} else if method.ReturnType != "void" && method.ReturnType != "" {
				sb.WriteString("        //Add your code here\n")
				ret := "null"
				if method.ReturnType == "int" || method.ReturnType == "double" {
					ret = "0"
				}
				if method.ReturnType == "boolean" {
					ret = "false"
				}
				sb.WriteString(fmt.Sprintf("        return %s;\n", ret))
			} else {
				// Void
				sb.WriteString("        //Add your code here\n")
			}

			sb.WriteString("    }\n\n")
		}
	}

	// 4. Auto-Generate Getters/Setters
	// 4. Tự động tạo Getters/Setters
	if needGettersSetters && cls.Type == models.Class {
		for _, field := range cls.Fields {
			if strings.Contains(strings.ToLower(field.Original), "getters/setters") {
				continue
			}
			if field.IsStatic {
				continue
			}

			// Getter
			uName := strings.ToUpper(field.Name[:1]) + field.Name[1:]
			sb.WriteString(fmt.Sprintf("    public %s get%s() {\n", field.Type, uName))
			sb.WriteString(fmt.Sprintf("        return %s;\n", field.Name))
			sb.WriteString("    }\n\n")
			getterList = append(getterList, field.Name)

			// Setter
			sb.WriteString(fmt.Sprintf("    public void set%s(%s %s) {\n", uName, field.Type, field.Name))
			sb.WriteString(fmt.Sprintf("        this.%s = %s;\n", field.Name, field.Name))
			sb.WriteString("    }\n\n")
			setterList = append(setterList, field.Name)
		}
	}

	sb.WriteString("}\n")

	// Generate Report
	// Tạo báo cáo
	var rpt strings.Builder
	rpt.WriteString(fmt.Sprintf("# %s [.]\n", cls.Name))

	// Attributes
	// Thuộc tính
	if len(attrList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo các thuộc tính (Created attributes): {%s}\n", strings.Join(attrList, ", ")))
	} else {
		rpt.WriteString("- [.] Đã tạo các thuộc tính (Created attributes): {}\n")
	}

	// Constructors
	// Hàm khởi tạo
	if constructorCount > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo {%d} constructor(s)\n", constructorCount))
	}

	// Getters
	if len(getterList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo getter cho (Created getters for): { %s }\n", strings.Join(getterList, ", ")))
	}

	// Setters
	if len(setterList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo setter cho (Created setters for): { %s }\n", strings.Join(setterList, ", ")))
	}

	// Inherited
	// Kế thừa
	if len(inheritedList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã thừa kế (override) các phương thức (Overridden methods): { %s }\n", strings.Join(inheritedList, ", ")))
	}

	// Custom Methods
	// Phương thức tùy chỉnh
	for _, cm := range customMethodList {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo phương thức (Created method): %s\n", cm))
	}

	rpt.WriteString("\n")

	return &GeneratedArtifact{
		FileName:    fileName,
		Content:     sb.String(),
		ReportEntry: rpt.String(),
	}, nil
}

// checkImports identifies necessary imports for the class.
// checkImports xác định các mục nhập khẩu cần thiết cho lớp.
func (jg *JavaGenerator) checkImports(cls *models.ClassModel) []string {
	imports := make(map[string]bool)

	// Check types
	// Kiểm tra các kiểu
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
