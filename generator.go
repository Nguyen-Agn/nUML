package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func generateJavaFile(cls *JavaClass) string {
	fileName := cls.Name + ".java"
	if targetPackage != "" {
		fileName = filepath.Join(targetPackage, fileName)
	}

	// Check Overwrite
	if !OverwriteMode {
		if _, err := os.Stat(fileName); err == nil {
			logVerbose(fmt.Sprintf("Skipped %s (exists, use -o to overwrite)", fileName))
			return fmt.Sprintf("# %s [Skipped]\n- File exists and -o not set.\n\n", cls.Name)
		}
	}

	f, err := os.Create(fileName)
	if err != nil {
		logInfo(fmt.Sprintf("Failed to create file %s: %v", fileName, err))
		return ""
	}
	defer f.Close()

	logVerbose(fmt.Sprintf("Generating class: %s", cls.Name))

	var sb strings.Builder

	// Tracking for Report
	var attrList []string
	var constructorCount int
	var getterList []string
	var setterList []string
	var inheritedList []string
	var customMethodList []string

	// Package Decl
	if targetPackage != "" {
		sb.WriteString("package " + targetPackage + ";\n\n")
	}

	// Imports
	imports := checkImports(cls)
	for _, imp := range imports {
		sb.WriteString("import " + imp + ";\n")
	}
	if len(imports) > 0 {
		sb.WriteString("\n")
	}

	// 1. Declaration
	access := "public"
	typeStr := "class"

	if cls.Type == Interface {
		typeStr = "interface"
	}
	if cls.Type == Enum {
		typeStr = "enum"
	}
	if cls.Type == Record {
		typeStr = "record"
	}
	if cls.Type == Abstract {
		typeStr = "abstract class"
	}

	if cls.Type == Record {
		// Record Syntax: public record Name(Type field1, Type field2) { ... }
		// Collect fields first
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
	if cls.Type == Enum {
		var constants []string
		var fields []Field

		for _, field := range cls.Fields {
			// Heuristic for constants vs fields in Enum
			if strings.Contains(field.Original, ":") || strings.HasPrefix(field.Original, "-") || strings.HasPrefix(field.Original, "#") || strings.HasPrefix(field.Original, "+") {
				fields = append(fields, field)
			} else {
				// Also sanitize constant names
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

	} else if cls.Type == Record {
		// Records don't list component fields inside body, only static/other fields
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
	} else if cls.Type != Interface {
		for _, field := range cls.Fields {
			// ... (Normal Class Field logic) ...
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
	for _, method := range cls.Methods {
		// Skip placeholders
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
			lowerName := strings.ToLower(method.Name)
			if strings.HasPrefix(lowerName, "get") {
				getterList = append(getterList, lowercaseFirst(strings.TrimPrefix(method.Name, "get"))) // Just storing the attribute name for report
			} else if strings.HasPrefix(lowerName, "set") {
				setterList = append(setterList, lowercaseFirst(strings.TrimPrefix(method.Name, "set")))
			} else {
				customMethodList = append(customMethodList, method.Name)
			}
		}

		mod := method.Visibility
		if cls.Type == Interface {
			mod = ""
		} else {
			if method.IsStatic {
				mod += " static"
			}
			if method.IsAbstract && cls.Type != Interface {
				mod += " abstract"
			}
		}

		// Params
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

		if cls.Type == Interface || method.IsAbstract {
			sb.WriteString(";\n\n")
		} else {
			sb.WriteString(" {\n")

			// Auto-Body
			if isConstructor {
				for _, pName := range paramNames {
					sb.WriteString(fmt.Sprintf("        this.%s = %s;\n", pName, pName))
				}
			} else if strings.HasPrefix(strings.ToLower(method.Name), "get") {
				fieldName := lowercaseFirst(strings.TrimPrefix(method.Name, "get"))
				sb.WriteString(fmt.Sprintf("        return %s;\n", fieldName))
			} else if strings.HasPrefix(strings.ToLower(method.Name), "set") {
				fieldName := lowercaseFirst(strings.TrimPrefix(method.Name, "set"))
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
	if needGettersSetters && cls.Type == Class {
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

	f.WriteString(sb.String())
	logVerbose(fmt.Sprintf("Generated %s", fileName))

	// Generate Markdown Report Segment
	var rpt strings.Builder
	rpt.WriteString(fmt.Sprintf("# %s [.]\n", cls.Name))

	// Attributes
	if len(attrList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo các thuộc tính: {%s}\n", strings.Join(attrList, ", ")))
	} else {
		rpt.WriteString("- [.] Đã tạo các thuộc tính: {}\n")
	}

	// Constructors
	if constructorCount > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo {%d} constructor\n", constructorCount))
	}

	// Getters
	if len(getterList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo getter cho: { %s }\n", strings.Join(getterList, ", ")))
	}

	// Setters
	if len(setterList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo setter cho: { %s }\n", strings.Join(setterList, ", ")))
	}

	// Inherited
	if len(inheritedList) > 0 {
		rpt.WriteString(fmt.Sprintf("- [.] Đã thừa kế (override) các phương thức: { %s }\n", strings.Join(inheritedList, ", ")))
	}

	// Custom Methods
	for _, cm := range customMethodList {
		rpt.WriteString(fmt.Sprintf("- [.] Đã tạo phương thức: %s\n", cm))
	}

	rpt.WriteString("\n")
	return rpt.String()
}
