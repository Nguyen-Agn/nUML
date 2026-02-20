package analyzer

import (
	"nUML/models"
)

// AnalyzerService orchestrates the analysis process.
// AnalyzerService điều phối quá trình phân tích.
type AnalyzerService struct {
	classExtractor        *ClassExtractor
	featureExtractor      *FeatureExtractor
	relationshipExtractor *RelationshipExtractor
	hierarchyResolver     *HierarchyResolver
}

// NewAnalyzerService creates a new instance of AnalyzerService.
// NewAnalyzerService tạo một phiên bản mới của AnalyzerService.
func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		classExtractor:        NewClassExtractor(),
		featureExtractor:      NewFeatureExtractor(),
		relationshipExtractor: NewRelationshipExtractor(),
		hierarchyResolver:     NewHierarchyResolver(),
	}
}

// AnalyzeDiagram processes the raw cells to produce a semantic model of the classes.
// AnalyzeDiagram xử lý các ô thô để tạo ra mô hình ngữ nghĩa của các lớp.
func (as *AnalyzerService) AnalyzeDiagram(cells []models.MxCell) map[string]*models.ClassModel {
	// 1. Identify Classes
	// 1. Xác định các lớp
	classes := as.classExtractor.Extract(cells)

	// 2. Identify Features (Fields, Methods in Swimlanes)
	// 2. Xác định các đặc điểm (Trường, Phương thức trong Swimlanes)
	as.featureExtractor.Extract(cells, classes)

	// 3. Identify Relationships
	// 3. Xác định các mối quan hệ
	as.relationshipExtractor.Extract(cells, classes)

	// 4. Resolve Inheritance
	// 4. Giải quyết kế thừa
	as.hierarchyResolver.Resolve(classes)

	return classes
}
