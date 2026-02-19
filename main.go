package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func printHelp() {
	fmt.Println("nUML: The Java Class Diagram Generator")
	fmt.Println("Author: Thai Thanh Nguyen")
	fmt.Println("Usage: nUML [options] <file.drawio>")
	fmt.Println("Options:")
	fmt.Println("  -f <folder>   Generate files in the specified folder and add package declaration.")
	fmt.Println("  -o            Overwrite existing files (default: false).")
	fmt.Println("  -v            Verbose mode (print detailed progress).")
	fmt.Println("  -l            Skip generation of Report.md.")
	fmt.Println("  -h            Show this help message.")
}

func main() {
	setupLogging() // Initialize log buffer

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
				fmt.Println("Error: -f requires a folder name")
				return
			}
		} else if arg == "-o" {
			OverwriteMode = true
		} else if arg == "-v" {
			VerboseMode = true
		} else if arg == "-l" {
			NoReportMode = true
		} else {
			inputFile = arg
		}
	}

	if inputFile == "" {
		fmt.Println("Error: No input file specified")
		return
	}

	logInfo(fmt.Sprintf("Processing file: %s", inputFile))
	if targetPackage != "" {
		// Only show output folder in verbose or if pertinent? User said "2 lines".
		// We'll show it in verbose.
		logVerbose(fmt.Sprintf("Target Package/Folder: %s", targetPackage))
		if _, err := os.Stat(targetPackage); os.IsNotExist(err) {
			os.Mkdir(targetPackage, 0755)
		}
	}

	// 1. Parsing
	cells, err := ParseXML(inputFile)
	if err != nil {
		logInfo(fmt.Sprintf("Error: %v", err))
		writeLog()
		return
	}

	// 2. Analysis
	classes := AnalyzeDiagram(cells)

	// 3. Generation
	var overallReport strings.Builder
	overallReport.WriteString("# Generation Report\n\n")

	for _, cls := range classes {
		rpt := generateJavaFile(cls)
		overallReport.WriteString(rpt)
	}

	// Save Report
	if !NoReportMode {
		ioutil.WriteFile("Report.md", []byte(overallReport.String()), 0644)
		logInfo("Generated Report.md")
	}

	// 4. Finalize
	writeLog()
}
