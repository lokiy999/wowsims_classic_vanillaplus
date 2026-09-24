package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Bear abilities with the server's values (CSV's/Spell.csv). Values marked "unverified" are not in the server data.

// Bear paw "weapon": the same 54.8 base DPS as the cat claws, at 2.5 speed.
func (druid *Druid) GetBearWeapon() core.Weapon {
	return core.Weapon{
		BaseDamageMin:        109.6,
		BaseDamageMax:        164.4,
		SwingSpeed:           2.5,
		NormalizedSwingSpeed: 2.5,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
	}
}

// Bear Form (5487, passive 1178): armor from items +180%, health +10%, attack power +10%.
// Dire Bear Form (9634, passive 9635): armor from items +360%, health +20%, attack power +180.
// Both: threat +30% (21178). Leader of the Pack (17007) doubles the form effects.
func (druid *Druid) registerBearFormSpell() {
	dire := druid.Level >= 40
	actionID := core.ActionID{SpellID: core.TernaryInt32(dire, 9634, 5487)}
	healthMetrics := druid.NewHealthMetrics(actionID)

	lotp := core.TernaryFloat64(druid.Talents.LeaderOfThePack, 2, 1)
	armorMultiplier := 1 + lotp*core.TernaryFloat64(dire, 3.6, 1.8)
	healthMultiplier := 1 + lotp*core.TernaryFloat64(dire, 0.20, 0.10)

	statBonus := druid.GetFormShiftStats()
	var apDep *stats.StatDependency
	if dire {
		statBonus[stats.AttackPower] += lotp * 180
	} else {
		apDep = druid.NewDynamicMultiplyStat(stats.AttackPower, 1+lotp*0.10)
	}

	// Feral Instinct (16947-16949): Bear Form threat +5% per rank.
	threatMultiplier := 1.3 + 0.05*float64(druid.Talents.FeralInstinct)

	healthDep := druid.NewDynamicMultiplyStat(stats.Health, healthMultiplier)
	feralApDep := druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1)

	// Heart of the Wild (17003-17006, 24894): Stamina +3% per rank in Bear Form.
	var hotwDep *stats.StatDependency
	if druid.Talents.HeartOfTheWild > 0 {
		hotwDep = druid.NewDynamicMultiplyStat(stats.Stamina, 1.0+0.03*float64(druid.Talents.HeartOfTheWild))
	}

	pawWeapon := druid.GetBearWeapon()

	druid.BearFormAura = druid.RegisterAura(core.Aura{
		Label:      "Bear Form",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(druid.StartingForm.Matches(Bear), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if !druid.Env.MeasuringStats && druid.form != Humanoid {
				druid.CancelShapeshift(sim)
			}
			druid.form = Bear
			druid.SetCurrentPowerBar(core.RageBar)

			druid.AutoAttacks.SetMH(pawWeapon)
			druid.PseudoStats.ThreatMultiplier *= threatMultiplier
			druid.SetShapeshift(aura)

			druid.AddStatsDynamic(sim, statBonus)
			druid.ApplyDynamicEquipScaling(sim, stats.Armor, armorMultiplier)
			druid.EnableDynamicStatDep(sim, feralApDep)
			if apDep != nil {
				druid.EnableDynamicStatDep(sim, apDep)
			}

			// Keep the same fraction of health when shifting.
			healthFrac := druid.CurrentHealth() / druid.MaxHealth()
			druid.EnableDynamicStatDep(sim, healthDep)
			if hotwDep != nil {
				druid.EnableDynamicStatDep(sim, hotwDep)
			}
			if !druid.Env.MeasuringStats {
				druid.GainHealth(sim, healthFrac*druid.MaxHealth()-druid.CurrentHealth(), healthMetrics)
				druid.AutoAttacks.SetReplaceMHSwing(druid.ReplaceBearMHFunc)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.manageCooldownsEnabled()
				druid.UpdateManaRegenRates()
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.form = Humanoid
			druid.SetCurrentPowerBar(core.ManaBar)
			druid.AutoAttacks.SetMH(druid.WeaponFromMainHand())
			druid.PseudoStats.ThreatMultiplier /= threatMultiplier
			druid.SetShapeshift(nil)

			druid.AddStatsDynamic(sim, statBonus.Invert())
			druid.RemoveDynamicEquipScaling(sim, stats.Armor, armorMultiplier)
			druid.DisableDynamicStatDep(sim, feralApDep)
			if apDep != nil {
				druid.DisableDynamicStatDep(sim, apDep)
			}

			healthFrac := druid.CurrentHealth() / druid.MaxHealth()
			druid.DisableDynamicStatDep(sim, healthDep)
			if hotwDep != nil {
				druid.DisableDynamicStatDep(sim, hotwDep)
			}
			if !druid.Env.MeasuringStats {
				druid.RemoveHealth(sim, druid.CurrentHealth()-healthFrac*druid.MaxHealth())
				druid.AutoAttacks.SetReplaceMHSwing(nil)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.manageCooldownsEnabled()
				druid.UpdateManaRegenRates()
				if druid.EnrageAura != nil {
					druid.EnrageAura.Deactivate(sim)
				}
				if druid.MaulQueueAura != nil {
					druid.MaulQueueAura.Deactivate(sim)
				}
			}
		},
	})

	rageMetrics := druid.NewRageMetrics(actionID)
	furorProcChance := 0.2 * float64(druid.Talents.Furor)

	druid.BearForm = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost:   0.55,
			Multiplier: 100 - 10*druid.Talents.NaturalShapeshifter,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// Shifting sets rage to 0 (10 with a Furor proc).
			rageDelta := core.TernaryFloat64(sim.RandomFloat("Furor") < furorProcChance, 10, 0) - druid.CurrentRage()
			if rageDelta > 0 {
				druid.AddRage(sim, rageDelta, rageMetrics)
			} else if rageDelta < 0 {
				druid.SpendRage(sim, -rageDelta, rageMetrics)
			}
			druid.BearFormAura.Activate(sim)
		},
	})
}

// Maul (6807-9881): increases the next attack by 20-130 damage, 15 rage. Savage Fury +20% per rank, Ferocity -1
// rage per rank. The 1.75 threat multiplier is the classic value (unverified).
func (druid *Druid) registerMaulSpell() {
	ranks := []struct {
		level   int32
		spellID int32
		damage  float64
	}{
		{10, 6807, 20}, {18, 6808, 30}, {26, 6809, 40}, {34, 8972, 50}, {42, 9745, 70}, {50, 9880, 100}, {58, 9881, 130},
	}
	rank := ranks[0]
	for _, r := range ranks {
		if druid.Level >= r.level {
			rank = r
		}
	}

	druid.Maul = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: rank.spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeMHAuto,
		Flags:       SpellFlagOmen | core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		RageCost: core.RageCostOptions{
			Cost:   15 - float64(druid.Talents.Ferocity),
			Refund: 0.8,
		},

		DamageMultiplier: 1 + 0.2*float64(druid.Talents.SavageFury),
		ThreatMultiplier: 1.75,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := rank.damage + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
			druid.MaulQueueAura.Deactivate(sim)
		},
	})

	druid.MaulQueueAura = druid.RegisterAura(core.Aura{
		Label:    "Maul Queue Aura",
		ActionID: druid.Maul.ActionID,
		Duration: core.NeverExpires,
	})

	druid.MaulQueueSpell = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: druid.Maul.WithTag(1),
		Flags:    core.SpellFlagAPL | core.SpellFlagMeleeMetrics,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !druid.MaulQueueAura.IsActive() && druid.CurrentRage() >= druid.Maul.Cost.GetCurrentCost()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			druid.MaulQueueAura.Activate(sim)
		},
	})
}

// TryMaul replaces the next main hand swing with Maul when it is queued.
func (druid *Druid) TryMaul(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if druid.MaulQueueAura == nil || !druid.MaulQueueAura.IsActive() {
		return mhSwingSpell
	}
	if !druid.Maul.Spell.CanCast(sim, druid.CurrentTarget) {
		druid.MaulQueueAura.Deactivate(sim)
		return mhSwingSpell
	}
	return druid.Maul.Spell
}

// Demoralizing Roar (9898): -132 attack power for 30 sec, 10 rage. Feral Aggression +20% per rank (in core).
func (druid *Druid) registerDemoralizingRoarSpell() {
	druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.DemoralizingRoarAura(target, druid.Talents.FeralAggression)
	})

	druid.DemoralizingRoar = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 9898},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  42, // classic value, unverified

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.TargetUnits {
				result := spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
				if result.Landed() {
					druid.DemoralizingRoarAuras.Get(aoeTarget).Activate(sim)
				}
			}
		},

		RelatedAuras: []core.AuraArray{druid.DemoralizingRoarAuras},
	})
}

// Enrage (5229): 20 rage over 10 sec, 1 min cooldown, lowers armor while active. Improved Enrage (17080/17081) adds
// 10/20 rage at once; Primal Tenacity (33761-33763) lowers the cooldown by 10 sec per rank. The armor loss (27% in
// Bear Form, 16% in Dire Bear Form) is the classic value; the server data only has a script dummy for it.
func (druid *Druid) registerEnrageSpell() {
	actionID := core.ActionID{SpellID: 5229}
	rageMetrics := druid.NewRageMetrics(actionID)
	instantRage := []float64{0, 10, 20}[min(druid.Talents.ImprovedEnrage, 2)]
	armorLoss := core.TernaryFloat64(druid.Level >= 40, 0.16, 0.27)
	armorRemoved := 0.0

	druid.EnrageAura = druid.RegisterAura(core.Aura{
		Label:    "Enrage",
		ActionID: actionID,
		Duration: 10 * time.Second,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			armorRemoved = armorLoss * druid.GetStat(stats.Armor)
			druid.AddStatDynamic(sim, stats.Armor, -armorRemoved)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.AddStatDynamic(sim, stats.Armor, armorRemoved)
		},
	})

	druid.Enrage = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute - 10*time.Second*time.Duration(druid.Talents.PrimalTenacity),
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if instantRage > 0 {
				druid.AddRage(sim, instantRage, rageMetrics)
			}
			druid.EnrageAura.Activate(sim)
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				NumTicks: 10,
				Period:   time.Second,
				OnAction: func(sim *core.Simulation) {
					if druid.EnrageAura.IsActive() {
						druid.AddRage(sim, 2, rageMetrics)
					}
				},
			})
		},
	})
}

// Frenzied Regeneration (22842/22895/22896): converts up to 10 rage per second into 10/20/30 health per rage for
// 10 sec, 5 min cooldown.
func (druid *Druid) registerFrenziedRegenerationCD() {
	if druid.Level < 36 {
		return
	}
	spellID, healthPerRage := int32(22842), 10.0
	if druid.Level >= 56 {
		spellID, healthPerRage = 22896, 30
	} else if druid.Level >= 46 {
		spellID, healthPerRage = 22895, 20
	}
	actionID := core.ActionID{SpellID: spellID}
	healthMetrics := druid.NewHealthMetrics(actionID)
	rageMetrics := druid.NewRageMetrics(actionID)

	druid.FrenziedRegenerationAura = druid.RegisterAura(core.Aura{
		Label:    "Frenzied Regeneration",
		ActionID: actionID,
		Duration: 10 * time.Second,
	})

	druid.FrenziedRegeneration = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 5 * time.Minute,
			},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.FrenziedRegenerationAura.Activate(sim)
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				NumTicks: 10,
				Period:   time.Second,
				OnAction: func(sim *core.Simulation) {
					if !druid.FrenziedRegenerationAura.IsActive() {
						return
					}
					rage := min(druid.CurrentRage(), 10.0)
					if rage > 0 {
						druid.SpendRage(sim, rage, rageMetrics)
						druid.GainHealth(sim, rage*healthPerRage*druid.PseudoStats.HealingTakenMultiplier, healthMetrics)
					}
				},
			})
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.FrenziedRegeneration.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}

// Grizzly's Fury (35721, server spell, level 24): attack speed +20% for 10 sec, 10 rage, 10 sec cooldown. Feral
// Instinct makes it last 20% longer per rank. The "extra rage when taking damage" part has no value in the server
// data and is not modeled.
func (druid *Druid) registerGrizzlysFurySpell() {
	if druid.Level < 24 {
		return
	}
	actionID := core.ActionID{SpellID: 35721}

	druid.GrizzlysFuryAura = druid.RegisterAura(core.Aura{
		Label:    "Grizzly's Fury",
		ActionID: actionID,
		Duration: time.Duration(float64(10*time.Second) * (1 + 0.2*float64(druid.Talents.FeralInstinct))),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.MultiplyMeleeSpeed(sim, 1.2)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.MultiplyMeleeSpeed(sim, 1/1.2)
		},
	})

	druid.GrizzlysFury = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 10 * time.Second,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.GrizzlysFuryAura.Activate(sim)
		},
	})
}

// Primal Fury (16958/16961, effect 16959): 50% chance per rank to gain 5 rage on a critical strike in Bear Form.
func (druid *Druid) applyPrimalFury() {
	if druid.Talents.PrimalFury == 0 {
		return
	}
	procChance := 0.5 * float64(druid.Talents.PrimalFury)
	rageMetrics := druid.NewRageMetrics(core.ActionID{SpellID: 16959})

	core.MakePermanent(druid.RegisterAura(core.Aura{
		Label: "Primal Fury",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if druid.InForm(Bear) && result.DidCrit() && spell.ProcMask.Matches(core.ProcMaskMelee) &&
				sim.RandomFloat("Primal Fury") < procChance {
				druid.AddRage(sim, 5, rageMetrics)
			}
		},
	}))
}
