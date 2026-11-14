package colorkit_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/b1ug/colorkit"
)

func requireColorError(t *testing.T, err, want error, field string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected color error")
	}
	if !errors.Is(err, want) {
		t.Fatalf("errors.Is(%v,%v)=false", err, want)
	}
	var detail *colorkit.Error
	if !errors.As(err, &detail) || detail == nil || detail.Op == "" {
		t.Fatalf("missing structured error: %#v", err)
	}
	if field != "" && detail.Field != field {
		t.Fatalf("error field %q, want %q", detail.Field, field)
	}
	if detail.Error() == "" {
		t.Fatal("empty diagnostic")
	}
}

func TestErrorsAreInspectableAfterWrapping(t *testing.T) {
	_, err := colorkit.Parse("not-a-color")
	wrapped := fmt.Errorf("caller context: %w", err)
	requireColorError(t, wrapped, colorkit.ErrSyntax, "")
	_, err = colorkit.FromHSVDecimal("0", "101", "100")
	requireColorError(t, fmt.Errorf("outer: %w", err), colorkit.ErrRange, "s")
	_, err = colorkit.FromColor(fixedColor{a: 1})
	requireColorError(t, fmt.Errorf("outer: %w", err), colorkit.ErrAlpha, "a")
}
