package tools

import (
	"fmt"
	"strings"
)

type EnumStringFlag[T ~string] struct {
    Value   T
    Allowed []T
}

type DBStringFlag[T ~string] struct {
    StringConnection T
}


func (e *EnumStringFlag[T]) String() string {
    return string(e.Value)
}

func (e *EnumStringFlag[T]) Set(s string) error {
    for _, v := range e.Allowed {
        if s == string(v) {
            e.Value = v
            return nil
        }
    }
    return fmt.Errorf("invalid value %q (allowed: %s)", s, e.joinAllowed())
}

func (e *EnumStringFlag[T]) joinAllowed() string {
    strs := make([]string, len(e.Allowed))
    for i, v := range e.Allowed {
        strs[i] = string(v)
    }
    return strings.Join(strs, ", ")
}

func NewEnumStringFlag[T ~string](defaultVal T, allowed []T, description string) (*EnumStringFlag[T], string) {
    enum := &EnumStringFlag[T]{Value: defaultVal, Allowed: allowed}
    desc := fmt.Sprintf("%s (allowed: %s)", description, enum.joinAllowed())
    return enum, desc
}