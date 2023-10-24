package serversystems

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/systems"
)

type ReplicationSystem struct {
}

func (s *ReplicationSystem) Update(time.Duration, systems.GameWorld) {
	fmt.Println("HELLO")
}
