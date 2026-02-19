package main

import "encoding/xml"

// XML Structures
type MxFile struct {
	XMLName xml.Name `xml:"mxfile"`
	Diagram Diagram  `xml:"diagram"`
}

type Diagram struct {
	MxGraphModel MxGraphModel `xml:"mxGraphModel"`
}

type MxGraphModel struct {
	Root Root `xml:"root"`
}

type Root struct {
	MxCells []MxCell `xml:"mxCell"`
}

type MxCell struct {
	ID       string     `xml:"id,attr"`
	Parent   string     `xml:"parent,attr"`
	Value    string     `xml:"value,attr"`
	Style    string     `xml:"style,attr"`
	Vertex   string     `xml:"vertex,attr"`
	Edge     string     `xml:"edge,attr"`
	Source   string     `xml:"source,attr"`
	Target   string     `xml:"target,attr"`
	Geometry MxGeometry `xml:"mxGeometry"`
}

type MxGeometry struct {
	X      string `xml:"x,attr"`
	Y      string `xml:"y,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

// Internal Models
type ClassType string

const (
	Class     ClassType = "class"
	Interface ClassType = "interface"
	Enum      ClassType = "enum"
	Record    ClassType = "record"
	Abstract  ClassType = "abstract"
)

type Field struct {
	Original     string
	Name         string
	Type         string
	Visibility   string
	IsStatic     bool
	IsFinal      bool
	InitialValue string // New: Value after =
}

type Method struct {
	Original   string
	Name       string
	Parameters string
	ReturnType string
	Visibility string
	IsStatic   bool
	IsAbstract bool
	IsOverride bool // New: Track if method is an override from parent/interface
}

type JavaClass struct {
	ID         string
	Name       string
	RawName    string
	Type       ClassType
	Extends    string
	Implements []string
	Fields     []Field
	Methods    []Method
	LogEntries []string
}
