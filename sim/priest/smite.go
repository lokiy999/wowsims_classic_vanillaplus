package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const SmiteRanks = 8

var SmiteSpellId = [SmiteRanks + 1]int32{0, 585, 591, 598, 984, 1004, 6060, 10933, 10934}

// CSV's/Spell.csv, same EffectBasePoints[1]+EffectDieSides[0] pattern validated
// against the in-game Mind Blast rank 9 tooltip (sim/priest/mind_blast.go). Not
// independently spot-checked in-game for Smite specifically - verify if in doubt.
var SmiteBaseDamage = [SmiteRanks + 1][]float64{{0}, {13, 17}, {28, 34}, {58, 66}, {97, 111}, {158, 178}, {222, 250}, {298, 334}, {384, 428}}
var SmiteSpellCoef = [SmiteRanks + 1]float64{0, 0.123, 0.271, 0.554, 0.714, 0.714, 0.714, 0.714, 0.714}
var SmiteCastTime = [SmiteRanks + 1]int{0, 1500, 2000, 2500, 2500, 2500, 2500, 2500, 2500}
var SmiteManaCost = [SmiteRanks + 1]float64{0, 20, 30, 60, 95, 140, 185, 230, 280}
var SmiteLevel = [SmiteRanks + 1]int{0, 1, 6, 14, 22, 30, 38, 46, 54}

func (priest *Priest) registerSmiteSpell() {
	priest.Smite = make([]*core.Spell, SmiteRanks+1)

	for rank := 1; rank <= SmiteRanks; rank++ {
		config := priest.getSmiteBaseConfig(rank)

		if config.RequiredLevel <= int(priest.Level) {
			priest.Smite[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

func (priest *Priest) getSmiteBaseConfig(rank int) core.SpellConfig {
	spellId := SmiteSpellId[rank]
	baseDamageLow := SmiteBaseDamage[rank][0]
	baseDamageHigh := SmiteBaseDamage[rank][1]
	spellCoeff := SmiteSpellCoef[rank]
	castTime := SmiteCastTime[rank]
	manaCost := SmiteManaCost[rank]
	level := SmiteLevel[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_PriestSmite,
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*time.Duration(castTime) - time.Millisecond*100*time.Duration(priest.Talents.DivineFury),
			},
		},

		BonusCoefficient: spellCoeff,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	}
}
