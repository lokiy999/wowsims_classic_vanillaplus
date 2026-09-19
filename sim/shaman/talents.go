package shaman

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (shaman *Shaman) ApplyTalents() {
	// Elemental Talents
	shaman.applyConcussion()
	shaman.applyCallOfThunder()
	shaman.applyLightningMastery()
	shaman.applyElementalPrecision()
	shaman.applyElementalFocus()
	shaman.applyElementalDevastation()
	shaman.applyElementalFury()
	shaman.registerElementalMasteryCD()

	// Enhancement Talents
	shaman.applyFlurry()

	if shaman.Talents.AncestralKnowledge > 0 {
		// DBC: 2/3/4/5% total mana.
		shaman.MultiplyStat(stats.Mana, 1.0+[]float64{0, .02, .03, .04, .05}[shaman.Talents.AncestralKnowledge])
	}

	shaman.AddStat(stats.Block, 1*float64(int32(0) /*removed*/))

	shaman.AddStat(stats.MeleeCrit, core.CritRatingPerCritChance*1*float64(shaman.Talents.ThunderingStrikes))

	shaman.applyShamanExtras()

	shaman.ApplyEquipScaling(stats.Armor, 1+.05*float64(shaman.Talents.Toughness)) // DBC: 5%/rank

	// Parry talent removed from custom shaman tree

	// TODO: Check whether this does what it should.
	// From all I've seen this appears to not actually be a school modifier at all, but instead simply applies
	// to all attacks done with a weapon. The weaponmask seems to take precedence and the school mask is actually ignored.
	// Will also be the case for similar talents like the one for retribution.
	shaman.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1 + (.02 * float64(shaman.Talents.WeaponMastery))

	// Restoration Talents
	// TODO: Healing Way
	// TODO: Ancestral Healing
	shaman.registerNaturesSwiftnessCD()
	// shaman.registerManaTideTotemCD()

	if shaman.Talents.TidalFocus > 0 {
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShaman) && spell.ProcMask.Matches(core.ProcMaskSpellHealing) && spell.Cost != nil {
				spell.Cost.Multiplier -= shaman.Talents.TidalFocus
			}
		})
	}

	shaman.AddStat(stats.MeleeHit, float64(shaman.Talents.NaturesGuidance))
	shaman.AddStat(stats.SpellHit, float64(shaman.Talents.NaturesGuidance))

	if int32(0) /*removed*/ > 0 {
		threatMultiplier := 1 - .05*float64(int32(0) /*removed*/)
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShaman) && spell.ProcMask.Matches(core.ProcMaskSpellHealing) {
				spell.ThreatMultiplier *= threatMultiplier
			}
		})
	}

	if shaman.Talents.TidalMastery > 0 {
		critBonus := float64(shaman.Talents.TidalMastery) * core.CritRatingPerCritChance
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagShaman) && (spell.ProcMask.Matches(core.ProcMaskSpellHealing) ||
				spell.Flags.Matches(SpellFlagLightning)) {
				spell.BonusCritRating += critBonus
			}
		})
	}
}

func (shaman *Shaman) applyConcussion() {
	if shaman.Talents.Concussion == 0 {
		return
	}

	additiveMultiplier := 0.01 * float64(shaman.Talents.Concussion)
	affectedSpellCodes := []int32{SpellCode_ShamanLightningBolt, SpellCode_ShamanChainLightning, SpellCode_ShamanEarthShock, SpellCode_ShamanFlameShock, SpellCode_ShamanFrostShock}

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if slices.Contains(affectedSpellCodes, spell.SpellCode) {
			spell.DamageMultiplierAdditive += additiveMultiplier
		}
	})
}

// Call of Thunder (DBC): +2%/rank crit for Lightning Bolt and Chain Lightning.
func (shaman *Shaman) applyCallOfThunder() {
	if shaman.Talents.CallOfThunder == 0 {
		return
	}
	bonusCrit := 2 * float64(shaman.Talents.CallOfThunder) * core.SpellCritRatingPerCritChance
	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_ShamanLightningBolt || spell.SpellCode == SpellCode_ShamanChainLightning {
			spell.BonusCritRating += bonusCrit
		}
	})
}

// Lightning Mastery (DBC): -0.2s/rank cast time for Lightning Bolt and Chain Lightning.
func (shaman *Shaman) applyLightningMastery() {
	if shaman.Talents.LightningMastery == 0 {
		return
	}
	reduction := time.Millisecond * 200 * time.Duration(shaman.Talents.LightningMastery)
	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_ShamanLightningBolt || spell.SpellCode == SpellCode_ShamanChainLightning {
			spell.DefaultCast.CastTime -= reduction
		}
	})
}

// Elemental Precision (DBC): +5%/rank spell hit for Fire/Frost/Nature.
func (shaman *Shaman) applyElementalPrecision() {
	if shaman.Talents.ElementalPrecision == 0 {
		return
	}
	bonusHit := 5 * float64(shaman.Talents.ElementalPrecision) * core.SpellHitRatingPerHitChance
	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagShaman) && (spell.SpellSchool.Matches(core.SpellSchoolFire) ||
			spell.SpellSchool.Matches(core.SpellSchoolFrost) || spell.SpellSchool.Matches(core.SpellSchoolNature)) {
			spell.BonusHitRating += bonusHit
		}
	})
}

func (shaman *Shaman) callOfFlameMultiplier() float64 {
	return 1 + .10*float64(shaman.Talents.CallOfFlame) // DBC: 10%/rank
}

func (shaman *Shaman) applyElementalFocus() {
	if shaman.Talents.ElementalFocus == 0 {
		return
	}

	var affectedSpells []*core.Spell

	shaman.ClearcastingAura = shaman.RegisterAura(core.Aura{
		Label:     "Clearcasting",
		ActionID:  core.ActionID{SpellID: 16246},
		Duration:  time.Second * 15,
		MaxStacks: 1,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = shaman.getClearcastingSpells()
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 100
				}
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				if spell.Cost != nil {
					spell.Cost.Multiplier += 100
				}
			})
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			if newStacks == 0 {
				aura.Deactivate(sim)
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}

			if aura.GetStacks() > 0 && shaman.isShamanDamagingSpell(spell) {
				aura.RemoveStack(sim)
			}
		},
	})

	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label: "Elemental Focus Trigger",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if shaman.isShamanDamagingSpell(spell) && sim.Proc(0.03*float64(shaman.Talents.ElementalFocus), "Elemental Focus") {
				shaman.ClearcastingAura.Activate(sim)
				shaman.ClearcastingAura.SetStacks(sim, shaman.ClearcastingAura.MaxStacks)
			}
		},
	}))
}

func (shaman *Shaman) isShamanDamagingSpell(spell *core.Spell) bool {
	return spell.Flags.Matches(SpellFlagShaman) && spell.ProcMask.Matches(core.ProcMaskSpellDamage)
}

func (shaman *Shaman) getClearcastingSpells() []*core.Spell {
	return core.FilterSlice(
		shaman.Spellbook,
		func(spell *core.Spell) bool {
			return spell != nil && shaman.isShamanDamagingSpell(spell)
		},
	)
}

func (shaman *Shaman) applyElementalDevastation() {
	if shaman.Talents.ElementalDevastation == 0 {
		return
	}

	spellID := []int32{0, 30165, 29177, 29178}[shaman.Talents.ElementalDevastation]
	critBonus := 10.0 * core.CritRatingPerCritChance // Vanilla+ calculator: 10% crit for 5/10/15s by rank
	procAura := shaman.NewTemporaryStatsAura("Elemental Devastation Proc", core.ActionID{SpellID: spellID}, stats.Stats{stats.MeleeCrit: critBonus, stats.SpellCrit: critBonus}, time.Second*5*time.Duration(shaman.Talents.ElementalDevastation))

	shaman.RegisterAura(core.Aura{
		Label:    "Elemental Devastation",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskSpellDamage) && result.Outcome.Matches(core.OutcomeCrit) {
				procAura.Activate(sim)
			}
		},
	})
}

func (shaman *Shaman) applyElementalFury() {
	if shaman.Talents.ElementalFury == 0 {
		return
	}

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if (spell.Flags.Matches(SpellFlagShaman) || spell.Flags.Matches(SpellFlagTotem)) && spell.DefenseType == core.DefenseTypeMagic {
			spell.CritDamageBonus += 0.2 * float64(shaman.Talents.ElementalFury) // DBC: +20%/rank
		}
	})
}

func (shaman *Shaman) registerElementalMasteryCD() {
	if !shaman.Talents.ElementalMastery {
		return
	}

	actionID := core.ActionID{SpellID: 16166}

	cdTimer := shaman.NewTimer()
	cd := time.Minute * 3

	var affectedSpells []*core.Spell

	emAura := shaman.RegisterAura(core.Aura{
		Label:    "Elemental Mastery",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(
				shaman.Spellbook,
				func(spell *core.Spell) bool { return spell != nil && shaman.isShamanDamagingSpell(spell) },
			)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.BonusCritRating += core.CritRatingPerCritChance * 100
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 100
				}
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.BonusCritRating -= core.CritRatingPerCritChance * 100
				if spell.Cost != nil {
					spell.Cost.Multiplier += 100
				}
			})
			shaman.ElementalMastery.CD.Use(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if shaman.isShamanDamagingSpell(spell) {
				// Elemental mastery can be batched
				core.StartDelayedAction(sim, core.DelayedActionOptions{
					DoAt: sim.CurrentTime + core.SpellBatchWindow,
					OnAction: func(sim *core.Simulation) {
						if aura.IsActive() {
							// Remove the buff and put skill on CD
							aura.Deactivate(sim)
							cdTimer.Set(sim.CurrentTime + cd)
							shaman.UpdateMajorCooldowns()
						}
					},
				})
			}
		},
	})

	shaman.ElementalMastery = shaman.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: cd,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			emAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: shaman.ElementalMastery,
		Type:  core.CooldownTypeDPS,
	})
}

func (shaman *Shaman) registerNaturesSwiftnessCD() {
	if !shaman.Talents.NaturesSwiftness {
		return
	}
	actionID := core.ActionID{SpellID: 16188}
	cdTimer := shaman.NewTimer()
	cd := time.Minute * 3

	var affectedSpells []*core.Spell

	nsAura := shaman.RegisterAura(core.Aura{
		Label:    "Natures Swiftness",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(
				shaman.Spellbook,
				func(spell *core.Spell) bool {
					return spell != nil && spell.SpellSchool.Matches(core.SpellSchoolNature) && spell.DefaultCast.CastTime > 0
				},
			)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier -= 1 })
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier += 1 })
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolNature) && spell.DefaultCast.CastTime > 0 {
				// Remove the buff and put skill on CD
				aura.Deactivate(sim)
				cdTimer.Set(sim.CurrentTime + cd)
				shaman.UpdateMajorCooldowns()
			}
		},
	})

	nsSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: cd,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// Don't use NS unless we're casting a full-length lightning bolt, which is
			// the only spell shamans have with a cast longer than GCD.
			return !shaman.HasTemporarySpellCastSpeedIncrease()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			nsAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: nsSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (shaman *Shaman) applyFlurry() {
	if shaman.Talents.Flurry == 0 {
		return
	}

	talentAura := shaman.makeFlurryAura(shaman.Talents.Flurry)

	// This must be registered before the below trigger because in-game a crit weapon swing consumes a stack before the refresh, so you end up with:
	// 3 => 2
	// refresh
	// 2 => 3
	shaman.makeFlurryConsumptionTrigger(talentAura)

	shaman.RegisterAura(core.Aura{
		Label:    "Flurry Proc Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Outcome.Matches(core.OutcomeCrit) {
				talentAura.Activate(sim)
				if talentAura.IsActive() {
					talentAura.SetStacks(sim, 3)
				}
				return
			}
		},
	})
}

// These are separated out because of the T1 Shaman Tank 2P that can proc Flurry separately from the talent.
// It triggers the max-rank Flurry aura but with dodge, parry, or block.
func (shaman *Shaman) makeFlurryAura(points int32) *core.Aura {
	if points == 0 {
		return nil
	}

	spellID := []int32{16257, 16277, 16278, 16279, 16280}[points-1]
	attackSpeed := []float64{1.05, 1.10, 1.15, 1.20, 1.25}[points-1] // DBC: 5%/rank

	aura := shaman.GetOrRegisterAura(core.Aura{
		Label:     fmt.Sprintf("Flurry Proc (%d)", spellID),
		ActionID:  core.ActionID{SpellID: spellID},
		Duration:  core.NeverExpires,
		MaxStacks: 3,
	})

	aura.NewExclusiveEffect("Flurry", true, core.ExclusiveEffect{
		Priority: attackSpeed,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, attackSpeed)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, 1/(attackSpeed))
		},
	})

	return aura
}

// With the Warden T1 2pc it's possible to have 2 different Flurry auras if using less than 5/5 points in Flurry.
// The two different buffs don't stack whatsoever. Instead the stronger aura takes precedence and each one is only refreshed by the corresponding triggers.
func (shaman *Shaman) makeFlurryConsumptionTrigger(flurryAura *core.Aura) *core.Aura {
	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: time.Millisecond * 500,
	}
	return core.MakePermanent(shaman.GetOrRegisterAura(core.Aura{
		Label: fmt.Sprintf("Flurry Consume Trigger - %d", flurryAura.ActionID.SpellID),
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Remove a stack.
			if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && icd.IsReady(sim) {
				icd.Use(sim)
				flurryAura.RemoveStack(sim)
			}
		},
	}))
}

func (shaman *Shaman) totemManaMultiplier() int32 {
	return 100 - 25*shaman.Talents.TotemicFocus // DBC: 25%/rank
}

// Restorative Totems uses Mod Spell Effectiveness (Base Value)
func (shaman *Shaman) restorativeTotemsModifier() float64 {
	return []float64{0, .30, .50}[shaman.Talents.RestorativeTotems] // DBC: 30/50%
}

// Purification uses Mod Spell Effectiveness (Base Healing)
func (shaman *Shaman) purificationHealingModifier() float64 {
	return .02 * float64(shaman.Talents.Purification)
}

// func (shaman *Shaman) registerManaTideTotemCD() {
// 	if !shaman.Talents.ManaTideTotem {
// 		return
// 	}

// 	mttAura := core.ManaTideTotemAura(shaman.GetCharacter(), shaman.Index)
// 	mttSpell := shaman.RegisterSpell(core.SpellConfig{
// 		ActionID: core.ManaTideTotemActionID,
// 		Flags:    core.SpellFlagNoOnCastComplete,
// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: time.Second,
// 			},
// 			IgnoreHaste: true,
// 			CD: core.Cooldown{
// 				Timer:    shaman.NewTimer(),
// 				Duration: time.Minute * 5,
// 			},
// 		},
// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
// 			mttAura.Activate(sim)

// 			// If healing stream is active, cancel it while mana tide is up.
// 			if shaman.HealingStreamTotem.Hot(&shaman.Unit).IsActive() {
// 				for _, agent := range shaman.Party.Players {
// 					shaman.HealingStreamTotem.Hot(&agent.GetCharacter().Unit).Cancel(sim)
// 				}
// 			}

// 			// TODO: Current water totem buff needs to be removed from party/raid.
// 			if shaman.Totems.Water != proto.WaterTotem_NoWaterTotem {
// 				shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + time.Second*12
// 			}
// 		},
// 	})

// 	shaman.AddMajorCooldown(core.MajorCooldown{
// 		Spell: mttSpell,
// 		Type:  core.CooldownTypeDPS,
// 		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
// 			return sim.CurrentTime > time.Second*30
// 		},
// 	})
// }

// Talents with DBC values not otherwise handled above.
func (shaman *Shaman) applyShamanExtras() {
	shaman.applyStaticField()
	if shaman.Talents.Stormforged {
		manaMetrics := shaman.NewManaMetrics(core.ActionID{SpellID: 34100})
		core.MakePermanent(shaman.RegisterAura(core.Aura{
			Label: "Stormforged Mana",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.SpellCode == SpellCode_ShamanStormstrike && result.Landed() {
					shaman.AddMana(sim, 0.05*shaman.MaxMana(), manaMetrics)
				}
			},
		}))
	}
	// Anticipation: -3%/rank shock resist chance.
	if shaman.Talents.Anticipation > 0 {
		hit := 3 * float64(shaman.Talents.Anticipation) * core.SpellHitRatingPerHitChance
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_ShamanEarthShock, SpellCode_ShamanFlameShock, SpellCode_ShamanFrostShock:
				spell.BonusHitRating += hit
			}
		})
	}

	// Thundering Strikes also adds 1%/rank crit to Shock spells.
	if shaman.Talents.ThunderingStrikes > 0 {
		crit := float64(shaman.Talents.ThunderingStrikes) * core.SpellCritRatingPerCritChance
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_ShamanEarthShock, SpellCode_ShamanFlameShock, SpellCode_ShamanFrostShock:
				spell.BonusCritRating += crit
			}
		})
	}

	// Call of Flame: +10%/rank Flame Shock crit.
	if shaman.Talents.CallOfFlame > 0 {
		crit := 10 * float64(shaman.Talents.CallOfFlame) * core.SpellCritRatingPerCritChance
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_ShamanFlameShock {
				spell.BonusCritRating += crit
			}
		})
	}

	// Tidal Focus: -1%/rank mana on lightning spells, -10%/rank on Frost Shock.
	if shaman.Talents.TidalFocus > 0 {
		shaman.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Cost == nil {
				return
			}
			switch spell.SpellCode {
			case SpellCode_ShamanLightningBolt, SpellCode_ShamanChainLightning:
				spell.Cost.Multiplier -= shaman.Talents.TidalFocus
			case SpellCode_ShamanFrostShock:
				spell.Cost.Multiplier -= 10 * shaman.Talents.TidalFocus
			}
		})
	}

	// Lightning Overlord: Lightning Bolt/Chain Lightning crits refund 10%/rank of base mana cost.
	if shaman.Talents.LightningOverlord > 0 {
		refund := 0.10 * float64(shaman.Talents.LightningOverlord)
		manaMetrics := shaman.NewManaMetrics(core.ActionID{SpellID: 33012})
		core.MakePermanent(shaman.RegisterAura(core.Aura{
			Label: "Lightning Overlord",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.DidCrit() || spell.Cost == nil {
					return
				}
				if spell.SpellCode == SpellCode_ShamanLightningBolt || spell.SpellCode == SpellCode_ShamanChainLightning {
					shaman.AddMana(sim, spell.Cost.BaseCost*refund, manaMetrics)
				}
			},
		}))
	}

	// Nature's Spirit: spell damage and healing up to 8%/rank of Spirit.
	if shaman.Talents.NatureSpirit > 0 {
		shaman.AddStatDependency(stats.Spirit, stats.SpellPower, 0.08*float64(shaman.Talents.NatureSpirit))
	}

	// Stormforged: spell damage and healing equal to 20% of Attack Power.
	if shaman.Talents.Stormforged {
		shaman.AddStatDependency(stats.AttackPower, stats.SpellPower, 0.20)
	}

	// Elemental Warding: -5%/rank Fire, Frost and Nature damage taken.
	if shaman.Talents.ElementalWarding > 0 {
		mult := 1 - 0.05*float64(shaman.Talents.ElementalWarding)
		shaman.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] *= mult
		shaman.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] *= mult
		shaman.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] *= mult
	}
}

// Static Field (Vanilla+ calculator): damage spells have a 20%/rank chance to add a stack
// (max 10, 20s) that gives +1% damage and -1% mana cost to Lightning Bolt and Chain Lightning.
func (shaman *Shaman) applyStaticField() {
	if shaman.Talents.StaticField == 0 {
		return
	}
	procChance := 0.20 * float64(shaman.Talents.StaticField)
	var affected []*core.Spell
	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_ShamanLightningBolt || spell.SpellCode == SpellCode_ShamanChainLightning {
			affected = append(affected, spell)
		}
	})
	aura := shaman.RegisterAura(core.Aura{
		Label:     "Static Field",
		ActionID:  core.ActionID{SpellID: 33610},
		Duration:  time.Second * 20,
		MaxStacks: 10,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			delta := newStacks - oldStacks
			for _, spell := range affected {
				spell.DamageMultiplierAdditive += 0.01 * float64(delta)
				if spell.Cost != nil {
					spell.Cost.Multiplier -= delta
				}
			}
		},
	})
	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label: "Static Field Trigger",
		OnSpellHitDealt: func(aura2 *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.Flags.Matches(SpellFlagShaman) || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}
			if sim.Proc(procChance, "Static Field") {
				aura.Activate(sim)
				aura.AddStack(sim)
			}
		},
	}))
}
