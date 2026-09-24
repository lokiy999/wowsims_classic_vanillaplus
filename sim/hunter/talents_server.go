package hunter

import (
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/hunter.json).
func (hunter *Hunter) applyServerTalents() {
	// Thick Hide: pet armor +15% per rank.
	if hunter.pet != nil && hunter.Talents.ThickHide > 0 {
		hunter.pet.MultiplyStat(stats.Armor, 1+0.15*float64(hunter.Talents.ThickHide))
	}
}
