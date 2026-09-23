package item_effects

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Trinkets whose effect had no code (see docs/TODO.md, 2026-09-23). Effect text is the
// private server's (VPlusItemDB.lua); cooldowns are the classic ones from Wowhead because
// the server data has none.
const (
	ChainedEssenceOfEranikus         = 10455
	SmokeysLighter                   = 13171
	HeartOfTheScale                  = 13164
	RamsteinsLightningBolts          = 13515
	TheLionHornOfStormwind           = 14557
	RaggedJohnsNeverendingCup        = 15873
	RoyalSealOfEldreThalasAttributes = 18466
	RoyalSealOfEldreThalasCostA      = 18471
	RoyalSealOfEldreThalasCostB      = 18472
	BlazingEmblem                    = 2802
	AegisOfPreservation              = 19345
	MarlisEye                        = 19930
	GrileksCharmOfValor              = 19952
	WushoolaysCharmOfNature          = 19955
	HazzarahsCharmOfHealing          = 19958
	BlessedPrayerBeads               = 19990
	GraceOfEarth                     = 21181
	FetishOfChitinousSpikes          = 21488
	FetishOfTheSandReaver            = 21647
	PetrifiedScarab                  = 21685
	ShardOfTheFallenStar             = 21891
	ShendralarBadgeOfDeterrence      = 26094
	MarkOfTheVeteranCritA            = 26160
	MarkOfTheVeteranCritB            = 26163
	MarkOfTheVeteranSpellHitA        = 26165
	MarkOfTheVeteranCritC            = 26170
	MarkOfTheVeteranCritD            = 26172
	MarkOfTheVeteranSpellHitB        = 26174
	MarkOfThirst                     = 26203
	SawtoothTalisman                 = 26212
	TwoFacedMedallion                = 26353
	OnyxEgg                          = 83003
	BlueMottledEgg                   = 83006
	PinkSpeckledEgg                  = 83007
)

var allResistances = []stats.Stat{stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance, stats.NatureResistance, stats.ShadowResistance}

func resistances(amount float64) stats.Stats {
	s := stats.Stats{}
	for _, r := range allResistances {
		s[r] = amount
	}
	return s
}

// registerOnUseAura makes an on-use trinket that activates `aura`. Offensive ones share the
// offensive trinket cooldown for the aura's duration, like the other on-use trinkets.
func registerOnUseAura(character *core.Character, itemID int32, aura *core.Aura, cooldown time.Duration, cdType core.CooldownType) {
	cast := core.CastConfig{
		CD: core.Cooldown{
			Timer:    character.NewTimer(),
			Duration: cooldown,
		},
	}
	if cdType == core.CooldownTypeDPS {
		cast.SharedCD = core.Cooldown{
			Timer:    character.GetOffensiveTrinketCD(),
			Duration: aura.Duration,
		}
	} else {
		cast.SharedCD = core.Cooldown{
			Timer:    character.GetDefensiveTrinketCD(),
			Duration: aura.Duration,
		}
	}
	spell := character.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{ItemID: itemID},
		ProcMask: core.ProcMaskEmpty,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast:     cast,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})
	character.AddMajorCooldown(core.MajorCooldown{
		Type:  cdType,
		Spell: spell,
	})
}

// registerOnUseDamage makes an on-use trinket that hits every target (or only the main target) for min-max damage.
func registerOnUseDamage(itemID int32, spellID int32, school core.SpellSchool, minDamage, maxDamage float64, cooldown time.Duration, aoe bool) {
	core.NewItemEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: spellID},
			SpellSchool:      school,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskEmpty,
			Flags:            core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: cooldown,
				},
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if !aoe {
					spell.CalcAndDealDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeMagicHitAndCrit)
					return
				}
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					spell.CalcAndDealDamage(sim, aoeTarget, sim.Roll(minDamage, maxDamage), spell.OutcomeMagicHitAndCrit)
				}
			},
		})
		character.AddMajorCooldown(core.MajorCooldown{
			Type:  core.CooldownTypeDPS,
			Spell: spell,
		})
	})
}

func init() {
	core.AddEffectsToTest = false

	///////////////////////////////////////////////////////////////////////////
	//                                 Equip effects
	///////////////////////////////////////////////////////////////////////////

	// Mark of the Veteran: Equip: Improves your critical strike chance for all attacks and spells by 2%.
	for _, id := range []int32{MarkOfTheVeteranCritA, MarkOfTheVeteranCritB, MarkOfTheVeteranCritC, MarkOfTheVeteranCritD} {
		core.NewItemEffect(id, func(agent core.Agent) {
			agent.GetCharacter().AddStats(stats.Stats{
				stats.MeleeCrit: 2 * core.CritRatingPerCritChance,
				stats.SpellCrit: 2 * core.SpellCritRatingPerCritChance,
			})
		})
	}

	// Mark of the Veteran: Equip: Grants +5% increased spell hit chance for 20 sec when one of your spells is resisted.
	for _, id := range []int32{MarkOfTheVeteranSpellHitA, MarkOfTheVeteranSpellHitB} {
		itemID := id
		core.NewItemEffect(itemID, func(agent core.Agent) {
			character := agent.GetCharacter()
			buff := character.RegisterAura(core.Aura{
				ActionID: core.ActionID{ItemID: itemID},
				Label:    "Mark of the Veteran (Spell Hit)",
				Duration: time.Second * 20,
			}).AttachStatBuff(stats.SpellHit, 5*core.SpellHitRatingPerHitChance)
			core.MakePermanent(character.RegisterAura(core.Aura{
				Label: "Mark of the Veteran (Spell Hit) Trigger",
				OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if spell.ProcMask.Matches(core.ProcMaskSpellDamage) && !result.Landed() {
						buff.Activate(sim)
					}
				},
			}))
		})
	}

	// Mark of Thirst: Equip: Increases your attack, casting and movement speed by 20% against targets below 20% health.
	core.NewItemEffect(MarkOfThirst, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := character.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: MarkOfThirst},
			Label:    "Mark of Thirst",
			Duration: core.NeverExpires,
		}).AttachMultiplyAttackSpeed(&character.Unit, 1.2).AttachMultiplyCastSpeed(&character.Unit, 1.2)
		character.RegisterResetEffect(func(sim *core.Simulation) {
			sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
				if isExecute == 20 {
					aura.Activate(sim)
				}
			})
		})
	})

	// Royal Seal of Eldre'Thalas: Equip: +12 to all attributes.
	core.NewItemEffect(RoyalSealOfEldreThalasAttributes, func(agent core.Agent) {
		agent.GetCharacter().AddStats(stats.Stats{
			stats.Strength: 12, stats.Agility: 12, stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 12,
		})
	})

	// Royal Seal of Eldre'Thalas: Equip: Reduces the cost of your spells by 2%.
	for _, id := range []int32{RoyalSealOfEldreThalasCostA, RoyalSealOfEldreThalasCostB} {
		core.NewItemEffect(id, func(agent core.Agent) {
			agent.GetCharacter().PseudoStats.SchoolCostMultiplier.AddToAllSchools(-2)
		})
	}

	// Blue Mottled Egg / Pink Speckled Egg: Equip: +10 All Resistances.
	for _, id := range []int32{BlueMottledEgg, PinkSpeckledEgg} {
		core.NewItemEffect(id, func(agent core.Agent) {
			agent.GetCharacter().AddStats(resistances(10))
		})
	}

	// Onyx Egg: Equip: Decreases damage taken by 1%.
	core.NewItemEffect(OnyxEgg, func(agent core.Agent) {
		agent.GetCharacter().PseudoStats.DamageTakenMultiplier *= 0.99
	})

	// Sawtooth Talisman: Equip: Your attacks ignore 250 of your enemies' Armor.
	// The second line ("ignore 5% of your enemies' Armor") is not modeled: the engine has no percentage armor penetration.
	core.NewItemEffect(SawtoothTalisman, func(agent core.Agent) {
		agent.GetCharacter().AddStat(stats.ArmorPenetration, 250)
	})

	// Shen'dralar Badge of Deterrence: Equip: Threat +5%
	core.NewItemEffect(ShendralarBadgeOfDeterrence, func(agent core.Agent) {
		agent.GetCharacter().PseudoStats.ThreatMultiplier *= 1.05
	})

	// Fetish of the Sand Reaver / Grace of Earth / Two-Faced Medallion: Equip: Reduces the threat you generate by 5%.
	// (Their Use effects are not modeled. Fetish of the Sand Reaver has the same equip effect, below.)
	for _, id := range []int32{GraceOfEarth, TwoFacedMedallion} {
		core.NewItemEffect(id, func(agent core.Agent) {
			agent.GetCharacter().PseudoStats.ThreatMultiplier *= 0.95
		})
	}

	// The Lion Horn of Stormwind: Equip: aura that increases nearby party members' armor by 250 and magic resistances by 10.
	// Only the wearer's own part is modeled.
	core.NewItemEffect(TheLionHornOfStormwind, func(agent core.Agent) {
		bonus := resistances(10)
		bonus[stats.BonusArmor] = 250
		agent.GetCharacter().AddStats(bonus)
	})

	///////////////////////////////////////////////////////////////////////////
	//                                 Use effects
	///////////////////////////////////////////////////////////////////////////

	// Gri'lek's Charm of Valor: Use: Increases the critical hit chance of Holy spells and physical attacks by 10% for 30 sec. (3 Min Cooldown)
	core.NewItemEffect(GrileksCharmOfValor, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := character.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: GrileksCharmOfValor},
			Label:    "Gri'lek's Charm of Valor",
			Duration: time.Second * 30,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				character.AddStatDynamic(sim, stats.MeleeCrit, 10*core.CritRatingPerCritChance)
				character.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexHoly] += 10 * core.SpellCritRatingPerCritChance
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				character.AddStatDynamic(sim, stats.MeleeCrit, -10*core.CritRatingPerCritChance)
				character.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexHoly] -= 10 * core.SpellCritRatingPerCritChance
			},
		})
		registerOnUseAura(character, GrileksCharmOfValor, aura, time.Minute*3, core.CooldownTypeDPS)
	})

	// Wushoolay's Charm of Nature: Use: Increases damage done and healing of your Nature spells by 20% for 15 sec. (3 Min Cooldown)
	core.NewItemEffect(WushoolaysCharmOfNature, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := character.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: WushoolaysCharmOfNature},
			Label:    "Wushoolay's Charm of Nature",
			Duration: time.Second * 15,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexNature] *= 1.2
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexNature] /= 1.2
			},
		})
		registerOnUseAura(character, WushoolaysCharmOfNature, aura, time.Minute*3, core.CooldownTypeDPS)
	})

	// Hazza'rah's Charm of Healing: Use: Increases the Priest's casting speed by 40% for 15 sec. (3 Min Cooldown)
	core.NewItemEffect(HazzarahsCharmOfHealing, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := character.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: HazzarahsCharmOfHealing},
			Label:    "Hazza'rah's Charm of Healing",
			Duration: time.Second * 15,
		}).AttachMultiplyCastSpeed(&character.Unit, 1.4)
		registerOnUseAura(character, HazzarahsCharmOfHealing, aura, time.Minute*3, core.CooldownTypeDPS)
	})

	// Mar'li's Eye: Use: Restores 60 mana every 5 sec for 30 sec. (3 Min Cooldown)
	core.NewItemEffect(MarlisEye, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{ItemID: MarlisEye}
		manaMetrics := character.NewManaMetrics(actionID)
		var pa *core.PendingAction
		aura := character.RegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Mar'li's Eye",
			Duration: time.Second * 30,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				pa = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period:   time.Second * 5,
					NumTicks: 6,
					OnAction: func(sim *core.Simulation) {
						character.AddMana(sim, 60, manaMetrics)
					},
				})
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				if pa != nil {
					pa.Cancel(sim)
				}
			},
		})
		registerOnUseAura(character, MarlisEye, aura, time.Minute*3, core.CooldownTypeMana)
	})

	// Blessed Prayer Beads: Use: Increases healing done by spells and effects by up to 190 for 20 sec. (2 Min Cooldown)
	core.NewSimpleStatOffensiveTrinketEffect(BlessedPrayerBeads, stats.Stats{stats.HealingPower: 190}, time.Second*20, time.Minute*2)

	// Defensive trinkets (stat part only).
	// Blazing Emblem: Use: Increases Fire resistance by 50 and reduces all Fire damage taken by up to 25 for 15 sec. (10 Min Cooldown)
	core.NewSimpleStatDefensiveTrinketEffect(BlazingEmblem, stats.Stats{stats.FireResistance: 50}, time.Second*15, time.Minute*10)
	// Heart of the Scale: Use: Increases Fire Resistance by 20 and deals 20 Fire damage to attackers for 5 min. (30 Min Cooldown)
	core.NewSimpleStatDefensiveTrinketEffect(HeartOfTheScale, stats.Stats{stats.FireResistance: 20}, time.Minute*5, time.Minute*30)
	// Petrified Scarab: Use: Increases your spell resistances by 100 for 1 min. (3 Min Cooldown)
	// The "-10 per hostile spell" decay is not modeled.
	core.NewSimpleStatDefensiveTrinketEffect(PetrifiedScarab, resistances(100), time.Minute, time.Minute*3)
	// Ragged John's Neverending Cup: Use: Increases Stamina by 28 and reduces physical damage taken by 22 for 10 min. (30 Min Cooldown)
	// The flat damage reduction is not modeled (no flat per-hit reduction in the engine).
	core.NewSimpleStatDefensiveTrinketEffect(RaggedJohnsNeverendingCup, stats.Stats{stats.Stamina: 28}, time.Minute*10, time.Minute*30)

	// Aegis of Preservation: Use: Decreases damage taken by 10%, and heals for 30% of damage taken for 20 sec. (5 Min Cooldown)
	// The healing part is not modeled.
	core.NewItemEffect(AegisOfPreservation, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := character.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: AegisOfPreservation},
			Label:    "Aegis of Preservation",
			Duration: time.Second * 20,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.DamageTakenMultiplier *= 0.9
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.DamageTakenMultiplier /= 0.9
			},
		})
		registerOnUseAura(character, AegisOfPreservation, aura, time.Minute*5, core.CooldownTypeSurvival)
	})

	// Fetish of the Sand Reaver: Use: Reduces the threat you generate by 70% for 20 sec. (3 Min Cooldown)
	// Not auto-used: it only lowers threat. Available for APLs.
	core.NewItemEffect(FetishOfTheSandReaver, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.PseudoStats.ThreatMultiplier *= 0.95 // Equip: Reduces the threat you generate by 5%.
		actionID := core.ActionID{ItemID: FetishOfTheSandReaver}
		aura := character.RegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Fetish of the Sand Reaver",
			Duration: time.Second * 20,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.ThreatMultiplier *= 0.3
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.ThreatMultiplier /= 0.3
			},
		})
		character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})
	})

	// Fetish of Chitinous Spikes: Use: Spikes sprout from you causing 82 Nature damage to attackers when hit. Lasts 30 sec. (3 Min Cooldown)
	core.NewItemEffect(FetishOfChitinousSpikes, func(agent core.Agent) {
		character := agent.GetCharacter()
		spikes := character.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 26168},
			SpellSchool:      core.SpellSchoolNature,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskEmpty,
			Flags:            core.SpellFlagBinary | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 82, spell.OutcomeMagicHit)
			},
		})
		aura := character.RegisterAura(core.Aura{
			ActionID: core.ActionID{ItemID: FetishOfChitinousSpikes},
			Label:    "Fetish of Chitinous Spikes",
			Duration: time.Second * 30,
			OnSpellHitTaken: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMelee) {
					spikes.Cast(sim, spell.Unit)
				}
			},
		})
		registerOnUseAura(character, FetishOfChitinousSpikes, aura, time.Minute*3, core.CooldownTypeSurvival)
	})

	// Damage trinkets.
	// Shard of the Fallen Star: Use: Calls down a meteor, burning all enemies within the area for 400 to 442 total Fire damage. (3 Min Cooldown)
	registerOnUseDamage(ShardOfTheFallenStar, 26789, core.SpellSchoolFire, 400, 442, time.Minute*3, true)
	// Ramstein's Lightning Bolts: Use: strike down all enemies around you for 200 to 440 Nature damage. (5 Min Cooldown)
	registerOnUseDamage(RamsteinsLightningBolts, 17668, core.SpellSchoolNature, 200, 440, time.Minute*5, true)
	// Smokey's Lighter: Use: Deals 125 Fire damage to all targets in a cone in front of the caster. (5 Min Cooldown)
	registerOnUseDamage(SmokeysLighter, 17283, core.SpellSchoolFire, 125, 125, time.Minute*5, true)

	// Chained Essence of Eranikus: Use: Poisons all enemies in an 8 yard radius around the caster.
	// Victims of the poison suffer 50 Nature damage every 5 sec for 45 sec. (15 Min Cooldown)
	core.NewItemEffect(ChainedEssenceOfEranikus, func(agent core.Agent) {
		character := agent.GetCharacter()
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 12766},
			SpellSchool:      core.SpellSchoolNature,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskEmpty,
			Flags:            core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 15,
				},
			},
			Dot: core.DotConfig{
				Aura: core.Aura{
					Label: "Poison Cloud (Chained Essence of Eranikus)",
				},
				NumberOfTicks: 9,
				TickLength:    time.Second * 5,
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
					dot.Snapshot(target, 50, false)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					if spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit).Landed() {
						spell.Dot(aoeTarget).Apply(sim)
					}
				}
			},
		})
		character.AddMajorCooldown(core.MajorCooldown{
			Type:  core.CooldownTypeDPS,
			Spell: spell,
		})
	})

	core.AddEffectsToTest = true
}
