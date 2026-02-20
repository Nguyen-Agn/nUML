package main

import (
	"fmt"
	"io/ioutil"
	"nUML/analyzer"
	"nUML/generator"
	"nUML/models"
	"nUML/utils"
	"os"
	"strings"
)

var targetPackage string
var OverwriteMode bool
var NoReportMode bool

func printHelp() {
	fmt.Println("nUML: The Java Class Diagram Generator")
	fmt.Println("Author: Thai Thanh Nguyen")
	fmt.Println("Usage: nUML [options] <file.drawio>")
	fmt.Println("Options:")
	fmt.Println("  -f <folder>   Generate files in the specified folder and add package declaration (Tạo tệp trong thư mục và thêm khai báo gói).")
	fmt.Println("  -o            Overwrite existing files (default: false) (Ghi đè tệp hiện có (mặc định: sai)).")
	fmt.Println("  -v            Verbose mode (print detailed progress) (Chế độ chi tiết (in tiến trình chi tiết)).")
	fmt.Println("  -l            Skip generation of Report.md (Bỏ qua việc tạo Report.md).")
	fmt.Println("  -h            Show this help message (Hiển thị tin nhắn trợ giúp này).")
}

func main() {
	utils.SetupLogging() // Initialize log buffer

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	var inputFile string
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" {
			printHelp()
			return
		} else if arg == "-f" {
			if i+1 < len(args) {
				targetPackage = args[i+1]
				i++
			} else {
				fmt.Println("Error: -f requires a folder name (Lỗi: -f yêu cầu tên thư mục)")
				return
			}
		} else if arg == "-o" {
			OverwriteMode = true
		} else if arg == "-v" {
			utils.VerboseMode = true
		} else if arg == "-l" {
			NoReportMode = true
		} else {
			inputFile = arg
		}
	}

	if inputFile == "" {
		fmt.Println("Error: No input file specified (Lỗi: Không có tệp đầu vào nào được chỉ định)")
		return
	}

	utils.LogInfo(fmt.Sprintf("Processing file (Đang xử lý tệp): %s", inputFile))
	if targetPackage != "" {
		utils.LogVerbose(fmt.Sprintf("Target Package/Folder (Gói/Thư mục đích): %s", targetPackage))
		if _, err := os.Stat(targetPackage); os.IsNotExist(err) {
			os.Mkdir(targetPackage, 0755)
		}
	}

	// 1. Parsing
	// 1. Phân tích cú pháp
	cells, err := models.ParseXML(inputFile)
	if err != nil {
		utils.LogInfo(fmt.Sprintf("Error (Lỗi): %v", err))
		return
	}

	// 2. Analysis
	// 2. Phân tích
	ana := analyzer.NewAnalyzerService()
	classes := ana.AnalyzeDiagram(cells)

	// 3. Generation
	// 3. Tạo code
	gen := generator.NewJavaGenerator(targetPackage)

	var overallReport strings.Builder
	overallReport.WriteString("# Generation Report (Báo cáo tạo code)\n\n")

	for _, cls := range classes {
		artifact, err := gen.Generate(cls)
		if err != nil {
			utils.LogInfo(fmt.Sprintf("Failed to generate code for (Không thể tạo code cho) %s: %v", cls.Name, err))
			continue
		}

		// Handle File Writing
		// Xử lý ghi tệp
		finalPath := artifact.FileName
		// Check overwrite
		// Kiểm tra ghi đè
		skip := false
		if !OverwriteMode {
			if _, err := os.Stat(finalPath); err == nil {
				utils.LogVerbose(fmt.Sprintf("Skipped (Đã bỏ qua) %s (exists, use -o to overwrite)", finalPath))
				overallReport.WriteString(fmt.Sprintf("# %s [Skipped (Đã bỏ qua)]\n- File exists and -o not set.\n\n", cls.Name))
				skip = true
			}
		}

		if !skip {
			f, err := os.Create(finalPath)
			if err != nil {
				utils.LogInfo(fmt.Sprintf("Failed to create file (Không thể tạo tệp) %s: %v", finalPath, err))
			} else {
				f.WriteString(artifact.Content)
				f.Close()
				utils.LogVerbose(fmt.Sprintf("Generated (Đã tạo) %s", finalPath))
				overallReport.WriteString(artifact.ReportEntry)
			}
		}
	}

	// Save Report
	// Lưu báo cáo
	if !NoReportMode {
		ioutil.WriteFile("Report.md", []byte(overallReport.String()), 0644)
		utils.LogInfo("Generated Report.md")
	}

	// 4. Finalize
	// 4. Hoàn tất
	// utils.WriteLog() // Removed as per user request
}
