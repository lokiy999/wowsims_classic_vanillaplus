package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ConflagrateRanks = 4

func (warlock *Warlock) getConflagrateConfig(rank int) core.SpellConfig {
	spellId := [ConflagrateRanks + 1]int32{0, 17962, 18930, 18931, 18932}[rank]
	// Server (Spell.csv) at level 60 (per level +1.6/+1.8/+2.0 up to max level 36/46/56).
	baseDamageMin := [ConflagrateRanks + 1]float64{0, 170, 236, 307, 447}[rank]
	baseDamageMax := [ConflagrateRanks + 1]float64{0, 236, 316, 403, 557}[rank]
	manaCost := [ConflagrateRanks + 1]float64{0, 110, 140, 175, 255}[rank]
	level := [ConflagrateRanks + 1]int{0, 0, 40, 50, 60}[rank]

	spCoeff := 0.429

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockConflagrate,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | WarlockFlagDestruction,
		Rank:          rank,
		RequiredLevel: level,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer: warlock.NewTimer(),
				// The talent spell (35737) has 10 sec, confirmed in game (Part AU); the ranked spells have 15 sec.
				Duration: time.Second * 10,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.getActiveImmolateSpell(target) != nil
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageMin, baseDamageMax)

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			immoSpell := warlock.getActiveImmolateSpell(target)
			if immoSpell != nil {
				immoSpell.Dot(target).Deactivate(sim)
			}
		},
	}
}

func (warlock *Warlock) registerConflagrateSpell() {
	if !warlock.Talents.Conflagrate {
		return
	}

	warlock.Conflagrate = make([]*core.Spell, 0)
	for rank := 1; rank <= ConflagrateRanks; rank++ {
		config := warlock.getConflagrateConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.Conflagrate = append(warlock.Conflagrate, warlock.GetOrRegisterSpell(config))
		}
	}
}
