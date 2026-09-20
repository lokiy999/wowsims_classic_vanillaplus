package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Aspects other than Hawk, and the Savage Blow and Whirling Axe talents.
// Only the pieces Savage Blow needs are modeled for the aspects: they exist as exclusive auras, their stat bonuses are
// not applied yet (see docs/TODO.md).

const savageBlowManaPct = 0.12 // 12% of base mana

func (hunter *Hunter) registerExtraAspects() {
	makeAspect := func(spellID int32, level int32, mana float64, label string) *core.Aura {
		if hunter.Level < level {
			return nil
		}
		aura := hunter.RegisterAura(core.Aura{
			Label:    label,
			ActionID: core.ActionID{SpellID: spellID},
			Duration: core.NeverExpires,
		})
		aura.NewExclusiveEffect("Aspect", true, core.ExclusiveEffect{})

		hunter.RegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{SpellID: spellID},
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
			ManaCost: core.ManaCostOptions{FlatCost: mana},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{GCD: core.GCDDefault},
			},
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return !aura.IsActive()
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})
		return aura
	}

	hunter.aspectOfTheMonkey = makeAspect(13163, 4, 20, "Aspect of the Monkey")
	if hunter.aspectOfTheMonkey != nil {
		// +5% dodge and +5% melee crit (confirmed in game); Improved Aspect of the Monkey adds 1% per rank (DBC 19549-19551).
		bonus := 5 + float64(hunter.Talents.ImprovedAspectOfTheMonkey)
		hunter.aspectOfTheMonkey.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatsDynamic(sim, stats.Stats{
				stats.Dodge:     bonus * core.DodgeRatingPerDodgeChance,
				stats.MeleeCrit: bonus * core.CritRatingPerCritChance,
			})
		}).ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatsDynamic(sim, stats.Stats{
				stats.Dodge:     -bonus * core.DodgeRatingPerDodgeChance,
				stats.MeleeCrit: -bonus * core.CritRatingPerCritChance,
			})
		})
	}
	hunter.aspectOfTheBeast = makeAspect(13161, 30, 50, "Aspect of the Beast")
	hunter.aspectOfThePack = makeAspect(13159, 40, 100, "Aspect of the Pack")
}

// Savage Blow (spell 33590): 12% base mana, 6 sec cooldown, instant. A strike with both weapons whose bonus depends on the
// active aspect: Beast hits additional nearby enemies, Monkey allows Mongoose Bite, Pack reduces costs by 10% for 20 sec.
// The damage is a placeholder (one main-hand and one off-hand weapon hit), see docs/TODO.md.
func (hunter *Hunter) registerSavageBlow() {
	if !hunter.Talents.SavageBlow || hunter.Level < 40 {
		return
	}

	hunter.registerExtraAspects()

	actionID := core.ActionID{SpellID: 33590}

	packAura := hunter.RegisterAura(core.Aura{
		Label:    "Savage Blow Pack",
		ActionID: core.ActionID{SpellID: 34237},
		Duration: time.Second * 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range hunter.Spellbook {
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 10
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range hunter.Spellbook {
				if spell.Cost != nil {
					spell.Cost.Multiplier += 10
				}
			}
		},
	})

	mhHit := hunter.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(1),
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		CritDamageBonus:  hunter.mortalShots(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := hunter.MHWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})

	ohHit := hunter.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(2),
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeOHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		CritDamageBonus:  hunter.mortalShots(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := hunter.OHWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})

	strike := func(sim *core.Simulation, target *core.Unit) {
		mhHit.Cast(sim, target)
		if hunter.AutoAttacks.IsDualWielding {
			ohHit.Cast(sim, target)
		}
	}

	hunter.SavageBlow = hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,

		ManaCost: core.ManaCostOptions{
			BaseCost: savageBlowManaPct,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.DistanceFromTarget <= core.MaxMeleeAttackDistance
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			strike(sim, target)

			switch {
			case hunter.aspectOfTheBeast != nil && hunter.aspectOfTheBeast.IsActive():
				for _, other := range sim.Encounter.TargetUnits {
					if other != target {
						strike(sim, other)
					}
				}
			case hunter.aspectOfTheMonkey != nil && hunter.aspectOfTheMonkey.IsActive():
				hunter.DefensiveState.Activate(sim)
			case hunter.aspectOfThePack != nil && hunter.aspectOfThePack.IsActive():
				packAura.Activate(sim)
			}
		},
	})
}

// Whirling Axe (spell 34216): 10% of melee attack power as physical damage, always crits, slows by 60% for 10 sec and
// interrupts the school for 5 sec (the slow and the interrupt are not modeled). No weapon needed. Cooldown, mana cost
// and range are unknown, so it has none (see docs/TODO.md).
func (hunter *Hunter) registerWhirlingAxe() {
	if !hunter.Talents.WhirlingAxe || hunter.Level < 40 {
		return
	}

	hunter.WhirlingAxe = hunter.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 34216},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
		},

		CritDamageBonus:  hunter.mortalShots(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 0.1*spell.MeleeAttackPower(), spell.OutcomeMeleeSpecialCritOnly)
		},
	})
}
