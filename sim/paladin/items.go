package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Libram IDs
const (
	SanctifiedOrb    = 20512
	LibramOfHope     = 22401
	LibramOfFervor   = 23203
	LibramOfDivinity = 23201
)

func init() {
	core.NewSimpleStatOffensiveTrinketEffect(SanctifiedOrb, stats.Stats{stats.MeleeCrit: 5 * core.CritRatingPerCritChance, stats.SpellCrit: 5 * core.CritRatingPerCritChance}, time.Minute*1, time.Minute*3) // server: 5% for 1 min
}
