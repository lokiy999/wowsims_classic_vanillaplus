package paladin

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func (paladin *Paladin) registerBlessingOfSanctuary() {
	if paladin.Options.PersonalBlessing != proto.Blessings_BlessingOfSanctuary {
		return
	}

	sanctuaryValues := []struct {
		minLevel int32
		maxLevel int32
		spellID  int32
		absorb   float64
		damage   float64
	}{
		// Server (Spell.csv 20911-20914): damage taken reduced by 10/15/20/30, attackers take the same Holy damage.
		{minLevel: 30, maxLevel: 39, spellID: 20911, absorb: 10, damage: 10},
		{minLevel: 40, maxLevel: 49, spellID: 20912, absorb: 15, damage: 15},
		{minLevel: 50, maxLevel: 59, spellID: 20913, absorb: 20, damage: 20},
		{minLevel: 60, maxLevel: 60, spellID: 20914, absorb: 30, damage: 30},
	}

	for i, values := range sanctuaryValues {

		if (values.minLevel <= paladin.Level) && (paladin.Level <= values.maxLevel) {

			rank := i + 1
			actionID := core.ActionID{SpellID: values.spellID}
			// Guardian's Favor: the effect of Blessing of Sanctuary +10% per rank.
			guardiansFavor := 1 + 0.1*float64(paladin.Talents.GuardiansFavor)
			damage := values.damage * guardiansFavor
			values.absorb *= guardiansFavor

			sanctuaryProc := paladin.RegisterSpell(core.SpellConfig{
				ActionID:    actionID,
				SpellSchool: core.SpellSchoolHoly,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskSpellDamage,
				Flags:       core.SpellFlagIgnoreResists,

				Rank: rank,

				DamageMultiplier: 1,
				ThreatMultiplier: 1,

				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
				},
			})

			paladin.RegisterAura(core.Aura{
				Label:    "Blessing of Sanctuary Trigger",
				Duration: core.NeverExpires,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					aura.Activate(sim)
				},
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					for i := range paladin.PseudoStats.BonusDamageTakenBeforeModifiers {
						paladin.PseudoStats.BonusDamageTakenBeforeModifiers[i] -= values.absorb
					}
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					for i := range paladin.PseudoStats.BonusDamageTakenBeforeModifiers {
						paladin.PseudoStats.BonusDamageTakenBeforeModifiers[i] += values.absorb
					}
				},
				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if result.DidBlock() {
						sanctuaryProc.Cast(sim, spell.Unit)
					}
				},
			})

		}
	}
}
