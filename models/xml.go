package models

import "encoding/xml"

// MxFile represents the root structure of a draw.io XML file.
// MxFile đại diện cho cấu trúc gốc của tệp XML draw.io.
type MxFile struct {
	XMLName xml.Name `xml:"mxfile"`
	Diagram Diagram  `xml:"diagram"`
}

// Diagram represents the diagram node within the XML.
// Diagram đại diện cho nút biểu đồ trong XML.
type Diagram struct {
	MxGraphModel MxGraphModel `xml:"mxGraphModel"`
}

// MxGraphModel represents the graph model containing the root element.
// MxGraphModel đại diện cho mô hình đồ thị chứa phần tử gốc.
type MxGraphModel struct {
	Root Root `xml:"root"`
}

// Root contains the list of cells (graph elements).
// Root chứa danh sách các ô (các phần tử đồ thị).
type Root struct {
	MxCells []MxCell `xml:"mxCell"`
}

// MxCell represents a single element in the diagram (vertex or edge).
// MxCell đại diện cho một phần tử đơn lẻ trong biểu đồ (đỉnh hoặc cạnh).
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

// MxGeometry represents the geometric properties of a cell.
// MxGeometry đại diện cho các thuộc tính hình học của một ô.
type MxGeometry struct {
	X      string `xml:"x,attr"`
	Y      string `xml:"y,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}
