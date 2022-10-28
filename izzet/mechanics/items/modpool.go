package items

import (
	"fmt"
	"math/rand"
)

type AffixType string

const (
	AffixTypePrefix AffixType = "PREFIX"
	AffixTypeSuffix AffixType = "SUFFIX"
)

type ModPool struct {
	prefixPool map[int]*Mod
	suffixPool map[int]*Mod

	prefixList []*Mod
	suffixList []*Mod
}

func NewModPool() *ModPool {
	return &ModPool{
		prefixPool: map[int]*Mod{},
		suffixPool: map[int]*Mod{},
		prefixList: []*Mod{},
		suffixList: []*Mod{},
	}
}

func (m *ModPool) AddMod(mod *Mod) {
	if mod.AffixType == AffixTypePrefix {
		m.prefixPool[mod.ID] = mod
		m.prefixList = append(m.prefixList, mod)
	} else if mod.AffixType == AffixTypeSuffix {
		m.suffixPool[mod.ID] = mod
		m.suffixList = append(m.suffixList, mod)
	}
}

func (m *ModPool) ChooseMods(rarity Rarity) []*Mod {
	maxPrefix, maxSuffix := maxCountsByRarity(rarity)

	prefixCount := 1 + rand.Intn(maxPrefix)
	suffixCount := 1 + rand.Intn(maxSuffix)

	mods := []*Mod{}
	guard := 0
	maxGuard := 100

	seen := map[int]any{}
	for i := 0; i < prefixCount; i++ {
		for guard < maxGuard {
			guard = +1
			idx := rand.Intn(len(m.prefixList))
			if _, ok := seen[idx]; ok {
				continue
			}

			prefix := m.prefixList[idx]
			mods = append(mods, prefix)
			seen[prefix.ID] = true
			break
		}
	}

	for i := 0; i < suffixCount; i++ {
		for guard < maxGuard {
			guard = +1
			idx := rand.Intn(len(m.suffixList))
			if _, ok := seen[idx]; ok {
				continue
			}

			suffix := m.suffixList[idx]
			mods = append(mods, suffix)
			seen[suffix.ID] = true
			break
		}
	}

	if guard >= maxGuard {
		fmt.Println("WARNING, max guard hit in ChooseMods")
	}

	return mods
}

type Mod struct {
	ID        int
	AffixType AffixType
	Effect    Effect
}

func (m *Mod) String() string {
	return fmt.Sprintf("Mod{%d, %s}", m.ID, m.AffixType)
}

type Effect interface {
	ApplyDamage()
}
