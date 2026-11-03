package druid

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Hurricane is a talent on the server, Moonkin Form only: a 2 sec cast (not channeled) that leaves a 10 sec storm, so
// the druid can keep casting while it ticks. Ranks from the server data (16914 / 17401 / 17402): damage per second
// grows per level up to the rank's max level; 2 min cooldown.
func (druid *Druid) registerHurricaneSpell() {
	if !druid.Talents.Hurricane {
		return
	}

	ranks := []struct {
		level      int32
		spellID    int32
		manaCost   float64
		scaleLevel int32
		damage     float64
		scale      float64
	}{
		{level: 20, spellID: 16914, manaCost: 500, scaleLevel: 30, damage: 22, scale: 0.2},
		{level: 40, spellID: 17401, manaCost: 880, scaleLevel: 50, damage: 70, scale: 0.4},
		{level: 60, spellID: 17402, manaCost: 1495, scaleLevel: 70, damage: 134, scale: 0.6},
	}

	for i, rank := range ranks {
		if druid.Level < rank.level {
			break
		}

		damage := rank.damage + float64(min(druid.Level, rank.scaleLevel)-rank.level)*rank.scale
		spell := druid.RegisterSpell(Moonkin, core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       SpellFlagOmen | core.SpellFlagBinary | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					CastTime: time.Second * 2,
					GCD:      core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    druid.NewTimer(),
					Duration: time.Minute * 2, // 2 minute cooldown
				},
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			Dot: core.DotConfig{
				IsAOE: true,
				Aura: core.Aura{
					Label: "Hurricane" + druid.Label + strconv.Itoa(i+1),
				},
				NumberOfTicks: 10,
				TickLength:    time.Second * 1,

				BonusCoefficient: 0.03,

				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.Snapshot(target, damage, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					for _, aoeTarget := range sim.Encounter.TargetUnits {
						dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
					}
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				spell.AOEDot().Apply(sim)
			},
		})

		druid.Hurricane = append(druid.Hurricane, spell)
	}
}
