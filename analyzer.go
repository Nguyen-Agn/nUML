package main

import (
	"fmt"
	"strings"
)

func AnalyzeDiagram(cells []MxCell) map[string]*JavaClass {
	classes := make(map[string]*JavaClass)

	// cell ID -> Parent ID (Swimlane) mapping for attributes
	itemParents := make(map[string]string)

	// 2. Pass 1: Identify Classes (Swimlanes)
	for _, cell := range cells {
		if isSwimlane(cell) {
			rawName := cleanHTML(cell.Value)
			name, classType := parseClassNameAndType(cell.Value) // Pass RAW for abstract detection

			classes[cell.ID] = &JavaClass{
				ID:      cell.ID,
				Name:    name,
				RawName: rawName,
				Type:    classType,
			}
			logVerbose(fmt.Sprintf("Found %s: %s", classType, name))
		}
	}

	// 3. Pass 2: Identify Attributes, Methods and Relationships
	for _, cell := range cells {
		// Relationships (Edges)
		if cell.Edge == "1" && cell.Source != "" && cell.Target != "" {
			sourceClass, sourceOk := classes[cell.Source]
			targetClass, targetOk := classes[cell.Target]

			if sourceOk && targetOk {
				style := cell.Style

				// Determine intended relationship from style
				isImplements := strings.Contains(style, "dashed=1")
				isExtends := strings.Contains(style, "endArrow=") // Broad check for solid arrows

				// Logic Verification / Auto-Correction
				if isExtends && !isImplements {
					// User drew "Extends". Check validity.
					if targetClass.Type == Interface {
						// Class extends Interface -> ERROR. Should be Implements.
						sourceClass.Implements = append(sourceClass.Implements, targetClass.Name)
						logVerbose(fmt.Sprintf("Auto-Correct: %s implements %s (was extends)", sourceClass.Name, targetClass.Name))
					} else if targetClass.Type == Enum {
						// Class extends Enum -> ERROR. Impossible in Java.
						// Ignore it.
						logVerbose(fmt.Sprintf("Auto-Correct: Ignoring %s extends Enum %s", sourceClass.Name, targetClass.Name))
					} else {
						// Class extends Class -> OK
						sourceClass.Extends = targetClass.Name
						logVerbose(fmt.Sprintf("Relationship: %s extends %s", sourceClass.Name, targetClass.Name))
					}
				} else if isImplements {
					// User drew "Implements".
					if targetClass.Type != Interface {
						// Implements non-interface?
						// Maybe they meant extends if it's a class?
						// Let's stick to valid Java: Only interfaces can be implemented.
						if targetClass.Type == Class || targetClass.Type == Abstract {
							sourceClass.Extends = targetClass.Name
							logVerbose(fmt.Sprintf("Auto-Correct: %s extends %s (was implements)", sourceClass.Name, targetClass.Name))
						} else {
							// E.g. Enum? Cannot implement enum.
							sourceClass.Implements = append(sourceClass.Implements, targetClass.Name)
							logVerbose(fmt.Sprintf("Relationship: %s implements %s", sourceClass.Name, targetClass.Name))
						}
					} else {
						sourceClass.Implements = append(sourceClass.Implements, targetClass.Name)
						logVerbose(fmt.Sprintf("Relationship: %s implements %s", sourceClass.Name, targetClass.Name))
					}
				}
			}
			continue
		}

		// Items inside swimlanes
		if cell.Parent != "" {
			itemParents[cell.ID] = cell.Parent

			parentClass, ok := classes[cell.Parent]
			if ok {
				// It's a field or method or separator
				val := cleanHTML(cell.Value)
				rawVal := cell.Value
				if val == "" {
					continue
				}

				// Identify if Method or Field based on parenthesis
				if strings.Contains(val, "(") && strings.Contains(val, ")") {
					// Method
					m := parseMethod(rawVal) // Pass RAW for italics check
					if parentClass.Type == Interface {
						m.IsAbstract = true
						m.Visibility = "public" // Force public for interface
					}
					parentClass.Methods = append(parentClass.Methods, m)
				} else {
					// Field (if not just a line separator)
					if !strings.Contains(cell.Style, "line") {
						f := parseField(val)
						parentClass.Fields = append(parentClass.Fields, f)
					}
				}
			}
		}
	}

	// 4. Pass 3: Resolve Inheritance (Auto-Override)
	resolveInheritance(classes)

	return classes
}

func resolveInheritance(classes map[string]*JavaClass) {
	// Helper to find class by Name (since relationships use Name, not ID)
	nameToClass := make(map[string]*JavaClass)
	for _, cls := range classes {
		nameToClass[cls.Name] = cls
	}

	for _, cls := range classes {
		if cls.Type == Interface || cls.Type == Enum {
			continue
		}

		// 1. Check Extends (Abstract Class)
		if cls.Extends != "" {
			parent, ok := nameToClass[cls.Extends]
			if ok && parent.Type == Abstract {
				// Inherit abstract methods
				for _, pm := range parent.Methods {
					if pm.IsAbstract {
						// Check if cls already has it
						hasIt := false
						for _, cm := range cls.Methods {
							if cm.Name == pm.Name { // Simple name check for now
								hasIt = true
								break
							}
						}
						if !hasIt {
							// Add stub
							newM := pm
							newM.IsAbstract = false // Concrete implementation
							newM.IsOverride = true
							cls.Methods = append(cls.Methods, newM)
							logVerbose(fmt.Sprintf("Auto-Override: %s inherits %s from %s", cls.Name, pm.Name, parent.Name))
						}
					}
				}
			}
		}

		// 2. Check Implements (Interfaces)
		for _, impl := range cls.Implements {
			iface, ok := nameToClass[impl]
			if ok && iface.Type == Interface {
				for _, im := range iface.Methods {
					// Check if cls already has it
					hasIt := false
					for _, cm := range cls.Methods {
						if cm.Name == im.Name {
							hasIt = true
							break
						}
					}

					// If Interface method is DEFAULT, we don't *need* to override, but user said "smart create".
					// Actually, if it's default, we usually DON'T force override unless requested.
					// But if it's abstract (no default), we MUST.
					// The parser sets Visibility="default" for default methods.
					isDefault := im.Visibility == "default"

					if !hasIt && !isDefault {
						// Add stub
						newM := im
						newM.IsAbstract = false
						newM.IsOverride = true
						// Ensure public visibility for interface impl
						newM.Visibility = "public"
						cls.Methods = append(cls.Methods, newM)
						logVerbose(fmt.Sprintf("Auto-Implements: %s implements %s from %s", cls.Name, im.Name, iface.Name))
					}
				}
			}
		}
	}
}
