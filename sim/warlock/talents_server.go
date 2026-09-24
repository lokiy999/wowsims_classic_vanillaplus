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

	warlock.applyDefiler()
}

// Defiler (DBC 34305): time between periodic ticks and duration -20%, global cooldown -0.3 sec on the
// warlock's own spells (not the pet's). The number of ticks stays the same, so DoTs and channels deal the
// same damage in less time. The extra Drain target is not modeled (single target sim).
func (warlock *Warlock) applyDefiler() {
	if !warlock.Talents.Defiler {
		return
	}
	warlock.GetOrRegisterAura(core.Aura{
		Label: "Defiler",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.Spellbook {
				if spell.DefaultCast.GCD > 0 {
					spell.DefaultCast.GCD = max(core.GCDMin, spell.DefaultCast.GCD-time.Millisecond*300)
				}
				dots := append([]*core.Dot{spell.AOEDot()}, spell.Dots()...)
				for _, dot := range dots {
					if dot != nil {
						dot.TickLength = time.Duration(float64(dot.TickLength) * 0.8)
						dot.RecomputeAuraDuration()
					}
				}
			}
		},
	})
}

// Prolonged Misery: Corruption, Curse of Agony, Curse of Exhaustion and Immolate last 2 sec longer per rank.
// Returns the extra ticks for a DoT with the given tick length.
func (warlock *Warlock) prolongedMiseryTicks(tickLength time.Duration) int32 {
	return int32(time.Duration(warlock.Talents.ProlongedMisery) * 2 * time.Second / tickLength)
}
