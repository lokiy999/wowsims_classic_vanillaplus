package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Talents added in the audit follow-ups (see docs/CHANGES.md Part BE).
func (paladin *Paladin) applyAuditTalents() {
	paladin.applyImprovedPurifying()
	paladin.applyHolyGrasp()

	// Blessed Strikes (DBC 33472-33476): -4%/rank threat from all actions, except under Righteous Fury.
	if paladin.Talents.BlessedStrikes > 0 && !paladin.Options.RighteousFury {
		paladin.PseudoStats.ThreatMultiplier *= 1 - 0.04*float64(paladin.Talents.BlessedStrikes)
	}
}

// Improved Purifying (DBC 33445-33447): -15%/rank mana cost of Consecration, Exorcism, Holy Wrath and Hammer of Wrath, and
// +15%/rank crit chance on Exorcism, Holy Wrath and Hammer of Wrath.
func (paladin *Paladin) applyImprovedPurifying() {
	rank := paladin.Talents.ImprovedPurifying
	if rank == 0 {
		return
	}

	crit := 15 * float64(rank) * core.SpellCritRatingPerCritChance
	paladin.OnSpellRegistered(func(spell *core.Spell) {
		switch spell.SpellCode {
		case SpellCode_PaladinConsecration:
			if spell.Cost != nil {
				spell.Cost.Multiplier -= 15 * rank
			}
		case SpellCode_PaladinExorcism, SpellCode_PaladinHolyWrath, SpellCode_PaladinHammerOfWrath:
			if spell.Cost != nil {
				spell.Cost.Multiplier -= 15 * rank
			}
			spell.BonusCritRating += crit
		}
	})
}

// Holy Grasp (DBC 20359-20361): -0.5 sec/rank cast time of Holy Wrath and Hammer of Wrath.
func (paladin *Paladin) applyHolyGrasp() {
	rank := paladin.Talents.HolyGrasp
	if rank == 0 {
		return
	}

	reduction := time.Millisecond * 500 * time.Duration(rank)
	paladin.OnSpellRegistered(func(spell *core.Spell) {
		switch spell.SpellCode {
		case SpellCode_PaladinHolyWrath, SpellCode_PaladinHammerOfWrath:
			spell.DefaultCast.CastTime = max(0, spell.DefaultCast.CastTime-reduction)
		}
	})
}

// Improved Retribution Aura (DBC 20091/20092): while the raid has a Retribution Aura, it deals additional Holy damage to the
// attacker equal to 5%/10% of the damage the paladin takes.
func (paladin *Paladin) applyImprovedRetributionAura(raidBuffs *proto.RaidBuffs) {
	if paladin.Talents.ImprovedRetributionAura == 0 || raidBuffs.RetributionAura == proto.TristateEffect_TristateEffectMissing {
		return
	}

	fraction := 0.05 * float64(paladin.Talents.ImprovedRetributionAura)

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 20091},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagBinary | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {},
	})

	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label: "Improved Retribution Aura",
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && result.Damage > 0 && spell.ProcMask.Matches(core.ProcMaskMelee) {
				procSpell.CalcAndDealDamage(sim, spell.Unit, fraction*result.Damage, procSpell.OutcomeMagicHit)
			}
		},
	}))
}
