package common

import "testing"

func TestIsBlank(t *testing.T) {
	if !IsBlank("") || !IsBlank("  \t") || IsBlank("a") {
		t.Fatal("IsBlank")
	}
}

func TestCoalesce(t *testing.T) {
	if Coalesce("", "x") != "x" || Coalesce("a", "x") != "a" {
		t.Fatal("Coalesce")
	}
}

func TestCoalesceTrim(t *testing.T) {
	if CoalesceTrim("  ", "x") != "x" || CoalesceTrim(" a ", "x") != "a" {
		t.Fatal("CoalesceTrim")
	}
}

func TestFirstNonBlank(t *testing.T) {
	if FirstNonBlank("", "  ", "b") != "b" || FirstNonBlank("", " ") != "" {
		t.Fatal("FirstNonBlank")
	}
}

func TestRequireNonBlank(t *testing.T) {
	if RequireNonBlank("ok", "x") != nil {
		t.Fatal("should pass")
	}
	err := RequireNonBlank(" ", "缺少字段")
	if err == nil || err.Error() != ErrParams.WithMessage("缺少字段").Error() {
		t.Fatalf("got %v", err)
	}
}

func TestIsNilOrEmpty(t *testing.T) {
	var nilSlice []int
	if !IsNilOrEmpty(nilSlice) || !IsNilOrEmpty([]string{}) || IsNilOrEmpty([]int{1}) {
		t.Fatal("IsNilOrEmpty")
	}
}
