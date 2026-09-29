package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const SearingPainRanks = 6

func (warlock *Warlock) getSearingPainBaseConfig(rank int) core.SpellConfig {
	spellCoeff := [SearingPainRanks + 1]float64{0, .396, .429, .429, .429, .429, .429}[rank]
	// Server (Spell.csv) at level 60: base damage plus damage per level up to the rank's max level.
	baseDamage := [SearingPainRanks + 1][]float64{{0}, {50, 61}, {87, 103}, {124, 148}, {174, 206}, {224, 264}, {276, 324}}[rank]
	spellId := [SearingPainRanks + 1]int32{0, 5676, 17919, 17920, 17921, 17922, 17923}[rank]
	manaCost := [SearingPainRanks + 1]float64{0, 45, 68, 91, 118, 141, 168}[rank]
	level := [SearingPainRanks + 1]int{0, 18, 26, 34, 42, 50, 58}[rank]
	castTime := time.Millisecond * 1500

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockSearingPain,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagDestruction,
		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
		},
		BonusCritRating: 2.0 * float64(int32(0) /*removed*/) * core.CritRatingPerCritChance,

		DamageMultiplier: 1,
		ThreatMultiplier: 2,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := sim.Roll(baseDamage[0], baseDamage[1])
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
		},
	}
}

func (warlock *Warlock) registerSearingPainSpell() {
	warlock.SearingPain = make([]*core.Spell, 0)
	for rank := 1; rank <= SearingPainRanks; rank++ {
		config := warlock.getSearingPainBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.SearingPain = append(warlock.SearingPain, warlock.GetOrRegisterSpell(config))
		}
	}
}
