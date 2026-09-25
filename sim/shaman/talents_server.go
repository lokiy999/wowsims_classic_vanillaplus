package shaman

import (
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/shaman.json).
func (shaman *Shaman) applyServerTalents() {
	// Rockhide: all damage taken -2% per rank (the chance to hit and stun nearby enemies is not modeled).
	shaman.PseudoStats.DamageTakenMultiplier *= 1 - 0.02*float64(shaman.Talents.Rockhide)

	// Primal Endurance: +2% total Health per rank (the fatal damage reduction is not modeled).
	if shaman.Talents.PrimalEndurance > 0 {
		shaman.MultiplyStat(stats.Health, 1+0.02*float64(shaman.Talents.PrimalEndurance))
	}

	// Nature's Grace: threat -10% per rank.
	shaman.PseudoStats.ThreatMultiplier *= 1 - 0.1*float64(shaman.Talents.NatureGrace)

	// Nature's Guardian: -2% chance to be critically hit per rank.
	shaman.PseudoStats.ReducedCritTakenChance += 0.02 * float64(shaman.Talents.NaturesGuardian)

	// Meditation: 10% of mana regeneration per rank continues while casting.
	shaman.PseudoStats.SpiritRegenRateCasting += 0.1 * float64(shaman.Talents.Meditation)
}
