package validate

import (
	"testing"
	"time"
)

type pointerField struct {
	Count *int
}

func TestCustomValidator_ExampleUsage(t *testing.T) {
	TestValidator()
}

func TestCustomValidator_Validate(t *testing.T) {
	validator := NewCustomValidator()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "空结构体应该应用默认值",
			input:   &UserRequest{},
			wantErr: false,
		},
		{
			name: "部分字段有值应该保留原值",
			input: &UserRequest{
				Name: "测试用户",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name: "验证失败的情况",
			input: &UserRequest{
				Name: "测试用户",
				Age:  200, // 超出范围
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CustomValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomValidator_SetDefault(t *testing.T) {
	validator := NewCustomValidator()

	tests := []struct {
		name         string
		input        any
		fieldName    string
		defaultValue string
		wantErr      bool
		check        func(*testing.T, any)
	}{
		{
			name:         "设置字符串默认值",
			input:        &UserRequest{},
			fieldName:    "Name",
			defaultValue: "默认名称",
			wantErr:      false,
		},
		{
			name:         "设置整数默认值",
			input:        &UserRequest{},
			fieldName:    "Age",
			defaultValue: "25",
			wantErr:      false,
		},
		{
			name:         "设置布尔默认值",
			input:        &UserRequest{},
			fieldName:    "IsActive",
			defaultValue: "true",
			wantErr:      false,
		},
		{
			name:         "设置指针默认值",
			input:        &pointerField{},
			fieldName:    "Count",
			defaultValue: "12",
			wantErr:      false,
			check: func(t *testing.T, input any) {
				pf := input.(*pointerField)
				if pf.Count == nil || *pf.Count != 12 {
					t.Fatalf("Count default value = %v, want %v", pf.Count, 12)
				}
			},
		},
		{
			name:         "字段不存在",
			input:        &UserRequest{},
			fieldName:    "NonExistentField",
			defaultValue: "value",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.SetDefault(tt.input, tt.fieldName, tt.defaultValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("CustomValidator.SetDefault() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.check != nil {
				tt.check(t, tt.input)
			}
		})
	}
}

func TestCustomValidator_applyDefaults(t *testing.T) {
	validator := NewCustomValidator()

	// 测试空结构体
	userReq := &UserRequest{}
	err := validator.applyDefaults(userReq)
	if err != nil {
		t.Errorf("applyDefaults() error = %v", err)
	}

	// 验证默认值是否正确应用
	if userReq.Name != "匿名用户" {
		t.Errorf("Name default value = %v, want %v", userReq.Name, "匿名用户")
	}
	if userReq.Age != 18 {
		t.Errorf("Age default value = %v, want %v", userReq.Age, 18)
	}
	if userReq.Email != "user@example.com" {
		t.Errorf("Email default value = %v, want %v", userReq.Email, "user@example.com")
	}
	if !userReq.IsActive {
		t.Errorf("IsActive default value = %v, want %v", userReq.IsActive, true)
	}
	if userReq.Page != 1 {
		t.Errorf("Page default value = %v, want %v", userReq.Page, 1)
	}
	if userReq.Size != 10 {
		t.Errorf("Size default value = %v, want %v", userReq.Size, 10)
	}
}

func TestCustomValidator_NestedStruct(t *testing.T) {
	validator := NewCustomValidator()

	// 测试嵌套结构体
	productReq := &ProductRequest{
		Name: "测试产品",
	}

	err := validator.applyDefaults(productReq)
	if err != nil {
		t.Errorf("applyDefaults() error = %v", err)
	}

	// 验证主结构体默认值
	if productReq.Price != 0.0 {
		t.Errorf("Price default value = %v, want %v", productReq.Price, 0.0)
	}
	if productReq.Description != "暂无描述" {
		t.Errorf("Description default value = %v, want %v", productReq.Description, "暂无描述")
	}
	if productReq.Category != "未分类" {
		t.Errorf("Category default value = %v, want %v", productReq.Category, "未分类")
	}
	if productReq.InStock {
		t.Errorf("InStock default value = %v, want %v", productReq.InStock, false)
	}

	// 验证嵌套结构体默认值
	if productReq.Address.Street != "未知街道" {
		t.Errorf("Address.Street default value = %v, want %v", productReq.Address.Street, "未知街道")
	}
	if productReq.Address.City != "未知城市" {
		t.Errorf("Address.City default value = %v, want %v", productReq.Address.City, "未知城市")
	}
	if productReq.Address.Province != "未知省份" {
		t.Errorf("Address.Province default value = %v, want %v", productReq.Address.Province, "未知省份")
	}
	if productReq.Address.ZipCode != "000000" {
		t.Errorf("Address.ZipCode default value = %v, want %v", productReq.Address.ZipCode, "000000")
	}

	// 验证嵌套结构体指针默认值
	if productReq.Contact == nil {
		t.Error("Contact should not be nil")
	} else {
		if productReq.Contact.Phone != "无" {
			t.Errorf("Contact.Phone default value = %v, want %v", productReq.Contact.Phone, "无")
		}
		if productReq.Contact.Email != "contact@example.com" {
			t.Errorf("Contact.Email default value = %v, want %v", productReq.Contact.Email, "contact@example.com")
		}
		if productReq.Contact.Website != "https://example.com" {
			t.Errorf("Contact.Website default value = %v, want %v", productReq.Contact.Website, "https://example.com")
		}
	}
}

func TestCustomValidator_PointerAndCompositeDefaults(t *testing.T) {
	validator := NewCustomValidator()

	type complexDefaults struct {
		Enabled *bool             `default:"true"`
		Limit   *int              `default:"99"`
		Tags    []string          `default:"[\"base\",\"beta\"]"`
		Meta    map[string]string `default:"{\"env\":\"dev\"}"`
		Payload any               `default:"{\"role\":\"user\"}"`
		Timeout time.Duration     `default:"5s"`
	}

	req := &complexDefaults{}
	if err := validator.applyDefaults(req); err != nil {
		t.Fatalf("applyDefaults() error = %v", err)
	}

	if req.Enabled == nil || !*req.Enabled {
		t.Fatalf("Enabled default value = %v, want %v", req.Enabled, true)
	}
	if req.Limit == nil || *req.Limit != 99 {
		t.Fatalf("Limit default value = %v, want %v", req.Limit, 99)
	}
	if len(req.Tags) != 2 {
		t.Fatalf("Tags default length = %d, want %d", len(req.Tags), 2)
	}
	if req.Meta == nil || req.Meta["env"] != "dev" {
		t.Fatalf("Meta default value = %v, want env=dev", req.Meta)
	}
	if req.Timeout != 5*time.Second {
		t.Fatalf("Timeout default value = %v, want %v", req.Timeout, 5*time.Second)
	}

	payload, ok := req.Payload.(map[string]any)
	if !ok {
		t.Fatalf("Payload type = %T, want map[string]any", req.Payload)
	}
	if payload["role"] != "user" {
		t.Fatalf("Payload role = %v, want %v", payload["role"], "user")
	}
}
