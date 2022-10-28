package sometest

import (
	"testing"
	"time"

	"github.com/kkevinchou/izzet/animation"
)

func TestBlah(t *testing.T) {
	a := animation.Load("../assets/animations/izzet")
	a.Update(time.Second * 1)
	t.Fail()
}
