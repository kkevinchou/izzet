package js

import (
	"fmt"
	"testing"

	"github.com/robertkrimen/otto"
)

func TestBlah2(t *testing.T) {
	vm := otto.New()
	vm.Call(`
		2 + 2;
	`, nil, nil)

	if value, err := vm.Get("abc"); err == nil {
		if value_int, err := value.ToInteger(); err == nil {
			fmt.Printf("", value_int, err)
		}
	}

	t.Fail()
}
