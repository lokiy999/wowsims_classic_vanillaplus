package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerRendSpell() {

	rend := map[int32]struct {
		ticks   int32
		damage  float64
		spellID int32
	}{
		// server: 72 over 18 sec, 192 over 24, 270 over 27, 360 over 30 (3 sec ticks)
		25: {spellID: 6547, damage: 12, ticks: 6},
		40: {spellID: 11572, damage: 24, ticks: 8},
		50: {spellID: 11573, damage: 30, ticks: 9},
		60: {spellID: 11574, damage: 36, ticks: 10},
	}[warrior.Level]

	baseDamage := rend.damage

	damageMultiplier := warrior.bleedDamageMultiplier() // Two-Handed Weapon Specialization bleed bonus

	// Improved Rend (server tree): Rend stacks up to 2 / 3 times; each stack adds the full tick damage.
	maxStacks := 1 + warrior.Talents.ImprovedRend

	warrior.Rend = warrior.RegisterSpell(BattleStance|DefensiveStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorRend,
		ActionID:    core.ActionID{SpellID: rend.spellID},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   15, // server (Spell.csv)
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Rend",
				Tag:       "Rend",
				MaxStacks: maxStacks,
			},
			NumberOfTicks: rend.ticks,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage*float64(max(dot.GetStacks(), 1)), isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if result.Landed() {
				dot := spell.Dot(target)
				if maxStacks == 1 {
					dot.Apply(sim)
				} else {
					dot.ApplyOrRefresh(sim)
					if dot.GetStacks() < dot.MaxStacks {
						dot.AddStack(sim)
					}
					dot.TakeSnapshot(sim, false)
				}
			} else {
				spell.IssueRefund(sim)
			}

			spell.DealOutcome(sim, result)
		},
	})

}
