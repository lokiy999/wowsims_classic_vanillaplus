package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
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

	rogue.applyDeathMark()
}

// Death Mark (DBC 33694/33695, marks 33696/33697 and 35827/35828): Ambush and Garrote mark the target for
// 10/25 sec; the marked target has +10% chance to be critically hit by melee, ranged and spells (from everyone).
// Cheap Shot is not in the sim. The tracking and movement parts are not modeled.
func (rogue *Rogue) applyDeathMark() {
	rank := min(rogue.Talents.DeathMark, 2)
	if rank <= 0 {
		return
	}
	actionID := core.ActionID{SpellID: []int32{0, 35827, 35828}[rank]}
	duration := []time.Duration{0, time.Second * 10, time.Second * 25}[rank]

	marks := rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Death Mark",
			ActionID: actionID,
			Duration: duration,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for i := range aura.Unit.PseudoStats.SchoolCritTakenChance {
					aura.Unit.PseudoStats.SchoolCritTakenChance[i] += 0.10
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for i := range aura.Unit.PseudoStats.SchoolCritTakenChance {
					aura.Unit.PseudoStats.SchoolCritTakenChance[i] -= 0.10
				}
			},
		})
	})

	core.MakePermanent(rogue.RegisterAura(core.Aura{
		Label: "Death Mark Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && (spell.SpellCode == SpellCode_RogueAmbush || spell.SpellCode == SpellCode_RogueGarrote) {
				marks.Get(result.Target).Activate(sim)
			}
		},
	}))
}
