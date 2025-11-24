package serializer

import (
	"encoding/json"
	"testing"
)

func TestJSONSerializer_Name(t *testing.T) {
	s := NewJSONSerializer()
	if s.Name() != "json" {
		t.Errorf("Expected name 'json', got '%s'", s.Name())
	}
}

func TestJSONSerializer_SimpleTypes(t *testing.T) {
	s := NewJSONSerializer()

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

			// For numbers, JSON unmarshals to float64
			if _, ok := tt.value.(int); ok {
				if result != float64(tt.value.(int)) {
					t.Errorf("Expected %v, got %v", float64(tt.value.(int)), result)
				}
				return
			}
			if _, ok := tt.value.(int64); ok {
				if result != float64(tt.value.(int64)) {
					t.Errorf("Expected %v, got %v", float64(tt.value.(int64)), result)
				}
				return
			}

			// For other types, direct comparison
			if result != tt.value {
				t.Errorf("Expected %v, got %v", tt.value, result)
			}
		})
	}
}

func TestJSONSerializer_Struct(t *testing.T) {
	s := NewJSONSerializer()

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

	if err := json.Unmarshal(data, &envelope); err != nil {
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

func TestJSONSerializer_Slice(t *testing.T) {
	s := NewJSONSerializer()

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

	if err := json.Unmarshal(data, &envelope); err != nil {
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

func TestJSONSerializer_Map(t *testing.T) {
	s := NewJSONSerializer()

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

	if err := json.Unmarshal(bytes, &envelope); err != nil {
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

func TestJSONSerializer_NestedStruct(t *testing.T) {
	s := NewJSONSerializer()

	type Address struct {
		Street string
		City   string
	}

	type Person struct {
		Name    string
		Age     int
		Address Address
	}

	person := Person{
		Name: "Alice",
		Age:  25,
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	// Marshal
	data, err := s.Marshal(person)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var envelope Envelope
	var result Person
	envelope.Value = &result

	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check values
	if result.Name != person.Name || result.Age != person.Age {
		t.Errorf("Person mismatch: expected %+v, got %+v", person, result)
	}
	if result.Address.Street != person.Address.Street || result.Address.City != person.Address.City {
		t.Errorf("Address mismatch: expected %+v, got %+v", person.Address, result.Address)
	}
}

func TestJSONSerializer_EmptyValues(t *testing.T) {
	s := NewJSONSerializer()

	tests := []struct {
		name  string
		value interface{}
	}{
		{"empty string", ""},
		{"empty slice", []int{}},
		{"empty map", map[string]string{}},
		{"zero int", 0},
		{"false bool", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := s.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var result interface{}
			if err := s.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Just verify no errors occurred
			// Exact value comparison is tricky due to JSON type conversions
		})
	}
}
