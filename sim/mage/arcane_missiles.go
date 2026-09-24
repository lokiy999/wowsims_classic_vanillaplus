package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const ArcaneMissilesRanks = 8

var ArcaneMissilesSpellId = [ArcaneMissilesRanks + 1]int32{0, 5143, 5144, 5145, 8416, 8417, 10211, 10212, 25345}
var ArcaneMissilesBaseTickDamage = [ArcaneMissilesRanks + 1]float64{0, 12, 18, 28, 42, 58, 76, 96, 115} // server: per missile, 10 missiles
// Per missile: the classic total (3/4/5 missiles) spread over the 10 server missiles.
var ArcaneMissilesSpellCoeff = [ArcaneMissilesRanks + 1]float64{0, .0396, .0816, .12, .12, .12, .12, .12, .12}
var ArcaneMissilesCastTime = [ArcaneMissilesRanks + 1]int32{0, 3, 4, 5, 5, 5, 5, 5, 5}
var ArcaneMissilesManaCost = [ArcaneMissilesRanks + 1]float64{0, 85, 140, 235, 320, 410, 500, 595, 655}
var ArcaneMissilesLevel = [ArcaneMissilesRanks + 1]int{0, 8, 16, 24, 32, 40, 48, 56, 56}

func (mage *Mage) registerArcaneMissilesSpell() {
	// Server: each cast "energizes you, increasing your chance to get a critical strike with Arcane
	// spells by 1% for 20 sec. The effect stacks up to 5 times." (one stack per cast)
	mage.arcaneMissilesEnergize = mage.RegisterAura(core.Aura{
		Label:     "Arcane Missiles Energize",
		ActionID:  core.ActionID{SpellID: 25345}.WithTag(2),
		Duration:  time.Second * 20,
		MaxStacks: 5,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			aura.Unit.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexArcane] += float64(newStacks-oldStacks) * core.SpellCritRatingPerCritChance
		},
	})

	mage.ArcaneMissiles = make([]*core.Spell, ArcaneMissilesRanks+1)
	mage.ArcaneMissilesTickSpell = make([]*core.Spell, ArcaneMissilesRanks+1)

	// TODO AQ <=
	for rank := 1; rank < ArcaneMissilesRanks; rank++ {
		config := mage.getArcaneMissilesSpellConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.ArcaneMissiles[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) getArcaneMissilesSpellConfig(rank int) core.SpellConfig {
	spellId := ArcaneMissilesSpellId[rank]
	baseTickDamage := ArcaneMissilesBaseTickDamage[rank]
	castTime := ArcaneMissilesCastTime[rank]
	manaCost := ArcaneMissilesManaCost[rank]
	level := ArcaneMissilesLevel[rank]

	_ = castTime
	numTicks := int32(10) // server: 10 missiles over 5 sec on every rank
	tickLength := time.Millisecond * 500

	tickSpell := mage.getArcaneMissilesTickSpell(rank)
	mage.ArcaneMissilesTickSpell[rank] = tickSpell

	return core.SpellConfig{
		SpellCode:   SpellCode_MageArcaneMissiles,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | core.SpellFlagAPL | core.SpellFlagChanneled | core.SpellFlagNoMetrics,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("ArcaneMissiles-%d-%d", +rank, numTicks),
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					// TODO: This check is necessary to ensure the final tick occurs before
					// Arcane Blast stacks are dropped. To fix this, ticks need to reliably
					// occur before aura expirations.

					//TODO: Test interaction in classic code without aura
					dot := mage.ArcaneMissiles[rank].Dot(aura.Unit)
					if dot.TickCount < dot.NumberOfTicks {
						dot.TickCount++
						dot.TickOnce(sim)
					}
				},
			},
			NumberOfTicks: numTicks,
			TickLength:    tickLength,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				tickSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
			mage.arcaneMissilesEnergize.Activate(sim)
			mage.arcaneMissilesEnergize.AddStack(sim)
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			return tickSpell.CalcDamage(sim, target, baseTickDamage, spell.OutcomeExpectedMagicHitAndCrit)
		},
	}
}

func (mage *Mage) getArcaneMissilesTickSpell(rank int) *core.Spell {
	spellId := ArcaneMissilesSpellId[rank]
	baseTickDamage := ArcaneMissilesBaseTickDamage[rank]
	spellCoeff := ArcaneMissilesSpellCoeff[rank]

	return mage.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_MageArcaneMissilesTick,
		ActionID:     core.ActionID{SpellID: spellId}.WithTag(1),
		SpellSchool:  core.SpellSchoolArcane,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage,
		MissileSpeed: 20,

		Rank: 1,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, baseTickDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
