package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const HolyFireRanks = 8

var HolyFireSpellId = [HolyFireRanks + 1]int32{0, 14914, 15262, 15263, 15264, 15265, 15266, 15267, 15261}

// CSV's/Spell.csv: EffectBasePoints[direct]+1 to EffectBasePoints[direct]+EffectDieSides[0]
// (die[0] is the direct-hit role regardless of which bp slot holds it per rank - rank 3
// keeps its direct effect in slot 0 instead of slot 1 like every other rank, but die[0]
// still pairs with it). Rank 8 (481-604) confirmed against the in-game tooltip.
var HolyFireBaseDamage = [HolyFireRanks + 1][]float64{{0}, {114, 140}, {144, 176}, {191, 236}, {237, 295}, {293, 364}, {363, 452}, {435, 542}, {481, 604}}

// CSV's/Spell.csv: (EffectBasePoints[dot]+1) * 5 ticks. Rank 8 (225) confirmed
// against the in-game tooltip.
var HolyFireDotDamage = [HolyFireRanks + 1]float64{0, 30, 50, 75, 100, 125, 150, 175, 225}
var HolyFireSpellCoef = [HolyFireRanks + 1]float64{0, 0.123, 0.271, 0.554, 0.714, 0.714, 0.714, 0.714, 0.714}
var HolyFireManaCost = [HolyFireRanks + 1]float64{0, 85, 95, 125, 145, 170, 200, 230, 255}
var HolyFireLevel = [HolyFireRanks + 1]int{0, 20, 24, 30, 36, 42, 48, 54, 60}

func (priest *Priest) registerHolyFire() {
	priest.HolyFire = make([]*core.Spell, HolyFireRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= HolyFireRanks; rank++ {
		config := priest.getHolyFireConfig(rank, cdTimer)

		if config.RequiredLevel <= int(priest.Level) {
			priest.HolyFire[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

func (priest *Priest) getHolyFireConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	ticks := int32(5)

	spellId := HolyFireSpellId[rank]
	baseDamageLow := HolyFireBaseDamage[rank][0]
	baseDamageHigh := HolyFireBaseDamage[rank][1]
	dotDamage := HolyFireDotDamage[rank] / float64(ticks)
	manaCost := HolyFireManaCost[rank]
	level := HolyFireLevel[rank]

	// Hybrid direct+DoT coefficient split (see docs/private-server-spell-rules.md
	// "Deriving spell coefficients for hybrid direct+DoT spells"):
	// castCoeff = castTime/3.5 = 3.5/3.5 = 1.0
	// dotCoeff_raw = dotDuration/15 = 10/15 = 2/3
	// direct = castCoeff^2 / (castCoeff+dotCoeff_raw) = 1 / (5/3) = 3/5 = 0.6
	// dot (total) = dotCoeff_raw^2 / (castCoeff+dotCoeff_raw) = (4/9) / (5/3) = 4/15,
	//   per-tick (5 ticks) = 4/75 = 0.05333...
	// Empirically confirmed against live in-game samples 2026-09-16 (fits better
	// than the old undocumented 0.75/0.05 pair on both components).
	directCoeff := 3.0 / 5.0
	dotCoeff := 4.0 / 75.0
	castTime := time.Millisecond * 3500

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestHolyFire,
		ActionID:    core.ActionID{SpellID: spellId},
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
				CastTime: castTime - time.Millisecond*100*time.Duration(priest.Talents.DivineFury),
			},
			// In-game tooltip confirms a 10 sec cooldown; this was previously missing entirely.
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 10,
			},
		},

		BonusCoefficient: directCoeff,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Holy Fire (Rank %d)", rank),
			},

			NumberOfTicks:    ticks,
			TickLength:       time.Second * 2,
			BonusCoefficient: dotCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, dotDamage, isRollover)
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealDamage(sim, result)
		},
	}
}
