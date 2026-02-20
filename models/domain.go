package models

// ClassType defines the type of the class (Class, Interface, Enum, etc.).
// ClassType định nghĩa loại của lớp (Lớp, Giao diện, Enum, v.v.).
type ClassType string

const (
	// Class represents a standard Java class.
	// Class đại diện cho một lớp Java tiêu chuẩn.
	Class ClassType = "class"

	// Interface represents a Java interface.
	// Interface đại diện cho một giao diện Java.
	Interface ClassType = "interface"

	// Enum represents a Java enumeration.
	// Enum đại diện cho một kiểu liệt kê Java.
	Enum ClassType = "enum"

	// Record represents a Java record.
	// Record đại diện cho một bản ghi Java.
	Record ClassType = "record"

	// Abstract represents an abstract Java class.
	// Abstract đại diện cho một lớp Java trừu tượng.
	Abstract ClassType = "abstract"
)

// Field represents a field (attribute) in a class.
// Field đại diện cho một trường (thuộc tính) trong một lớp.
type Field struct {
	Original     string // Original string from diagram // Chuỗi gốc từ biểu đồ
	Name         string // Name of the field // Tên của trường
	Type         string // Data type of the field // Kiểu dữ liệu của trường
	Visibility   string // Access modifier (public, private, etc.) // Phạm vi truy cập (public, private, v.v.)
	IsStatic     bool   // Is the field static? // Trường có phải là tĩnh không?
	IsFinal      bool   // Is the field final? // Trường có phải là hằng số không?
	InitialValue string // Initial value of the field // Giá trị khởi tạo của trường
}

// Method represents a method (function) in a class.
// Method đại diện cho một phương thức (hàm) trong một lớp.
type Method struct {
	Original   string // Original string from diagram // Chuỗi gốc từ biểu đồ
	Name       string // Name of the method // Tên của phương thức
	Parameters string // Parameters of the method // Các tham số của phương thức
	ReturnType string // Return type of the method // Kiểu trả về của phương thức
	Visibility string // Access modifier // Phạm vi truy cập
	IsStatic   bool   // Is the method static? // Phương thức có phải là tĩnh không?
	IsAbstract bool   // Is the method abstract? // Phương thức có phải là trừu tượng không?
	IsOverride bool   // Is the method overriding a parent method? // Phương thức có ghi đè phương thức cha không?
}

// ClassModel represents the semantic model of a class/interface parsed from the diagram.
// ClassModel đại diện cho mô hình ngữ nghĩa của một lớp/giao diện được phân tích từ biểu đồ.
// Previously known as JavaClass.
// Trước đây được gọi là JavaClass.
type ClassModel struct {
	ID         string    // Unique ID from the diagram // ID duy nhất từ biểu đồ
	Name       string    // Cleaned name of the class // Tên đã làm sạch của lớp
	RawName    string    // Raw name from the diagram (for reference) // Tên gốc từ biểu đồ (để tham khảo)
	Type       ClassType // The type of the construct (Class, Interface, etc.) // Loại cấu trúc (Lớp, Giao diện, v.v.)
	Extends    string    // Name of the parent class // Tên của lớp cha
	Implements []string  // List of implemented interfaces // Danh sách các giao diện được triển khai
	Fields     []Field   // List of fields // Danh sách các trường
	Methods    []Method  // List of methods // Danh sách các phương thức
	LogEntries []string  // Log entries specific to this class // Các mục nhật ký cụ thể cho lớp này
}
