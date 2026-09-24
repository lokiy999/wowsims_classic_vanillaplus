package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const PyroblastRanks = 8

var PyroblastSpellId = [PyroblastRanks + 1]int32{0, 11366, 12505, 12522, 12523, 12524, 12525, 12526, 18809}
var PyroblastBaseDamage = [PyroblastRanks + 1][]float64{{0}, {138, 184}, {193, 249}, {270, 343}, {347, 437}, {427, 536}, {525, 653}, {625, 776}, {716, 890}}
var PyroblastDotDamage = [PyroblastRanks + 1]float64{0, 84, 108, 144, 186, 234, 282, 342, 402}
var PyroblastManaCost = [PyroblastRanks + 1]float64{0, 125, 150, 195, 240, 285, 335, 385, 440}
var PyroblastLevel = [PyroblastRanks + 1]int{0, 20, 24, 30, 36, 42, 48, 54, 60}

func (mage *Mage) registerPyroblastSpell() {
	if !mage.Talents.Pyroblast {
		return
	}

	mage.Pyroblast = make([]*core.Spell, PyroblastRanks+1)
	cdTimer := mage.NewTimer()

	for rank := 1; rank <= PyroblastRanks; rank++ {
		config := mage.newPyroblastSpellConfig(rank, cdTimer)

		if config.RequiredLevel <= int(mage.Level) {
			mage.Pyroblast[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) newPyroblastSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {

	numTicks := int32(6) // server: DoT ticks every 2 sec for 12 sec
	tickLength := time.Second * 2

	spellId := PyroblastSpellId[rank]
	baseDamageLow := PyroblastBaseDamage[rank][0]
	baseDamageHigh := PyroblastBaseDamage[rank][1]
	baseDotDamage := PyroblastDotDamage[rank] / float64(numTicks)
	manaCost := PyroblastManaCost[rank]
	level := PyroblastLevel[rank]

	spellCoeff := 1.0
	dotCoeff := .1 // .6 total like classic, spread over the 6 server ticks
	castTime := time.Second * 6

	actionID := core.ActionID{SpellID: spellId}

	spellConfig := core.SpellConfig{
		ActionID:     actionID,
		SpellCode:    SpellCode_MagePyroblast,
		SpellSchool:  core.SpellSchoolFire,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage | core.SpellFlagAPL,
		MissileSpeed: 24,

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
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Minute, // Vanilla+: 1 minute cooldown
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Pyroblast (Rank %d)", rank),
				ActionID: actionID.WithTag(1),
			},
			NumberOfTicks:    numTicks,
			TickLength:       tickLength,
			BonusCoefficient: dotCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)

				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			})
		},
	}

	return spellConfig
}
