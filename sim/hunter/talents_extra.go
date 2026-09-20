package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents added in the second audit round. Values are from the DBC (CSV's/Spell.csv) unless noted.
func (hunter *Hunter) applyAuditTalents() {
	hunter.applyMeleeSpecialization()
	hunter.applyDualWieldSpecialization()
	hunter.applyWeaponExpertise()
	hunter.applyImprovedTracking()
}

// Melee Specialization: +30% melee attack speed, -30% ranged attack speed, +1.5 sec shot time on Aimed Shot and Multi-Shot.
func (hunter *Hunter) applyMeleeSpecialization() {
	if !hunter.Talents.MeleeSpecialization {
		return
	}
	hunter.PseudoStats.MeleeSpeedMultiplier *= 1.3
	hunter.PseudoStats.RangedSpeedMultiplier *= 0.7
	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_HunterAimedShot || spell.SpellCode == SpellCode_HunterMultiShot {
			spell.DefaultCast.CastTime += time.Millisecond * 1500
		}
	})
}

// Dual Wield Specialization: +20/30/40/50% off-hand damage.
func (hunter *Hunter) applyDualWieldSpecialization() {
	if hunter.Talents.DualWieldSpecialization == 0 {
		return
	}
	mult := 1 + 0.1*float64(hunter.Talents.DualWieldSpecialization+1)
	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
			spell.DamageMultiplier *= mult
		}
	})
}

// Weapon Expertise depends on the equipped ranged weapon:
//   - Bow: 1%/rank chance for an extra arrow (extra ranged attack) after dealing damage.
//   - Crossbow: +1%/rank crit with ranged attacks.
//   - Gun: shots ignore 2 armor per level per rank.
func (hunter *Hunter) applyWeaponExpertise() {
	rank := float64(hunter.Talents.WeaponExpertise)
	ranged := hunter.GetRangedWeapon()
	if rank == 0 || ranged == nil {
		return
	}

	switch ranged.RangedWeaponType {
	case proto.RangedWeaponType_RangedWeaponTypeBow:
		procChance := 0.01 * rank
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Bow Specialization",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskRanged) {
					return
				}
				if sim.RandomFloat("Bow Specialization") < procChance {
					hunter.AutoAttacks.ExtraRangedAttack(sim, 1, core.ActionID{SpellID: 33541}, spell.ActionID)
				}
			},
		}))
	case proto.RangedWeaponType_RangedWeaponTypeCrossbow:
		bonus := rank * core.CritRatingPerCritChance
		hunter.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShot) {
				spell.BonusCritRating += bonus
			}
		})
		hunter.AutoAttacks.RangedConfig().BonusCritRating += bonus
	case proto.RangedWeaponType_RangedWeaponTypeGun:
		// Applied to all attacks, melee included, since armor penetration is a character stat here.
		hunter.AddStat(stats.ArmorPenetration, 2*rank*float64(hunter.Level))
	}
}

// Improved Tracking: +3%/rank damage to the tracked creature types, assumed always active against those types.
func (hunter *Hunter) applyImprovedTracking() {
	if hunter.Talents.ImprovedTracking == 0 {
		return
	}
	multiplier := 1 + 0.03*float64(hunter.Talents.ImprovedTracking)
	hunter.Env.RegisterPostFinalizeEffect(func() {
		for _, t := range hunter.Env.Encounter.Targets {
			switch t.MobType {
			case proto.MobType_MobTypeBeast, proto.MobType_MobTypeDemon, proto.MobType_MobTypeDragonkin,
				proto.MobType_MobTypeElemental, proto.MobType_MobTypeGiant, proto.MobType_MobTypeHumanoid,
				proto.MobType_MobTypeUndead:
				for _, at := range hunter.AttackTables[t.UnitIndex] {
					at.DamageDealtMultiplier *= multiplier
				}
			}
		}
	})
}

// Deadeye (spell 33992): 2 min cooldown, lasts until used, +100% crit on the next Aimed Shot, Multi-Shot or Arcane Shot.
func (hunter *Hunter) registerDeadeye() {
	if !hunter.Talents.Deadeye {
		return
	}

	actionID := core.ActionID{SpellID: 33992}
	critBonus := 100.0 * core.CritRatingPerCritChance
	var shots []*core.Spell
	hunter.OnSpellRegistered(func(spell *core.Spell) {
		switch spell.SpellCode {
		case SpellCode_HunterAimedShot, SpellCode_HunterMultiShot, SpellCode_HunterArcaneShot:
			shots = append(shots, spell)
		}
	})

	aura := hunter.RegisterAura(core.Aura{
		Label:    "Deadeye",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range shots {
				spell.BonusCritRating += critBonus
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range shots {
				spell.BonusCritRating -= critBonus
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_HunterAimedShot, SpellCode_HunterMultiShot, SpellCode_HunterArcaneShot:
				aura.Deactivate(sim)
			}
		},
	})

	spell := hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

// Find Weakness: ranged crits have a 20%/rank chance to add +5% crit chance to all attacks against the target for 20 sec.
func (hunter *Hunter) applyFindWeakness() {
	if hunter.Talents.FindWeakness == 0 {
		return
	}

	procChance := 0.2 * float64(hunter.Talents.FindWeakness)
	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:    "Find Weakness",
			ActionID: core.ActionID{SpellID: 33589},
			Duration: time.Second * 20,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for i := range aura.Unit.PseudoStats.SchoolCritTakenChance {
					aura.Unit.PseudoStats.SchoolCritTakenChance[i] += 0.05
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for i := range aura.Unit.PseudoStats.SchoolCritTakenChance {
					aura.Unit.PseudoStats.SchoolCritTakenChance[i] -= 0.05
				}
			},
		})
	})

	core.MakePermanent(hunter.RegisterAura(core.Aura{
		Label: "Find Weakness Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() || !spell.ProcMask.Matches(core.ProcMaskRanged) {
				return
			}
			if procChance >= 1 || sim.RandomFloat("Find Weakness") < procChance {
				auras.Get(result.Target).Activate(sim)
			}
		},
	}))
}

// Kill Command (spell 33524): pet deals +100% damage for 5 sec, only usable when the target is at 20% health or less.
// 5 sec cooldown; the mana cost is assumed free (to be checked, see docs/TODO.md).
func (hunter *Hunter) registerKillCommand() {
	if !hunter.Talents.KillCommand || hunter.pet == nil {
		return
	}

	actionID := core.ActionID{SpellID: 33524}
	petAura := hunter.pet.RegisterAura(core.Aura{
		Label:    "Kill Command",
		ActionID: core.ActionID{SpellID: 33525},
		Duration: time.Second * 5,
	}).AttachMultiplicativePseudoStatBuff(&hunter.pet.PseudoStats.DamageDealtMultiplier, 2.0)

	hunter.KillCommand = hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 5,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.IsExecutePhase20()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			petAura.Activate(sim)
		},
	})
}
