package utils_test

import (
	"testing"

	"github.com/kkevinchou/izzet/internal/utils"
)

func TestPriorityQueue(t *testing.T) {
	pq := utils.NewPriorityQueue()
	pq.Push("zuko", 3)
	pq.Push("aang", 0)
	pq.Push("katara", 1)
	pq.Push("sokka", 6)
	pq.Push("toph", 2)

	order := []string{
		"aang",
		"katara",
		"toph",
		"zuko",
		"sokka",
	}

	index := 0
	for pq.Len() > 0 {
		actual := pq.Pop()
		if actual != order[index] {
			t.Fatalf("Expected %s but got %s\n", order[index], actual)
		}
		index++
	}
}
