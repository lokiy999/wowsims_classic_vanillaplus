package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (druid *Druid) ApplyTalents() {
	druid.applyServerTalents()
	// Balance
	druid.registerMoonkinFormSpell()
	druid.applyOmenOfClarity()
	druid.applyImprovedMoonfire()
	druid.applyVengeance()
	druid.applyNaturesGrace()
	druid.applyMoonglow()
	druid.applyMoonfury()

	// DBC: Natural Weapons +1%/rank to all damage.
	druid.PseudoStats.DamageDealtMultiplier *= 1 + 0.01*float64(druid.Talents.NaturalWeapons)
	// Unity with Nature: -5%/rank damage taken from Arcane and Nature (the Barkskin cooldown part is in barkskin.go).
	if druid.Talents.UnityWithNature > 0 {
		reduction := 1 - 0.05*float64(druid.Talents.UnityWithNature)
		druid.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] *= reduction
		druid.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] *= reduction
	}
	druid.applyBalanceExtras()
	druid.applyRestoExtras()

	// Feral
	druid.applyBloodFrenzy()

	druid.ApplyEquipScaling(stats.Armor, druid.ThickHideMultiplier())

	if druid.Talents.HeartOfTheWild > 0 {
		bonus := 0.06 * float64(druid.Talents.HeartOfTheWild) // DBC: 6%/rank Int and Spirit
		druid.MultiplyStat(stats.Intellect, 1.0+bonus)
		druid.MultiplyStat(stats.Spirit, 1.0+bonus)
	}

	// Restoration
	druid.applyFuror()

	druid.PseudoStats.SpiritRegenRateCasting += .10 * float64(druid.Talents.Reflection) // DBC: 10%/rank
}

func (druid *Druid) ThickHideMultiplier() float64 {
	thickHideMulti := 1.0

	if druid.Talents.ThickHide > 0 {
		thickHideMulti += 0.05 * float64(druid.Talents.ThickHide) // DBC: 5%/rank
	}

	return thickHideMulti
}

func (druid *Druid) BearArmorMultiplier() float64 {
	sotfMulti := 1.0 + 0.33/3.0
	return 4.7 * sotfMulti
}

func (druid *Druid) applyNaturesGrace() {
	if !druid.Talents.NaturesGrace {
		return
	}

	affectedSpells := []*DruidSpell{}
	druid.NaturesGraceProcAura = druid.RegisterAura(core.Aura{
		Label:     "Natures Grace Proc",
		ActionID:  core.ActionID{SpellID: 16886},
		Duration:  time.Second * 15,
		MaxStacks: 1,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(druid.DruidSpells, func(ds *DruidSpell) bool {
				return ds.DefaultCast.CastTime > 0
			})
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				// DBC (16886): next spell has -50% cast time and -50% mana cost.
				spell.CastTimeMultiplier -= 0.5
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 50
				}
				if spell.SpellCode == SpellCode_DruidWrath {
					spell.DefaultCast.GCD -= time.Millisecond * 500
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.CastTimeMultiplier += 0.5
				if spell.Cost != nil {
					spell.Cost.Multiplier += 50
				}
				if spell.SpellCode == SpellCode_DruidWrath {
					spell.DefaultCast.GCD += time.Millisecond * 500
				}
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}

			// Make sure the aura actually applied to the spell being cast before deactivating
			if spell.CurCast.CastTime > 0 && (sim.CurrentTime-spell.CurCast.CastTime >= aura.StartedAt()) {
				aura.Deactivate(sim)
			}
		},
	})

	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Natures Grace",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Spells with travel times have their own implementation because the proc occurs as the cast finishes
			if spell.MissileSpeed == 0 && spell.ProcMask.Matches(core.ProcMaskSpellDamage) && result.DidCrit() {
				druid.NaturesGraceProcAura.Activate(sim)
				druid.NaturesGraceProcAura.SetStacks(sim, druid.NaturesGraceProcAura.MaxStacks)
			}
		},
	}))
}

// func (druid *Druid) registerNaturesSwiftnessCD() {
// 	if !druid.Talents.NaturesSwiftness {
// 		return
// 	}
// 	actionID := core.ActionID{SpellID: 17116}

// 	var nsAura *core.Aura
// 	nsSpell := druid.RegisterSpell(Humanoid|Moonkin|Tree, core.SpellConfig{
// 		ActionID: actionID,
// 		Flags:    core.SpellFlagNoOnCastComplete,
// 		Cast: core.CastConfig{
// 			CD: core.Cooldown{
// 				Timer:    druid.NewTimer(),
// 				Duration: time.Minute * 3,
// 			},
// 		},
// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
// 			nsAura.Activate(sim)
// 		},
// 	})

// 	nsAura = druid.RegisterAura(core.Aura{
// 		Label:    "Natures Swiftness",
// 		ActionID: actionID,
// 		Duration: core.NeverExpires,
// 		OnGain: func(aura *core.Aura, sim *core.Simulation) {
// 			if druid.Starfire != nil {
// 				druid.Starfire.CastTimeMultiplier -= 1
// 			}
// 			if druid.Wrath != nil {
// 				druid.Wrath.CastTimeMultiplier -= 1
// 			}
// 		},
// 		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
// 			if druid.Starfire != nil {
// 				druid.Starfire.CastTimeMultiplier += 1
// 			}
// 			if druid.Wrath != nil {
// 				druid.Wrath.CastTimeMultiplier += 1
// 			}
// 		},
// 		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
// 			if !druid.Wrath.IsEqual(spell) && !druid.Starfire.IsEqual(spell) {
// 				return
// 			}

// 			// Remove the buff and put skill on CD
// 			aura.Deactivate(sim)
// 			nsSpell.CD.Use(sim)
// 			druid.UpdateMajorCooldowns()
// 		},
// 	})

// 	druid.AddMajorCooldown(core.MajorCooldown{
// 		Spell: nsSpell.Spell,
// 		Type:  core.CooldownTypeDPS,
// 		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
// 			// Don't use NS unless we're casting a full-length starfire or wrath.
// 			return !character.HasTemporarySpellCastSpeedIncrease()
// 		},
// 	})
// }

// TODO: Classic bear
// func (druid *Druid) applyPrimalFury() {
// 	if druid.Talents.PrimalFury == 0 {
// 		return
// 	}

// 	procChance := []float64{0, 0.5, 1}[druid.Talents.PrimalFury]
// 	actionID := core.ActionID{SpellID: 37117}
// 	rageMetrics := druid.NewRageMetrics(actionID)
// 	cpMetrics := druid.NewComboPointMetrics(actionID)

// 	druid.RegisterAura(core.Aura{
// 		Label:    "Primal Fury",
// 		Duration: core.NeverExpires,
// 		OnReset: func(aura *core.Aura, sim *core.Simulation) {
// 			aura.Activate(sim)
// 		},
// 		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
// 			if druid.InForm(Bear) {
// 				if result.Outcome.Matches(core.OutcomeCrit) {
// 					if sim.Proc(procChance, "Primal Fury") {
// 						druid.AddRage(sim, 5, rageMetrics)
// 					}
// 				}
// 			} else if druid.InForm(Cat) {
// 				if druid.IsMangle(spell) || druid.Shred.IsEqual(spell) || druid.Rake.IsEqual(spell) {
// 					if result.Outcome.Matches(core.OutcomeCrit) {
// 						if sim.Proc(procChance, "Primal Fury") {
// 							druid.AddComboPoints(sim, 1, cpMetrics)
// 						}
// 					}
// 				}
// 			}
// 		},
// 	})
// }

func (druid *Druid) applyBloodFrenzy() {
	if druid.Talents.BloodFrenzy == 0 {
		return
	}

	procChance := []float64{0, 0.5, 1}[druid.Talents.BloodFrenzy]
	actionID := core.ActionID{SpellID: 16953}
	cpMetrics := druid.NewComboPointMetrics(actionID)

	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Blood Frenzy",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if druid.InForm(Cat) &&
				result.Target == aura.Unit.CurrentTarget &&
				spell.Flags.Matches(SpellFlagBuilder) &&
				result.Outcome.Matches(core.OutcomeCrit) &&
				sim.Proc(procChance, "Blood Frenzy") {
				druid.AddComboPoints(sim, 1, result.Target, cpMetrics)
			}
		},
	}))
}

// We're using an aura so that the APL can know if the Druid has furor for powershifting logic
func (druid *Druid) applyFuror() {
	if druid.Talents.Furor == 0 {
		return
	}

	spellID := []int32{0, 17056, 17058, 17059, 17060, 17061}[druid.Talents.Furor]

	druid.FurorAura = druid.RegisterAura(core.Aura{
		Label:    "Furor",
		ActionID: core.ActionID{SpellID: spellID},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
}

func (druid *Druid) applyOmenOfClarity() {
	if !druid.Talents.OmenOfClarity {
		return
	}

	var affectedSpells []*core.Spell
	druid.ClearcastingAura = druid.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 16870},
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(druid.Spellbook, func(spell *core.Spell) bool { return spell.Flags.Matches(SpellFlagOmen) })
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.Cost.Multiplier -= 75
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.Cost.Multiplier += 75
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}

			if spell.Flags.Matches(SpellFlagOmen) && spell.DefaultCast.Cost > 0 {
				aura.Deactivate(sim)
			}
		},
	})

	ppmm := druid.AutoAttacks.NewPPMManager(2.0, core.ProcMaskMelee)
	icd := core.Cooldown{
		Timer:    druid.NewTimer(),
		Duration: time.Second * 10,
	}

	druid.RegisterAura(core.Aura{
		Label:    "Omen of Clarity",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !icd.IsReady(sim) {
				return
			}
			// TODO: Phase 3 "and non-instant spell casts" but we need to find out how the procs work for those
			if spell.ProcMask.Matches(core.ProcMaskMelee) && ppmm.ProcWithWeaponSpecials(sim, spell.ProcMask, "Omen of Clarity") {
				icd.Use(sim)
				druid.ClearcastingAura.Activate(sim)
			}
		},
	})
}

func (druid *Druid) applyMoonfury() {
	if druid.Talents.Moonfury == 0 {
		return
	}

	multiplier := 0.03 * float64(druid.Talents.Moonfury) // DBC: 3%/rank

	druid.RegisterAura(core.Aura{
		Label: "Moonfury",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells := core.FilterSlice(
				core.Flatten(
					[][]*DruidSpell{
						druid.Starfire,
						druid.Moonfire,
					},
				),
				func(spell *DruidSpell) bool { return spell != nil },
			)
			for _, spell := range affectedSpells {
				spell.BaseDamageMultiplierAdditive += multiplier
			}
		},
	})
}

func (druid *Druid) applyImprovedMoonfire() {
	if druid.Talents.ImprovedMoonfire == 0 {
		return
	}

	bonusCrit := 10 * float64(druid.Talents.ImprovedMoonfire) * core.SpellCritRatingPerCritChance // DBC: 10%/rank

	druid.RegisterAura(core.Aura{
		Label: "Improved moonfire",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			critAffectedSpells := core.FilterSlice(
				druid.Moonfire,
				func(spell *DruidSpell) bool { return spell != nil },
			)

			for _, spell := range critAffectedSpells {
				spell.BonusCritRating += bonusCrit
			}
		},
	})
}

func (druid *Druid) applyVengeance() {
	if druid.Talents.Vengeance == 0 {
		return
	}

	critDamageBonus := 0.20 * float64(druid.Talents.Vengeance)

	druid.RegisterAura(core.Aura{
		Label: "Vengeance",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells := core.FilterSlice(
				core.Flatten(
					[][]*DruidSpell{
						druid.Wrath,
						druid.Starfire,
						druid.Moonfire,
					},
				),
				func(spell *DruidSpell) bool { return spell != nil },
			)

			for _, spell := range affectedSpells {
				spell.CritDamageBonus += critDamageBonus
			}
		},
	})
}

func (druid *Druid) applyMoonglow() {
	if druid.Talents.Moonglow == 0 {
		return
	}

	multiplier := 5 * druid.Talents.Moonglow // DBC: 5%/rank

	druid.RegisterAura(core.Aura{
		Label: "Moonglow",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells := core.FilterSlice(
				core.Flatten(
					[][]*DruidSpell{
						druid.Wrath,
						druid.Starfire,
						druid.Moonfire,
					},
				),
				func(spell *DruidSpell) bool { return spell != nil },
			)

			for _, spell := range affectedSpells {
				spell.Cost.Multiplier -= multiplier
			}
		},
	})
}

// Balance talents not otherwise modeled (DBC values).
func (druid *Druid) applyBalanceExtras() {
	// Omnipresence: -2/-4% resist chance for Balance spells.
	if druid.Talents.Omnipresence > 0 {
		hit := 2 * float64(druid.Talents.Omnipresence) * core.SpellHitRatingPerHitChance
		druid.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_DruidWrath, SpellCode_DruidStarfire, SpellCode_DruidMoonfire, SpellCode_DruidInsectSwarm:
				spell.BonusHitRating += hit
			}
		})
	}

	// Nature Balancer: Wrath has a 5%/rank chance to give the next Moonfire/Starfire +50% crit,
	// and Moonfire/Starfire have a 5%/rank chance to give the next Wrath +50% crit.
	if druid.Talents.NatureBalancer > 0 {
		procChance := 0.05 * float64(druid.Talents.NatureBalancer)
		critBonus := 50.0 * core.SpellCritRatingPerCritChance
		var wrath, arcane []*core.Spell
		druid.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_DruidWrath:
				wrath = append(wrath, spell)
			case SpellCode_DruidStarfire, SpellCode_DruidMoonfire:
				arcane = append(arcane, spell)
			}
		})
		makeAura := func(label string, id int32, targets func() []*core.Spell) *core.Aura {
			return druid.RegisterAura(core.Aura{
				Label:    label,
				ActionID: core.ActionID{SpellID: id},
				Duration: time.Second * 15,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					for _, s := range targets() {
						s.BonusCritRating += critBonus
					}
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					for _, s := range targets() {
						s.BonusCritRating -= critBonus
					}
				},
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
					if aura.RemainingDuration(sim) == aura.Duration {
						return
					}
					for _, s := range targets() {
						if s == spell {
							aura.Deactivate(sim)
							return
						}
					}
				},
			})
		}
		arcaneAura := makeAura("Nature Balancer: Arcane", 33756, func() []*core.Spell { return arcane })
		wrathAura := makeAura("Nature Balancer: Nature", 33757, func() []*core.Spell { return wrath })
		core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Nature Balancer",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				switch spell.SpellCode {
				case SpellCode_DruidWrath:
					if sim.Proc(procChance, "Nature Balancer") {
						arcaneAura.Activate(sim)
					}
				case SpellCode_DruidStarfire, SpellCode_DruidMoonfire:
					if sim.Proc(procChance, "Nature Balancer") {
						wrathAura.Activate(sim)
					}
				}
			},
		}))
	}
}

// Restoration/Feral talents that are pure stats or damage modifiers (DBC values).
func (druid *Druid) applyRestoExtras() {
	// Accuracy: +1/2/3% hit with all attacks and spells.
	if druid.Talents.Accuracy > 0 {
		druid.AddStat(stats.MeleeHit, float64(druid.Talents.Accuracy)*core.MeleeHitRatingPerHitChance)
		druid.AddStat(stats.SpellHit, float64(druid.Talents.Accuracy)*core.SpellHitRatingPerHitChance)
	}
	// Animism: spell damage and healing up to 10%/rank of Spirit.
	if druid.Talents.Animism > 0 {
		druid.AddStatDependency(stats.Spirit, stats.SpellPower, 0.10*float64(druid.Talents.Animism))
	}
	// Gift of Nature: +2%/rank Nature damage.
	if druid.Talents.GiftOfNature > 0 {
		druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexNature] *= 1 + 0.02*float64(druid.Talents.GiftOfNature)
	}
	// Dreamstate: regenerates 1%/rank of total mana every 10 seconds.
	if druid.Talents.Dreamstate > 0 {
		pct := 0.01 * float64(druid.Talents.Dreamstate)
		manaMetrics := druid.NewManaMetrics(core.ActionID{SpellID: 33835})
		druid.RegisterResetEffect(func(sim *core.Simulation) {
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period: time.Second * 10,
				OnAction: func(sim *core.Simulation) {
					druid.AddMana(sim, pct*druid.MaxMana(), manaMetrics)
				},
			})
		})
	}
	// Stalking: +20%/rank crit for Shred (and Ravage).
	if druid.Talents.Stalking > 0 {
		crit := 20 * float64(druid.Talents.Stalking) * core.CritRatingPerCritChance
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidShred {
				spell.BonusCritRating += crit
			}
		})
	}
	// Killer Instincts: +1%/rank all damage.
	if druid.Talents.KillerInstincts > 0 {
		druid.PseudoStats.DamageDealtMultiplier *= 1 + 0.01*float64(druid.Talents.KillerInstincts)
	}
}
