package serializer

import (
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestMsgpackSerializer_Name(t *testing.T) {
	s := NewMsgpackSerializer()
	if s.Name() != "msgpack" {
		t.Errorf("Expected name 'msgpack', got '%s'", s.Name())
	}
}

func TestMsgpackSerializer_SimpleTypes(t *testing.T) {
	s := NewMsgpackSerializer()

	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello world"},
		{"int", 42},
		{"int64", int64(9223372036854775807)},
		{"float64", 3.14159},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := s.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			var result interface{}
			if err := s.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// For nil, result should be nil
			if tt.value == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			// Msgpack preserves types better than JSON
			// Just verify no errors for now
		})
	}
}

func TestMsgpackSerializer_Struct(t *testing.T) {
	s := NewMsgpackSerializer()

	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Marshal
	data, err := s.Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal into envelope
	var envelope Envelope
	var resultUser User
	envelope.Value = &resultUser

	if err := msgpack.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check type information
	if envelope.Type != "serializer.User" {
		t.Errorf("Expected type 'serializer.User', got '%s'", envelope.Type)
	}

	// Check values
	if resultUser.ID != user.ID || resultUser.Name != user.Name || resultUser.Email != user.Email {
		t.Errorf("User mismatch: expected %+v, got %+v", user, resultUser)
	}
}

func TestMsgpackSerializer_Slice(t *testing.T) {
	s := NewMsgpackSerializer()

	items := []int{1, 2, 3, 4, 5}

	// Marshal
	data, err := s.Marshal(items)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var envelope Envelope
	var result []int
	envelope.Value = &result

	if err := msgpack.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check type
	if envelope.Type != "[]int" {
		t.Errorf("Expected type '[]int', got '%s'", envelope.Type)
	}

	// Check values
	if len(result) != len(items) {
		t.Errorf("Length mismatch: expected %d, got %d", len(items), len(result))
	}
	for i, v := range result {
		if v != items[i] {
			t.Errorf("Value mismatch at index %d: expected %d, got %d", i, items[i], v)
		}
	}
}

func TestMsgpackSerializer_Map(t *testing.T) {
	s := NewMsgpackSerializer()

	data := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"admin": true,
	}

	// Marshal
	bytes, err := s.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var envelope Envelope
	var result map[string]interface{}
	envelope.Value = &result

	if err := msgpack.Unmarshal(bytes, &envelope); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check type
	if envelope.Type != "map[string]interface {}" {
		t.Errorf("Expected type 'map[string]interface {}', got '%s'", envelope.Type)
	}

	// Check values
	if result["name"] != "John" {
		t.Errorf("Expected name 'John', got '%v'", result["name"])
	}
}

// Benchmark comparison
func BenchmarkJSON_Marshal(b *testing.B) {
	s := NewJSONSerializer()
	type User struct {
		ID    int
		Name  string
		Email string
	}
	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Marshal(user)
	}
}

func BenchmarkMsgpack_Marshal(b *testing.B) {
	s := NewMsgpackSerializer()
	type User struct {
		ID    int
		Name  string
		Email string
	}
	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Marshal(user)
	}
}

func BenchmarkJSON_Unmarshal(b *testing.B) {
	s := NewJSONSerializer()
	type User struct {
		ID    int
		Name  string
		Email string
	}
	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	data, _ := s.Marshal(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result User
		_ = s.Unmarshal(data, &result)
	}
}

func BenchmarkMsgpack_Unmarshal(b *testing.B) {
	s := NewMsgpackSerializer()
	type User struct {
		ID    int
		Name  string
		Email string
	}
	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	data, _ := s.Marshal(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result User
		_ = s.Unmarshal(data, &result)
	}
}
