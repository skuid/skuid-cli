package testutil

import (
	"testing"
)

type EnvVar[T any] struct {
	EnvValue string
	Value    T
}

func SetupEnv[T any](t *testing.T, vars map[string]EnvVar[T]) {
	for key, val := range vars {
		t.Setenv(key, val.EnvValue)
	}
}

// testify evaluates pointer equality using referenced value, not pointer itself
// so this can be used to test pointer equality
func SameMatcher(a any) func(fn any) bool {
	return func(fn any) bool {
		return a == fn
	}
}
