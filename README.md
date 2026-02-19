# nUML: The Java Class Diagram Generator

**nUML** is a command-line tool written in Go that parses `draw.io` XML files and generates corresponding Java class files. It acts as a bridge between visual design and implementation by converting your class diagrams into boilerplate Java code.

**nUML** là công cụ dòng lệnh được viết bằng Go, phân tích cú pháp tệp XML `draw.io` và tạo các tệp Java tương ứng. Nó hoạt động như một cầu nối giữa thiết kế trực quan và triển khai bằng cách chuyển đổi sơ đồ lớp của bạn thành mã Java.

## Author
**Thai Thanh Nguyen**

## Features
- **XML Parsing**: Reads `draw.io` XML exports to understand your class structure.
- **Java Generation**: Automatically creates `.java` files for classes defined in the diagram.
- **Package Management**: Supports generating files into specific packages/folders.
- **Reporting**: Generates a `Report.md` summarizing the classes created.
- **Customizable**: Options to overwrite files, suppress reports, and control verbosity.
## Tính năng
- **Dịch XML**: Đọc file XML từ draw.io và chuyển đổi sang Java class.
- **Tạo package**: Tạo package và folder tương ứng với package.
- **Tạo báo cáo**: Tạo báo cáo tóm tắt các class đã tạo.
- **Tùy chỉnh**: Tùy chỉnh ghi đè file, tắt báo cáo và kiểm soát verbosity.

 
## Prerequisites
- **Go**: Version 1.25.3 or higher.

## Quick Use
1. **Download nUML.exe** from the [Releases](https://github.com/Nguyen-Agn/nUML/releases) page.
2. **Run** the following command:

```bash
./nUML [options] <file.drawio>
```

## Example
```bash
./nUML -f folderName -o my_diagram.drawio
```

## Installation & Usage Full

1.  **Clone the repository** (or download the source code).
2.  **Navigate directly** to the project folder.

### Running with Go
You can run the tool directly using `go run`:

```bash
go run . [options] <file.drawio>
```

### Running with Batch Script (Windows)
A convenience script `nUML.bat` is provided:

```cmd
nUML.bat [options] <file.drawio>
```

## Command Line Options

| Option | Description |
| :--- | :--- |
| `-f <folder>` | Generate files in the specified folder and add package declaration. |
| `-o` | Overwrite existing files (default: `false`). |
| `-v` | Verbose mode (print detailed progress). |
| `-l` | Skip generation of `Report.md`. |
| `-h` | Show help message. |

| Lựa chọn | Mô tả |
| :--- | :--- |
| `-f <folder>` | Tạo file trong thư mục được chỉ định và thêm khai báo package. |
| `-o` | Ghi đè file hiện có (mặc định: `false`). |
| `-v` | Chế độ verbose (in tiến trình chi tiết). |
| `-l` | Bỏ qua việc tạo `Report.md`. |
| `-h` | Hiển thị thông báo trợ giúp. |

## Examples

**Basic usage:**
```bash
go run . my_diagram.drawio
```

**Generate into a package "com.example.models" (folder "models") and overwrite existing files:**
```bash
go run . -f models -o my_diagram.drawio
```

# Build
```bash
go build -o nUML.exe main.go
```


## Note
- The folderName will be the package name and the folder name.
- The file name will be the class name.
- The file will be created in the same directory as the .exe file.

- folderName sẽ là tên package và tên folder.
- file name sẽ là tên class.
- file sẽ được tạo trong cùng thư mục với file .exe.
# nUML
