package rogue

import (
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/rogue.json).
func (rogue *Rogue) applyServerTalents() {
	// Bitter Experience: +1% Dodge and +0.2 all resistances per level, per rank.
	if rogue.Talents.BitterExperience > 0 {
		pts := float64(rogue.Talents.BitterExperience)
		res := 0.2 * float64(rogue.Level) * pts
		rogue.AddStats(stats.Stats{
			stats.Dodge:            pts,
			stats.ArcaneResistance: res,
			stats.FireResistance:   res,
			stats.FrostResistance:  res,
			stats.NatureResistance: res,
			stats.ShadowResistance: res,
		})
	}

	// Steadfast Determination: +5% Stamina per rank (the stun/fear resist part is not modeled).
	if rogue.Talents.SteadfastDetermination > 0 {
		rogue.MultiplyStat(stats.Stamina, 1+0.05*float64(rogue.Talents.SteadfastDetermination))
	}

	// Sleight of Hand: -2% chance to be critically hit by melee and ranged attacks per rank.
	rogue.PseudoStats.ReducedCritTakenChance += 0.02 * float64(rogue.Talents.SleightOfHand)
}
