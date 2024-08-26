package noise

import (
	"fmt"
	"testing"
)

func TestNoise(t *testing.T) {
	w := Worley3D(10, 10, 10)
	fmt.Println(w)
}
