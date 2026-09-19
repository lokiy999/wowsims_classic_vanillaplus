package warrior

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warrior *Warrior) ToughnessArmorMultiplier() float64 {
	return 1.0 + 0.03*float64(warrior.Talents.Toughness) // DBC: 3%/rank
}

func (warrior *Warrior) ApplyTalents() {
	warrior.AddStat(stats.MeleeCrit, core.CritRatingPerCritChance*1*float64(warrior.Talents.Cruelty))
	warrior.ApplyEquipScaling(stats.Armor, warrior.ToughnessArmorMultiplier())
	// DBC: Anticipation reduces the chance to be critically hit by 1%/rank.
	warrior.PseudoStats.ReducedCritTakenChance += 0.01 * float64(warrior.Talents.Anticipation)
	warrior.AddStat(stats.MeleeHit, core.MeleeHitRatingPerHitChance*float64(warrior.Talents.Precision)) // DBC: Precision 1%/rank
	warrior.MultiplyStat(stats.Health, 1+0.02*float64(warrior.Talents.Vitality))                        // DBC: Vitality 2%/rank
	warrior.applyArmsExtras()
	warrior.applyWeaponExpertiseTypes()
	warrior.AddStat(stats.Parry, []float64{0, 3, 5}[warrior.Talents.Deflection]) // DBC: 3/5%

	warrior.applyAngerManagement()
	warrior.applyDeepWounds()
	warrior.applyOneHandedWeaponSpecialization()
	warrior.applyTwoHandedWeaponSpecialization()
	warrior.applyWeaponExpertise()
	warrior.applyUnbridledWrath()
	warrior.applyEnrage()
	warrior.applyFlurry()
	warrior.applyShieldSpecialization()
	warrior.registerDeathWishCD()
	warrior.registerSweepingStrikesCD()
	warrior.registerLastStandCD()
}

func (warrior *Warrior) applyAngerManagement() {
	if !warrior.Talents.AngerManagement {
		return
	}

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12296})

	warrior.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: time.Second, // DBC: 1 rage per second in combat
			OnAction: func(sim *core.Simulation) {
				warrior.AddRage(sim, 1, rageMetrics)
				warrior.LastAMTick = sim.CurrentTime
			},
		})
	})
}

func (warrior *Warrior) applyTwoHandedWeaponSpecialization() {
	if warrior.Talents.TwoHandedWeaponSpecialization == 0 || warrior.MainHand().HandType != proto.HandType_HandTypeTwoHand {
		return
	}

	multiplier := 1 + 0.02*float64(warrior.Talents.TwoHandedWeaponSpecialization) // DBC: 2%/rank
	warrior.OnSpellRegistered(func(spell *core.Spell) {
		if spell.BonusCoefficient > 0 {
			spell.DamageMultiplier *= multiplier
		}
	})
}

func (warrior *Warrior) applyOneHandedWeaponSpecialization() {
	if warrior.Talents.OneHandedWeaponSpecialization == 0 || warrior.MainHand().HandType == proto.HandType_HandTypeTwoHand {
		return
	}

	multiplier := 1 + 0.02*float64(warrior.Talents.OneHandedWeaponSpecialization)
	warrior.OnSpellRegistered(func(spell *core.Spell) {
		if spell.BonusCoefficient > 0 {
			spell.DamageMultiplier *= multiplier
		}
	})
}

// Weapon Expertise (custom tree): 1% chance per rank to get an extra attack on
// the same target after dealing melee damage. Replaces the old per-weapon-type
// specialization talents.
func (warrior *Warrior) applyWeaponExpertise() {
	if warrior.Talents.WeaponExpertise == 0 {
		return
	}

	icd := core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: time.Millisecond * 200,
	}
	procChance := 0.01 * float64(warrior.Talents.WeaponExpertise)

	warrior.RegisterAura(core.Aura{
		Label:    "Weapon Expertise",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMelee) {
				return
			}
			if !icd.IsReady(sim) {
				return
			}
			if sim.RandomFloat("Weapon Expertise") < procChance {
				icd.Use(sim)
				warrior.AutoAttacks.ExtraMHAttack(sim, 1, core.ActionID{SpellID: 12815}, spell.ActionID)
			}
		},
	})
}

func (warrior *Warrior) applyUnbridledWrath() {
	if warrior.Talents.UnbridledWrath == 0 {
		return
	}

	procChance := 0.05 * float64(warrior.Talents.UnbridledWrath) // DBC: 5%/rank, 5 rage (spell 12964)

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12964})

	warrior.RegisterAura(core.Aura{
		Label:    "Unbridled Wrath",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}

			if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && sim.RandomFloat("Unbrided Wrath") < procChance {
				warrior.AddRage(sim, 5, rageMetrics)
			}
		},
	})
}

func (warrior *Warrior) applyEnrage() {
	if warrior.Talents.Enrage == 0 {
		return
	}

	// Confirmed by user: each critical hit taken adds a stack of +1%/rank melee damage,
	// up to 10 stacks, lasting 15s. Stacks are not consumed by swings.
	perStack := 0.01 * float64(warrior.Talents.Enrage)
	warrior.EnrageAura = warrior.GetOrRegisterAura(core.Aura{
		Label:     "Enrage",
		ActionID:  core.ActionID{SpellID: 13048},
		Duration:  time.Second * 15,
		MaxStacks: 10,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= 1 + perStack*float64(oldStacks)
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1 + perStack*float64(newStacks)
		},
	})

	warrior.RegisterAura(core.Aura{
		Label:    "Enrage Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskMelee) || !result.Outcome.Matches(core.OutcomeCrit) {
				return
			}
			warrior.EnrageAura.Activate(sim)
			warrior.EnrageAura.AddStack(sim)
		},
	})
}

// func (warrior *Warrior) applyFlurry() {
// 	if warrior.Talents.Flurry == 0 {
// 		return
// 	}

// 	haste := []float64{1, 1.1, 1.15, 1.2, 1.25, 1.3}[warrior.Talents.Flurry]

// 	procAura := warrior.RegisterAura(core.Aura{
// 		Label:     "Flurry Proc",
// 		ActionID:  core.ActionID{SpellID: 12974},
// 		Duration:  core.NeverExpires,
// 		MaxStacks: 3,
// 		OnGain: func(aura *core.Aura, sim *core.Simulation) {
// 			warrior.MultiplyMeleeSpeed(sim, haste)
// 		},
// 		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
// 			warrior.MultiplyMeleeSpeed(sim, 1/haste)
// 		},
// 	})

// 	warrior.RegisterAura(core.Aura{
// 		Label:    "Flurry",
// 		Duration: core.NeverExpires,
// 		OnReset: func(aura *core.Aura, sim *core.Simulation) {
// 			aura.Activate(sim)
// 		},
// 		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
// 			if !spell.ProcMask.Matches(core.ProcMaskMelee) {
// 				return
// 			}

// 			if result.Outcome.Matches(core.OutcomeCrit) {
// 				procAura.Activate(sim)
// 				procAura.SetStacks(sim, 3)
// 				return
// 			}

// 			// Remove a stack.
// 			if procAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
// 				procAura.RemoveStack(sim)
// 			}
// 		},
// 	})
// }

func (warrior *Warrior) applyFlurry() {
	if warrior.Talents.Flurry == 0 {
		return
	}

	talentAura := warrior.makeFlurryAura(warrior.Talents.Flurry)

	// This must be registered before the below trigger because in-game a crit weapon swing consumes a stack before the refresh, so you end up with:
	// 3 => 2
	// refresh
	// 2 => 3
	warrior.makeFlurryConsumptionTrigger(talentAura)

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Flurry Proc Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Outcome.Matches(core.OutcomeCrit) {
				talentAura.Activate(sim)
				if talentAura.IsActive() {
					talentAura.SetStacks(sim, 3)
				}
				return
			}
		},
	}))
}

// These are separated out because of the T1 Shaman Tank 2P that can proc Flurry separately from the talent.
// It triggers the max-rank Flurry aura but with dodge, parry, or block.
func (warrior *Warrior) makeFlurryAura(points int32) *core.Aura {
	if points == 0 {
		return nil
	}

	spellID := []int32{12319, 12971, 12972, 12973, 12974}[points-1]
	attackSpeed := []float64{1.05, 1.10, 1.15, 1.20, 1.25}[points-1] // DBC: 5%/rank

	aura := warrior.GetOrRegisterAura(core.Aura{
		Label:     fmt.Sprintf("Flurry Proc (%d)", spellID),
		ActionID:  core.ActionID{SpellID: spellID},
		Duration:  core.NeverExpires,
		MaxStacks: 3,
	})

	aura.NewExclusiveEffect("Flurry", true, core.ExclusiveEffect{
		Priority: attackSpeed,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			warrior.MultiplyMeleeSpeed(sim, attackSpeed)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			warrior.MultiplyMeleeSpeed(sim, 1/attackSpeed)
		},
	})

	return aura
}

// With the Protection T2 4pc it's possible to have 2 different Flurry auras if using less than 5/5 points in Flurry.
// The two different buffs don't stack whatsoever. Instead the stronger aura takes precedence and each one is only refreshed by the corresponding triggers.
func (warrior *Warrior) makeFlurryConsumptionTrigger(flurryAura *core.Aura) *core.Aura {
	icd := core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: time.Millisecond * 500,
	}
	return core.MakePermanent(warrior.GetOrRegisterAura(core.Aura{
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

func (warrior *Warrior) applyShieldSpecialization() {
	if warrior.Talents.ShieldSpecialization == 0 {
		return
	}

	warrior.AddStat(stats.Block, core.BlockRatingPerBlockChance*1*float64(warrior.Talents.ShieldSpecialization))

	procChance := 0.2 * float64(warrior.Talents.ShieldSpecialization)
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 12727})

	warrior.RegisterAura(core.Aura{
		Label:    "Shield Specialization",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidBlock() {
				if sim.Proc(procChance, "Shield Specialization") {
					warrior.AddRage(sim, 1.0, rageMetrics)
				}
			}
		},
	})
}

func (warrior *Warrior) registerDeathWishCD() {
	if !warrior.Talents.DeathWish {
		return
	}

	actionID := core.ActionID{SpellID: 12328}

	deathWishAura := warrior.RegisterAura(core.Aura{
		Label:    "Death Wish",
		ActionID: actionID,
		Duration: time.Second * 30,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.2
			warrior.PseudoStats.ArmorMultiplier *= 0.5 // DBC: -50% armor
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= 1.2
			warrior.PseudoStats.ArmorMultiplier /= 0.5
		},
	})
	core.RegisterPercentDamageModifierEffect(deathWishAura, 1.2)

	warrior.DeathWish = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagHelpful,
		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			IgnoreHaste: true,
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			deathWishAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: warrior.DeathWish.Spell,
		Type:  core.CooldownTypeDPS,
	})
}

func (warrior *Warrior) registerLastStandCD() {
	if !warrior.Talents.LastStand {
		return
	}

	actionID := core.ActionID{SpellID: 12975}
	healthMetrics := warrior.NewHealthMetrics(actionID)

	var bonusHealth float64
	lastStandAura := warrior.RegisterAura(core.Aura{
		Label:    "Last Stand",
		ActionID: actionID,
		Duration: time.Second * 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			bonusHealth = warrior.MaxHealth() * 0.3
			warrior.AddStatsDynamic(sim, stats.Stats{stats.Health: bonusHealth})
			warrior.GainHealth(sim, bonusHealth, healthMetrics)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.AddStatsDynamic(sim, stats.Stats{stats.Health: -bonusHealth})
		},
	})

	lastStandSpell := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Minute * 10,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			lastStandAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: lastStandSpell.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}

func (warrior *Warrior) impale() float64 {
	return 0.1 * float64(warrior.Talents.Impale)
}

// Arms/Prot talents with DBC values that were missing: Dog of War (-4%/rank ability cost),
// Para Bellum (-4%/rank ability cooldowns), Training and Discipline (-1 rage/rank).
func (warrior *Warrior) applyArmsExtras() {
	if warrior.Talents.DogOfWar > 0 || warrior.Talents.TrainingAndDiscipline > 0 {
		pct := 4 * warrior.Talents.DogOfWar
		flat := int32(warrior.Talents.TrainingAndDiscipline)
		warrior.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Cost != nil && spell.Cost.CostType() == core.CostTypeRage && spell.Cost.BaseCost > 0 {
				spell.Cost.Multiplier -= pct
				spell.Cost.FlatModifier -= flat
			}
		})
	}
	if warrior.Talents.ParaBellum > 0 {
		factor := 1 - 0.04*float64(warrior.Talents.ParaBellum)
		warrior.OnSpellRegistered(func(spell *core.Spell) {
			if spell.CD.Timer != nil && spell.CD.Duration > 0 && spell.SpellCode != SpellCode_WarriorNone {
				spell.CD.Duration = time.Duration(float64(spell.CD.Duration) * factor)
			}
		})
	}
// Slamcraft: +10%/rank Shield Slam crit (Slam cast time reduction is in slam.go).
	if warrior.Talents.Slamcraft > 0 {
		crit := 10 * float64(warrior.Talents.Slamcraft) * core.CritRatingPerCritChance
		warrior.OnSpellRegistered(func(spell *core.Spell) {
if spell.SpellCode == SpellCode_WarriorShieldSlam {
				spell.BonusCritRating += crit
			}
		})
	}
}

// Weapon Expertise (Vanilla+ calculator): +1%/rank crit with Axes and Polearms; Maces ignore
// 2 armor per level per rank. (The Sword extra attack is in applyWeaponExpertise.)
// Improved Mortal Strike: -0.5s cooldown and +5% damage per rank.
func (warrior *Warrior) applyWeaponExpertiseTypes() {
	if rank := float64(warrior.Talents.WeaponExpertise); rank > 0 {
		crit := core.CritRatingPerCritChance * rank
		switch warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypePolearm) {
		case core.ProcMaskMelee:
			warrior.AddStat(stats.MeleeCrit, crit)
		case core.ProcMaskMeleeMH:
			warrior.AddStat(stats.MeleeCrit, crit)
			warrior.OnSpellRegistered(func(spell *core.Spell) {
				if spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
					spell.BonusCritRating -= crit
				}
			})
		case core.ProcMaskMeleeOH:
			warrior.OnSpellRegistered(func(spell *core.Spell) {
				if spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
					spell.BonusCritRating += crit
				}
			})
		}
		if warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeMace) != core.ProcMaskUnknown {
			warrior.AddStat(stats.ArmorPenetration, 2*float64(warrior.Level)*rank)
		}
	}
	if warrior.Talents.ImprovedMortalStrike > 0 {
		points := float64(warrior.Talents.ImprovedMortalStrike)
		warrior.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_WarriorMortalStrike {
				spell.CD.Duration -= time.Millisecond * 500 * time.Duration(warrior.Talents.ImprovedMortalStrike)
				spell.DamageMultiplier *= 1 + 0.05*points
			}
		})
	}
}
