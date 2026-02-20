package models

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

// ParseXML reads and parses the draw.io XML file.
// ParseXML đọc và phân tích tệp XML draw.io.
func ParseXML(inputFile string) ([]MxCell, error) {
	byteValue, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file (lỗi đọc tệp): %v", err)
	}

	var mxFile MxFile
	if err := xml.Unmarshal(byteValue, &mxFile); err != nil {
		return nil, fmt.Errorf("error parsing XML (lỗi phân tích XML): %v", err)
	}

	return mxFile.Diagram.MxGraphModel.Root.MxCells, nil
}
