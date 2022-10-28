package items

import "math/rand"

type Rarity string
type ItemType string

var ItemTypeCoin ItemType = "COIN"

type Item struct {
	ID   int
	Type ItemType
}

const (
	RarityNormal Rarity = "NORMAL"
	RarityMagic  Rarity = "MAGIC"
	RarityRare   Rarity = "RARE"
)

func maxCountsByRarity(rarity Rarity) (int, int) {
	if rarity == RarityNormal {
		return 0, 0
	} else if rarity == RarityMagic {
		return 1, 1
	} else if rarity == RarityRare {
		return 3, 3
	}
	panic("MaxCountsByRarity unexpected rarity type")
}

func RarityToModCount(rarity Rarity) int {
	if rarity == RarityRare {
		return 3 + rand.Intn(4)
	} else if rarity == RarityMagic {
		return 1 + rand.Intn(2)
	}
	return 0
}

func SelectRarity(rarities []Rarity, weights []int) Rarity {
	var sum int
	for _, weight := range weights {
		sum += weight
	}

	roll := rand.Intn(sum)

	for i, weight := range weights {
		if roll < weight {
			return rarities[i]
		}
	}

	panic("unexpectedly didn't find a rarity")
}
