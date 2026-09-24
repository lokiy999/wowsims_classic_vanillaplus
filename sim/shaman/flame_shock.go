package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FlameShockRanks = 6

var FlameShockSpellId = [FlameShockRanks + 1]int32{0, 8050, 8052, 8053, 10447, 10448, 29228}
var FlameShockBaseDamage = [FlameShockRanks + 1]float64{0, 29, 56, 108, 191, 274, 320}
var FlameShockBaseDotDamage = [FlameShockRanks + 1]float64{0, 50, 75, 150, 250, 350, 450}
var FlameShockBaseSpellCoef = [FlameShockRanks + 1]float64{0, .134, .198, .214, .214, .214, .214}
var FlameShockDotSpellCoef = [FlameShockRanks + 1]float64{0, .063, .093, .1, .1, .1, .1}
var FlameShockManaCost = [FlameShockRanks + 1]float64{0, 45, 80, 135, 210, 295, 340}
var FlameShockLevel = [FlameShockRanks + 1]int{0, 10, 18, 28, 40, 52, 60}

func (shaman *Shaman) registerFlameShockSpell(shockTimer *core.Timer) {
	shaman.FlameShock = make([]*core.Spell, FlameShockRanks+1)

	for rank := 1; rank <= FlameShockRanks; rank++ {
		if FlameShockLevel[rank] <= int(shaman.Level) {
			shaman.FlameShock[rank] = shaman.RegisterSpell(shaman.newFlameShockSpell(rank, shockTimer))
		}
	}
}

func (shaman *Shaman) newFlameShockSpell(rank int, shockTimer *core.Timer) core.SpellConfig {
	numTicks := 5 // server: 15 sec DoT
	tickDuration := time.Second * 3

	spellId := FlameShockSpellId[rank]
	baseDamage := FlameShockBaseDamage[rank]
	baseDotDamage := FlameShockBaseDotDamage[rank] / float64(numTicks)
	baseSpellCoeff := FlameShockBaseSpellCoef[rank]
	dotSpellCoeff := FlameShockDotSpellCoef[rank] * 4 / float64(numTicks) // classic total, spread over the server ticks
	manaCost := FlameShockManaCost[rank]
	level := FlameShockLevel[rank]

	spell := shaman.newShockSpellConfig(
		core.ActionID{SpellID: spellId},
		core.SpellSchoolFire,
		manaCost,
		shockTimer,
	)

	spell.SpellCode = SpellCode_ShamanFlameShock
	spell.RequiredLevel = level
	spell.Rank = rank

	spell.Cast.IgnoreHaste = true

	spell.BonusCoefficient = baseSpellCoeff

	spell.Dot = core.DotConfig{
		Aura: core.Aura{
			Label: fmt.Sprintf("Flame Shock (Rank %d)", rank),
		},

		NumberOfTicks:    int32(numTicks),
		TickLength:       tickDuration,
		BonusCoefficient: dotSpellCoeff,

		OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
			dot.Snapshot(target, baseDotDamage, isRollover)
		},

		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
		},
	}

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(result.Target).Apply(sim)
		}
	}

	return spell
}
