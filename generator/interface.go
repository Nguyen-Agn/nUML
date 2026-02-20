package generator

import "nUML/models"

// GeneratedArtifact represents the output of a generation process.
// GeneratedArtifact đại diện cho đầu ra của quá trình tạo.
type GeneratedArtifact struct {
	FileName    string // The name of the file to be created // Tên của tệp sẽ được tạo
	Content     string // The content of the file // Nội dung của tệp
	ReportEntry string // A markdown entry for the generation report // Một mục markdown cho báo cáo tạo
}

// CodeGenerator defines the interface for generating code from ClassModels.
// CodeGenerator định nghĩa giao diện để tạo code từ ClassModel.
type CodeGenerator interface {
	// Generate produces an artifact for a given class model.
	// Generate tạo ra một sản phẩm (artifact) cho mô hình lớp đã cho.
	Generate(cls *models.ClassModel) (*GeneratedArtifact, error)
}
