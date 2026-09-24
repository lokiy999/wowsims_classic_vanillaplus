package warlock

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warlock *Warlock) ApplyTalents() {
	warlock.applyServerTalents()
	warlock.applyWeaponImbue()

	// Affliction
	warlock.applySuppression()
	warlock.applyNightfall()
	warlock.applyShadowMastery()

	// Demonology
	warlock.applyDemonicEmbrace()
	warlock.applyFelIntellect()
	warlock.registerFelDominationCD()
	warlock.applyFelStamina()
	warlock.applyMasterSummoner()
	warlock.applyMasterDemonologist()
	warlock.applyDemonicSacrifice()
	warlock.applySoulLink()

	// Destruction
	warlock.applyImprovedShadowBolt()
	warlock.applyCataclysm()
	warlock.applyBane()
	warlock.applyDevastation()
	warlock.applyRuin()
	warlock.applyEmberstorm()
	warlock.applyShadowstorm()
	warlock.applyDestructionCrit()
	warlock.applyIntensity()
	warlock.applySadism()
	warlock.registerMayhemCD()
}

func (warlock *Warlock) applyWeaponImbue() {
	if warlock.GetCharacter().Equipment.OffHand().Type != proto.ItemType_ItemTypeUnknown {
		return
	}

	level := warlock.Level
	if warlock.Options.WeaponImbue == proto.WarlockOptions_Firestone {
		warlock.applyFirestone()
	}
	if warlock.Options.WeaponImbue == proto.WarlockOptions_Spellstone {
		if level >= 55 {
			warlock.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
		}
	}
}

func (warlock *Warlock) applyFirestone() {
	level := warlock.Level

	damageMin := 0.0
	damageMax := 0.0

	// TODO: Test for spell scaling
	spellCoeff := 0.0
	spellId := int32(0)

	// TODO: Test PPM
	ppm := warlock.AutoAttacks.NewPPMManager(8, core.ProcMaskMelee)

	firestoneMulti := 1.0 + float64(int32(0) /*removed*/)*0.15

	if level >= 56 {
		warlock.AddStat(stats.FirePower, 21*firestoneMulti)
		damageMin = 80.0
		damageMax = 120.0
		spellId = 17949
	} else if level >= 46 {
		warlock.AddStat(stats.FirePower, 17*firestoneMulti)
		damageMin = 60.0
		damageMax = 90.0
		spellId = 17947
	} else if level >= 36 {
		warlock.AddStat(stats.FirePower, 14*firestoneMulti)
		damageMin = 40.0
		damageMax = 60.0
		spellId = 17945
	} else if level >= 28 {
		warlock.AddStat(stats.FirePower, 10*firestoneMulti)
		damageMin = 25.0
		damageMax = 35.0
		spellId = 758
	}

	if level >= 28 && warlock.Consumes.MainHandImbue == proto.WeaponImbue_WeaponImbueUnknown {
		fireProcSpell := warlock.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: spellId},
			SpellSchool: core.SpellSchoolFire,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,

			DamageMultiplier:         firestoneMulti,
			ThreatMultiplier:         1,
			DamageMultiplierAdditive: 1,
			BonusCoefficient:         spellCoeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(damageMin, damageMax)

				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicCrit)
			},
		})

		core.MakePermanent(warlock.GetOrRegisterAura(core.Aura{
			Label: "Firestone Proc",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}

				if !spell.ProcMask.Matches(core.ProcMaskMelee) {
					return
				}

				if !ppm.Proc(sim, core.ProcMaskMelee, "Firestone Proc") {
					return
				}

				fireProcSpell.Cast(sim, result.Target)
			},
		}))
	}
}

///////////////////////////////////////////////////////////////////////////
//                            Affliction
///////////////////////////////////////////////////////////////////////////

func (warlock *Warlock) applySuppression() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	points := float64(warlock.Talents.Suppression)
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(WarlockFlagAffliction) {
			spell.BonusHitRating += 2 * points * core.SpellHitRatingPerHitChance
		}
	})
}

func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall <= 0 {
		return
	}

	shadowTranceAura := warlock.RegisterAura(core.Aura{
		Label:    "Nightfall Shadow Trance",
		ActionID: core.ActionID{SpellID: 17941},
		Duration: time.Second * 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.ShadowBolt {
				spell.CastTimeMultiplier -= 1
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.ShadowBolt {
				spell.CastTimeMultiplier += 1
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// Check if the shadowbolt was instant cast and not a normal one
			if spell.SpellCode == SpellCode_WarlockShadowBolt && spell.CurCast.CastTime == 0 {
				aura.Deactivate(sim)
			}
		},
	})

	// DBC: 1/2/3% chance per periodic Shadow damage tick, at most once per 10s.
	procChance := 0.01 * float64(warlock.Talents.Nightfall)
	var nextNightfall time.Duration

	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Nightfall Hidden Aura",
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolShadow) || sim.CurrentTime < nextNightfall {
				return
			}
			if sim.Proc(procChance, "Nightfall") {
				nextNightfall = sim.CurrentTime + time.Second*10
				shadowTranceAura.Activate(sim)
			}
		},
	}))
}

func (warlock *Warlock) applyShadowMastery() {
	if warlock.Talents.ShadowMastery == 0 {
		return
	}

	// These spells have their base damage modded instead
	// Apply Aura: Modifies Spell Effectiveness (8)
	excludedSpellCodes := []int32{SpellCode_WarlockCurseOfAgony, SpellCode_WarlockDeathCoil, SpellCode_WarlockDrainLife, SpellCode_WarlockDrainSoul}

	warlock.OnSpellRegistered(func(spell *core.Spell) {
		// Shadow Mastery applies a base damage modifier to all dots / channeled spells instead
		if spell.SpellSchool.Matches(core.SpellSchoolShadow) && isWarlockSpell(spell) && !slices.Contains(excludedSpellCodes, spell.SpellCode) {
			spell.DamageMultiplierAdditive += warlock.shadowMasteryBonus()
		}
	})
}

func (warlock *Warlock) shadowMasteryBonus() float64 {
	return .02 * float64(warlock.Talents.ShadowMastery)
}

///////////////////////////////////////////////////////////////////////////
//                            Demonology Talents
///////////////////////////////////////////////////////////////////////////

func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	// DBC: +10%/rank Demon Armor/Skin effectiveness (applied in applyDemonArmor);
	// the health regen and Shadow Ward reflect parts are not modeled.
}

func (warlock *Warlock) applyFelIntellect() {
	if warlock.Talents.FelIntellect == 0 {
		return
	}

	// DBC: +5%/rank demon mana and +5%/rank total Intellect.
	multiplier := 1 + 0.05*float64(warlock.Talents.FelIntellect)
	warlock.MultiplyStat(stats.Intellect, multiplier)
	for _, pet := range warlock.BasePets {
		pet.MultiplyStat(stats.Mana, multiplier)
	}
}

func (warlock *Warlock) applyFelStamina() {
	if warlock.Talents.FelStamina == 0 {
		return
	}

	// DBC: +5%/rank demon health and +5%/rank total Stamina.
	multiplier := 1 + 0.05*float64(warlock.Talents.FelStamina)
	warlock.MultiplyStat(stats.Stamina, multiplier)
	for _, pet := range warlock.BasePets {
		pet.MultiplyStat(stats.Health, multiplier)
	}
}

func (warlock *Warlock) applyMasterSummoner() {
	if warlock.Talents.MasterSummoner == 0 {
		return
	}

	// DBC: -3s cast time and -30% mana cost per rank.
	castTimeReduction := time.Second * 3 * time.Duration(warlock.Talents.MasterSummoner)
	costReduction := 30 * warlock.Talents.MasterSummoner

	// Use an aura because the summon spells aren't registered by this point
	warlock.RegisterAura(core.Aura{
		Label:    "Master Summoner Hidden Aura",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SummonDemonSpells {
				spell.DefaultCast.CastTime -= castTimeReduction
				spell.Cost.Multiplier -= costReduction
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SummonDemonSpells {
				spell.DefaultCast.CastTime += castTimeReduction
				spell.Cost.Multiplier += costReduction
			}
		},
	})
}

func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 {
		return
	}

	points := float64(warlock.Talents.MasterDemonologist)
	// DBC: Imp +3%/rank spell crit, Voidwalker -3%/rank physical damage taken,
	// Succubus +3%/rank all damage, Felhunter +0.2 resistance per level (per rank).
	damageDealtMultiplier := 1 + 0.03*points
	damageTakenMultiplier := 1 - 0.03*points
	critBonus := 3 * points * core.SpellCritRatingPerCritChance
	bonusResistance := 2 * points

	impConfig := core.Aura{
		Label:    "Master Demonologist (Imp)",
		ActionID: core.ActionID{SpellID: 23825, Tag: 1},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.SpellCrit, critBonus)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.SpellCrit, -critBonus)
		},
	}

	voidwalkerConfig := core.Aura{
		Label:    "Master Demonologist (Voidwalker)",
		ActionID: core.ActionID{SpellID: 23825, Tag: 2},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier *= damageTakenMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= damageTakenMultiplier
		},
	}

	succubusConfig := core.Aura{
		Label:    "Master Demonologist (Succubus)",
		ActionID: core.ActionID{SpellID: 23825, Tag: 3},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageDealtMultiplier *= damageDealtMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageDealtMultiplier /= damageDealtMultiplier
		},
	}

	felhunterConfig := core.Aura{
		Label:    "Master Demonologist (Felhunter)",
		ActionID: core.ActionID{SpellID: 23825, Tag: 4},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddResistancesDynamic(sim, bonusResistance)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddResistancesDynamic(sim, -bonusResistance)
		},
	}

	for _, pet := range warlock.BasePets {
		pet.ApplyOnPetEnable(func(sim *core.Simulation) {
			if warlock.MasterDemonologistAura != nil {
				warlock.MasterDemonologistAura.Deactivate(sim)
			}
		})
	}

	warlockImpAura := warlock.RegisterAura(impConfig)
	impAura := warlock.Imp.RegisterAura(impConfig)
	warlock.Imp.ApplyOnPetEnable(func(sim *core.Simulation) {
		impAura.Activate(sim)
		warlock.MasterDemonologistAura = warlockImpAura
	})
	warlock.Imp.ApplyOnPetDisable(func(sim *core.Simulation, isSacrifice bool) {
		impAura.Deactivate(sim)
	})

	warlockVoidwalkerAura := warlock.RegisterAura(voidwalkerConfig)
	voidwalkerAura := warlock.Voidwalker.RegisterAura(voidwalkerConfig)
	warlock.Voidwalker.ApplyOnPetEnable(func(sim *core.Simulation) {
		voidwalkerAura.Activate(sim)
		warlock.MasterDemonologistAura = warlockVoidwalkerAura
	})
	warlock.Voidwalker.ApplyOnPetDisable(func(sim *core.Simulation, isSacrifice bool) {
		voidwalkerAura.Deactivate(sim)
	})

	warlockSuccubusAura := warlock.RegisterAura(succubusConfig)
	succubusAura := warlock.Succubus.RegisterAura(succubusConfig)
	warlock.Succubus.ApplyOnPetEnable(func(sim *core.Simulation) {
		succubusAura.Activate(sim)
		warlock.MasterDemonologistAura = warlockSuccubusAura
	})
	warlock.Succubus.ApplyOnPetDisable(func(sim *core.Simulation, isSacrifice bool) {
		succubusAura.Deactivate(sim)
	})

	warlockFelhunterAura := warlock.RegisterAura(felhunterConfig)
	felhunterAura := warlock.Felhunter.RegisterAura(felhunterConfig)
	warlock.Felhunter.ApplyOnPetEnable(func(sim *core.Simulation) {
		felhunterAura.Activate(sim)
		warlock.MasterDemonologistAura = warlockFelhunterAura
	})
	warlock.Felhunter.ApplyOnPetDisable(func(sim *core.Simulation, isSacrifice bool) {
		felhunterAura.Deactivate(sim)
	})
}

func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink {
		return
	}

	actionID := core.ActionID{SpellID: 19028}
	soulLinkConfig := core.Aura{
		Label:    "Soul Link Aura",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= 1.3
			aura.Unit.PseudoStats.DamageDealtMultiplier *= 1.05
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageDealtMultiplier /= 1.05
			aura.Unit.PseudoStats.DamageTakenMultiplier *= 1.3
		},
	}

	warlock.SoulLinkAura = warlock.RegisterAura(soulLinkConfig)
	for _, pet := range warlock.BasePets {
		pet.SoulLinkAura = pet.RegisterAura(soulLinkConfig)

		oldOnPetDisable := pet.OnPetDisable
		pet.OnPetDisable = func(sim *core.Simulation, isSacrifice bool) {
			oldOnPetDisable(sim, isSacrifice)
			warlock.SoulLinkAura.Deactivate(sim)
			pet.SoulLinkAura.Deactivate(sim)
		}
	}

	warlock.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.ActivePet != nil
		},

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.2,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warlock.SoulLinkAura.Activate(sim)
			warlock.ActivePet.SoulLinkAura.Activate(sim)
		},
	})
}

func (warlock *Warlock) applyDemonicSacrifice() {
	if !warlock.Talents.DemonicSacrifice {
		return
	}

	impAura := warlock.GetOrRegisterAura(core.Aura{
		Label:    "Burning Wish",
		ActionID: core.ActionID{SpellID: 18789},
		Duration: 30 * time.Minute,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			// DBC (18789): -30% threat only; the +15% Fire damage was retail.
			warlock.PseudoStats.ThreatMultiplier *= 0.70
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warlock.PseudoStats.ThreatMultiplier /= 0.70
		},
	})

	var vwPa *core.PendingAction
	healthMetric := warlock.NewHealthMetrics(core.ActionID{SpellID: 18790})
	voidwalkerAura := warlock.GetOrRegisterAura(core.Aura{
		Label:    "Fel Stamina",
		ActionID: core.ActionID{SpellID: 18790},
		Duration: 30 * time.Minute,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			vwPa = core.NewPeriodicAction(sim, core.PeriodicActionOptions{
				Period: time.Second * 3,
				OnAction: func(s *core.Simulation) {
					warlock.GainHealth(sim, warlock.MaxHealth()*0.03, healthMetric)
				},
			})
			sim.AddPendingAction(vwPa)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			vwPa.Cancel(sim)
		},
	})

	succubusAura := warlock.GetOrRegisterAura(core.Aura{
		Label:    "Touch of Shadow",
		ActionID: core.ActionID{SpellID: 18791},
		Duration: 30 * time.Minute,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			// DBC (18791): +5% damage.
			warlock.PseudoStats.DamageDealtMultiplier *= 1.05
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warlock.PseudoStats.DamageDealtMultiplier /= 1.05
		},
	})

	var fhPa *core.PendingAction
	manaMetric := warlock.NewManaMetrics(core.ActionID{SpellID: 18792})
	felhunterAura := warlock.GetOrRegisterAura(core.Aura{
		Label:    "Fel Energy",
		ActionID: core.ActionID{SpellID: 18792},
		Duration: 30 * time.Minute,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			fhPa = core.NewPeriodicAction(sim, core.PeriodicActionOptions{
				Period: time.Second * 3,
				OnAction: func(s *core.Simulation) {
					warlock.AddMana(sim, warlock.MaxMana()*0.03, manaMetric)
				},
			})
			sim.AddPendingAction(fhPa)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			fhPa.Cancel(sim)
		},
	})

	dsAuras := []*core.Aura{felhunterAura, impAura, succubusAura, voidwalkerAura}
	for _, pet := range warlock.BasePets {
		oldOnPetEnable := pet.OnPetEnable
		pet.OnPetEnable = func(sim *core.Simulation) {
			oldOnPetEnable(sim)
			for _, dsAura := range dsAuras {
				dsAura.Deactivate(sim)
			}
		}
	}

	warlock.GetOrRegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_WarlockDemonicSacrifice,
		ActionID:    core.ActionID{SpellID: 18788},
		SpellSchool: core.SpellSchoolShadow,
		Flags:       core.SpellFlagAPL,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.ActivePet != nil
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			switch warlock.ActivePet {
			case warlock.Felhunter:
				felhunterAura.Activate(sim)
			case warlock.Imp:
				impAura.Activate(sim)
			case warlock.Succubus:
				succubusAura.Activate(sim)
			case warlock.Voidwalker:
				voidwalkerAura.Activate(sim)
			}

			warlock.changeActivePet(sim, nil, true)
		},
	})
}

///////////////////////////////////////////////////////////////////////////
//                            Destruction Talents
///////////////////////////////////////////////////////////////////////////

func (warlock *Warlock) applyImprovedShadowBolt() {
	if warlock.Talents.ImprovedShadowBolt == 0 {
		return
	}

	warlock.ImprovedShadowBoltAuras = warlock.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
		return core.ImprovedShadowBoltAura(unit, warlock.Talents.ImprovedShadowBolt)
	})

	affectedSpellCodes := []int32{SpellCode_WarlockShadowBolt}
	// DBC: 20%/rank chance on Shadow Bolt crit.
	procChance := 0.20 * float64(warlock.Talents.ImprovedShadowBolt)
	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "ISB Trigger",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.ShadowBolt {
				spell.RelatedAuras = []core.AuraArray{warlock.ImprovedShadowBoltAuras}
			}
			warlock.DebuffSpells = append(warlock.DebuffSpells, warlock.ShadowBolt...)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && result.DidCrit() && slices.Contains(affectedSpellCodes, spell.SpellCode) && sim.Proc(procChance, "ISB") {
				isbAura := warlock.ImprovedShadowBoltAuras.Get(result.Target)
				isbAura.Activate(sim)
				isbAura.SetStacks(sim, isbAura.MaxStacks)
			}
		},
	}))
}

func (warlock *Warlock) applyCataclysm() {
	if warlock.Talents.Cataclysm == 0 {
		return
	}

	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(WarlockFlagDestruction) && spell.Cost != nil {
			spell.Cost.Multiplier -= 5 * warlock.Talents.Cataclysm // DBC: 5%/rank
		}
	})
}

func (warlock *Warlock) applyBane() {
	if warlock.Talents.Bane == 0 {
		return
	}

	points := time.Duration(warlock.Talents.Bane)
	// DBC: Shadow Bolt -0.1s/rank, Soul Fire -0.5s/rank (Immolate is not affected).
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_WarlockShadowBolt {
			spell.DefaultCast.CastTime -= time.Millisecond * 100 * points
		} else if spell.SpellCode == SpellCode_WarlockSoulFire {
			spell.DefaultCast.CastTime -= time.Millisecond * 500 * points
		}
	})
}

func (warlock *Warlock) applyDevastation() {
	if warlock.Talents.Devastation == 0 {
		return
	}

	points := float64(warlock.Talents.Devastation)
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(WarlockFlagDestruction) {
			spell.BonusCritRating += points * core.CritRatingPerCritChance
		}
	})
}

func (warlock *Warlock) improvedImmolateBonus() float64 {
	return 0.05 * core.TernaryFloat64(warlock.Talents.ImprovedImmolate, 1, 0)
}

func (warlock *Warlock) applyRuin() {
	if warlock.Talents.Ruin == 0 {
		return
	}
	ruinBonus := 0.20 * float64(warlock.Talents.Ruin) // DBC: 20%/rank
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(WarlockFlagDestruction) {
			spell.CritDamageBonus += ruinBonus
		}
	})
}

func (warlock *Warlock) applyEmberstorm() {
	if warlock.Talents.Emberstorm == 0 {
		return
	}

	points := float64(warlock.Talents.Emberstorm)
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellSchool.Matches(core.SpellSchoolFire) && isWarlockSpell(spell) {
			spell.DamageMultiplierAdditive += 0.03 * points // DBC: 3%/rank
		}
	})
}

// Shadowstorm (DBC): +3%/rank Shadow damage.
func (warlock *Warlock) applyShadowstorm() {
	if warlock.Talents.Shadowstorm == 0 {
		return
	}
	points := float64(warlock.Talents.Shadowstorm)
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellSchool.Matches(core.SpellSchoolShadow) && isWarlockSpell(spell) {
			spell.DamageMultiplierAdditive += 0.03 * points
		}
	})
}

// Bring the Pain (DBC): +5%/rank crit for Searing Pain, Conflagrate, Shadowburn and Soul Fire.
func (warlock *Warlock) applyDestructionCrit() {
	if warlock.Talents.BringThePain == 0 {
		return
	}
	bonusCrit := 5 * float64(warlock.Talents.BringThePain) * core.SpellCritRatingPerCritChance
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_WarlockSearingPain || spell.SpellCode == SpellCode_WarlockConflagrate ||
			spell.SpellCode == SpellCode_WarlockShadowburn || spell.SpellCode == SpellCode_WarlockSoulFire {
			spell.BonusCritRating += bonusCrit
		}
	})
}

// Intensity (DBC): -2/-4% target resist chance for Destruction spells.
func (warlock *Warlock) applyIntensity() {
	if warlock.Talents.Intensity == 0 {
		return
	}
	bonusHit := 2 * float64(warlock.Talents.Intensity) * core.SpellHitRatingPerHitChance
	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(WarlockFlagDestruction) {
			spell.BonusHitRating += bonusHit
		}
	})
}

// Sadism (Vanilla+ calculator): spell crits have an 8%/rank chance to restore 666 mana.
func (warlock *Warlock) applySadism() {
	if warlock.Talents.Sadism == 0 {
		return
	}
	procChance := 0.08 * float64(warlock.Talents.Sadism)
	manaMetrics := warlock.NewManaMetrics(core.ActionID{SpellID: 33977})
	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Sadism",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
				return
			}
			if sim.Proc(procChance, "Sadism") {
				warlock.AddMana(sim, 666, manaMetrics)
			}
		},
	}))
}
