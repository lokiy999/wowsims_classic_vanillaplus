package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/warlock.json).
func (warlock *Warlock) applyServerTalents() {
	// Demonic Onslaught: +4% melee and spell crit per rank for the Imp, Voidwalker, Succubus and Felhunter.
	if warlock.Talents.DemonicOnslaught > 0 {
		crit := 4 * float64(warlock.Talents.DemonicOnslaught)
		for _, pet := range warlock.BasePets {
			pet.AddStats(stats.Stats{
				stats.MeleeCrit: crit * core.CritRatingPerCritChance,
				stats.SpellCrit: crit * core.SpellCritRatingPerCritChance,
			})
		}
	}

	// Destructive Reach: threat of Destruction spells -10% per rank.
	if warlock.Talents.DestructiveReach > 0 {
		mult := 1 - 0.1*float64(warlock.Talents.DestructiveReach)
		warlock.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_WarlockShadowBolt, SpellCode_WarlockImmolate, SpellCode_WarlockSearingPain,
				SpellCode_WarlockSoulFire, SpellCode_WarlockConflagrate, SpellCode_WarlockShadowburn:
				spell.ThreatMultiplier *= mult
			}
		})
	}
}

// Prolonged Misery: Corruption, Curse of Agony, Curse of Exhaustion and Immolate last 2 sec longer per rank.
// Returns the extra ticks for a DoT with the given tick length.
func (warlock *Warlock) prolongedMiseryTicks(tickLength time.Duration) int32 {
	return int32(time.Duration(warlock.Talents.ProlongedMisery) * 2 * time.Second / tickLength)
}
