package web

import (
	"errors"
	"testing"
)

// TestErr_Error 测试 Error 方法
func TestErr_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Err
		want string
	}{
		{
			name: "nil error",
			err:  &Err{},
			want: "",
		},
		{
			name: "with error",
			err:  &Err{err: errors.New("test error")},
			want: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Err.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestErr_Wrap 测试 Wrap 方法
func TestErr_Wrap(t *testing.T) {
	originalErr := errors.New("original error")

	t.Run("wrap normal error", func(t *testing.T) {
		err := NewErr(500, 500, ErrInternal).Wrap(originalErr)
		if err.err != originalErr {
			t.Errorf("expected err.err to be originalErr")
		}
		if err.Error() != "original error" {
			t.Errorf("Error() = %v, want %v", err.Error(), "original error")
		}
	})

	t.Run("wrap Err should not wrap, return as-is", func(t *testing.T) {
		err1 := NewErr(500, 500, ErrInternal).Wrap(originalErr)
		err2 := NewErr(400, 400, ErrBindParams).Wrap(err1)

		// 传入 *Err 时,不再包装,err2.err 应该为 nil (初始状态)
		if err2.err != nil {
			t.Errorf("expected err2.err to be nil (not wrapped), got %v", err2.err)
		}
		if err2.Error() != "" {
			t.Errorf("Error() = %v, want empty string", err2.Error())
		}
	})

	t.Run("wrap nil error", func(t *testing.T) {
		err := NewErr(500, 500, ErrInternal).Wrap(nil)

		// 直接包装 nil
		if err.err != nil {
			t.Errorf("expected err.err to be nil, got %v", err.err)
		}
		if err.Error() != "" {
			t.Errorf("Error() = %v, want empty string", err.Error())
		}
	})
}

// TestErr_WithData 测试 WithData 方法
func TestErr_WithData(t *testing.T) {
	err := NewErr(400, 400, ErrBindParams)
	data := map[string]string{"field": "email"}

	result := err.WithData(data)

	if result.data == nil {
		t.Errorf("expected data to be set")
	}
	if result != err {
		t.Errorf("expected WithData to return same Err instance")
	}
}

// TestErr_WithParam 测试 WithParam 方法
func TestErr_WithParam(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		err := NewErr(400, 400, ErrBindParams)
		result := err.WithParam("key1", "value1", "key2", 123)

		if len(result.param) != 2 {
			t.Errorf("expected 2 params, got %d", len(result.param))
		}
		if result.param["key1"] != "value1" {
			t.Errorf("expected param key1 = value1, got %v", result.param["key1"])
		}
		if result.param["key2"] != 123 {
			t.Errorf("expected param key2 = 123, got %v", result.param["key2"])
		}
		if result != err {
			t.Errorf("expected WithParam to return same Err instance")
		}
	})

	t.Run("odd number of params should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for odd number of params")
			} else {
				// 验证 panic 消息是否正确
				if msg, ok := r.(string); ok && msg != "[WEB] WithParam kv must be even" {
					t.Errorf("expected panic message '[WEB] WithParam kv must be even', got %v", msg)
				}
			}
		}()

		err := NewErr(400, 400, ErrBindParams)
		err.WithParam("key1", "value1", "key2")
	})

	t.Run("non-string key should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for non-string key")
			} else {
				// 验证 panic 消息是否正确
				if msg, ok := r.(string); ok && msg != "[WEB] WithParam key must be string" {
					t.Errorf("expected panic message '[WEB] WithParam key must be string', got %v", msg)
				}
			}
		}()

		err := NewErr(400, 400, ErrBindParams)
		err.WithParam(123, "value1") // key 是 int 而非 string
	})
}

// TestNewErr 测试 NewErr 构造函数
func TestNewErr(t *testing.T) {
	err := NewErr(500, 1001, ErrInternal)

	if err.Status != 500 {
		t.Errorf("expected Status = 500, got %d", err.Status)
	}
	if err.code != 1001 {
		t.Errorf("expected code = 1001, got %d", err.code)
	}
	if err.id != ErrInternal {
		t.Errorf("expected id = ErrInternal, got %v", err.id)
	}
}

// TestNewBadRequestErr 测试 NewBadRequestErr 构造函数
func TestNewBadRequestErr(t *testing.T) {
	err := NewBadRequestErr(ErrBindParams)

	if err.Status != 400 {
		t.Errorf("expected Status = 400, got %d", err.Status)
	}
	if err.code != 400 {
		t.Errorf("expected code = 400, got %d", err.code)
	}
	if err.id != ErrBindParams {
		t.Errorf("expected id = ErrBindParams, got %v", err.id)
	}
}

// TestErr_ChainedCalls 测试链式调用
func TestErr_ChainedCalls(t *testing.T) {
	originalErr := errors.New("original error")
	data := map[string]string{"field": "email"}

	err := NewBadRequestErr(ErrBindParams).
		Wrap(originalErr).
		WithData(data).
		WithParam("key1", "value1", "key2", 123)

	if err.Error() != "original error" {
		t.Errorf("Error() = %v, want %v", err.Error(), "original error")
	}
	if err.data == nil {
		t.Errorf("expected data to be set")
	}
	if len(err.param) != 2 {
		t.Errorf("expected 2 params, got %d", len(err.param))
	}
	if err.Status != 400 {
		t.Errorf("expected Status = 400, got %d", err.Status)
	}
}
