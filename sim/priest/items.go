package priest

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const (
	// Keep these ordered by ID
	MarkOfTheVeteranFirebird    = 26166
	MarkOfTheVeteranRattlesnake = 26175
)

func init() {
	core.AddEffectsToTest = false

	// Keep these ordered by name

	// Equip: Increases spell damage by up to 8% of your total Intellect and
	// healing done by up to 15% of your total Spirit.
	markOfTheVeteranEffect := func(agent core.Agent) {
		priest := agent.(PriestAgent).GetPriest()
		priest.AddStatDependency(stats.Intellect, stats.SpellPower, 0.08)
		priest.AddStatDependency(stats.Spirit, stats.HealingPower, 0.15)
	}
	core.NewItemEffect(MarkOfTheVeteranFirebird, markOfTheVeteranEffect)
	core.NewItemEffect(MarkOfTheVeteranRattlesnake, markOfTheVeteranEffect)

	core.AddEffectsToTest = true
}
