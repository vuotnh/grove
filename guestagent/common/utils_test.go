package guestagent

import (
	"os"
	"testing"
)

// Tạo một struct mock codec để test
type MockCodec struct{}

func (m MockCodec) Serialize(data interface{}) ([]byte, error) {
	return []byte(`{"key":"value"}`), nil
}

func (m MockCodec) Deserialize(data []byte, v interface{}) error {
	// Giả lập decode JSON
	*(v.(*interface{})) = map[string]interface{}{"key": "value"}
	return nil
}

func TestReadFile(t *testing.T) {
	// Tạo file tạm để test
	tmpFile, err := os.CreateTemp("", "testfile.json")
	if err != nil {
		t.Fatalf("Không thể tạo file tạm: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Xóa file sau khi test xong

	// Ghi nội dung giả vào file
	testContent := `{"key": "value"}`
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Không thể ghi vào file tạm: %v", err)
	}
	tmpFile.Close() // Đóng file để đảm bảo có thể đọc lại

	// Tạo codec giả lập
	mockCodec := MockCodec{}

	// Test đọc file và decode
	data, err := ReadFile(tmpFile.Name(), mockCodec, true)
	if err != nil {
		t.Fatalf("readFile thất bại: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok || result["key"] != "value" {
		t.Errorf("Dữ liệu đọc được không đúng, nhận: %v", data)
	}

	_, err = ReadFile("khong_ton_tai.json", mockCodec, true)
	if err == nil {
		t.Errorf("readFile không báo lỗi khi đọc file không tồn tại")
	}
}
