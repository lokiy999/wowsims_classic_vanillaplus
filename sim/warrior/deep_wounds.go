package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

const deepWoundsMaxStacks = 5 // server talent text: "Stacks up to 5 times"

func (warrior *Warrior) applyDeepWounds() {
	if warrior.Talents.DeepWounds == 0 {
		return
	}

	spellID := map[int32]int32{
		1: 12834,
		2: 12849,
		3: 12867,
	}[warrior.Talents.DeepWounds]

	warrior.DeepWounds = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorDeepWounds,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: warrior.bleedDamageMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Deep Wounds",
				MaxStacks: deepWoundsMaxStacks,
			},
			NumberOfTicks: 4,
			TickLength:    time.Second * 3,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				attackTable := warrior.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
				dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(attackTable, true) // Double dips on attackers mods
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dot := spell.Dot(target)
			stacks := min(dot.GetStacks()+1, deepWoundsMaxStacks)
			dot.Apply(sim) //Resets the tick counter with Apply vs ApplyorRefresh
			dot.SetStacks(sim, stacks)
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHitNoHitCounter)
		},
	})

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Deep Wounds Talent",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskEmpty) || !spell.SpellSchool.Matches(core.SpellSchoolPhysical) {
				return
			}

			// Ravager doesn't proc Deep Wounds
			if spell.ActionID.SpellID == 9633 {
				return
			}

			if result.Outcome.Matches(core.OutcomeCrit) {
				warrior.procDeepWounds(sim, result.Target, spell.IsOH())
			}
		},
	}))
}

func (warrior *Warrior) procDeepWounds(sim *core.Simulation, target *core.Unit, isOh bool) {
	dot := warrior.DeepWounds.Dot(target)

	var awd float64
	if isOh {
		attackTableOh := warrior.AttackTables[target.UnitIndex][proto.CastType_CastTypeOffHand]
		adm := warrior.AutoAttacks.OHAuto().AttackerDamageMultiplier(attackTableOh, true)
		awd = warrior.AutoAttacks.OH().CalculateAverageWeaponDamage(dot.Spell.MeleeAttackPower()) * 0.5 * adm
	} else { // MH
		attackTableMh := warrior.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
		adm := warrior.AutoAttacks.MHAuto().AttackerDamageMultiplier(attackTableMh, true)
		awd = warrior.AutoAttacks.MH().CalculateAverageWeaponDamage(dot.Spell.MeleeAttackPower()) * adm
	}

	newDamage := awd * 0.1 * float64(warrior.Talents.DeepWounds) // DBC: 10%/20%/30% of the weapon's average damage

	// Server: stacks up to 5 times. Each crit adds its own bleed (spread over the 4 ticks) and refreshes the duration;
	// at 5 stacks the new one replaces an average stack.
	perTick := newDamage / 4.0
	switch {
	case !dot.IsActive():
		dot.SnapshotBaseDamage = perTick
	case dot.GetStacks() < deepWoundsMaxStacks:
		dot.SnapshotBaseDamage += perTick
	default:
		dot.SnapshotBaseDamage = dot.SnapshotBaseDamage*float64(deepWoundsMaxStacks-1)/float64(deepWoundsMaxStacks) + perTick
	}
	dot.SnapshotAttackerMultiplier = 1

	warrior.DeepWounds.Cast(sim, target)
}
