package item_sets

import (
	"github.com/wowsims/classic/sim/common/guardians"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
	"time"
)

///////////////////////////////////////////////////////////////////////////
//                                 Cloth
///////////////////////////////////////////////////////////////////////////

var ItemSetNecropileRaiment = core.NewItemSet(core.ItemSet{
	Name: "Necropile Raiment",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// +10 Intellect.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Intellect, 10)
		},
		// +15 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 15)
		},
		// +20 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(20)
		},
		// Increases damage and healing done by magical spells and effects by up to 40.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellPower, 40)
		},
	},
})

var ItemSetIronweaveBattlesuit = core.NewItemSet(core.ItemSet{
	Name: "Ironweave Battlesuit",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Increases your chance to resist Silence and Interrupt effects by 10%. (nothing to model)
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			_ = character
		},
		// +10 Resistances/+200 Armor.
		8: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(10)
			character.AddStat(stats.Armor, 200)
		},
	},
})

var ItemSetThePostmaster = core.NewItemSet(core.ItemSet{
	Name: "The Postmaster",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Restores 5 mana per 5 sec.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MP5, 5)
		},
		// +15 Intellect.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Intellect, 15)
		},
		// +15 Stamina.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 15)
		},
		// Increases run speed by 25%. (movement, nothing to model)
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			_ = character
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Leather
///////////////////////////////////////////////////////////////////////////

var ItemSetCadaverousGarb = core.NewItemSet(core.ItemSet{
	Name: "Cadaverous Garb",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to hit by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeHit, 1*core.MeleeHitRatingPerHitChance)
		},
		// +15 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 15)
		},
		// +20 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(20)
		},
		// +70 Attack Power.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 70)
			character.AddStat(stats.RangedAttackPower, 70)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Mail
///////////////////////////////////////////////////////////////////////////

var ItemSetBloodmailRegalia = core.NewItemSet(core.ItemSet{
	Name: "Bloodmail Regalia",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// +30 Attack Power.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 30)
			character.AddStat(stats.RangedAttackPower, 30)
		},
		// +20 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 20)
		},
		// +20 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(20)
		},
		// Improves your chance to get a critical strike by 2%.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 2*core.CritRatingPerCritChance)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Plate
///////////////////////////////////////////////////////////////////////////

var ItemSetDeathboneGuardian = core.NewItemSet(core.ItemSet{
	Name: "Deathbone Guardian",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Increased Defense +5.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Defense, 5)
		},
		// +15 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 15)
		},
		// +20 All Resistances.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddResistances(20)
		},
		// Improves your chance to hit by 3%.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeHit, 3*core.MeleeHitRatingPerHitChance)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Other
///////////////////////////////////////////////////////////////////////////

var ItemSetSpidersKiss = core.NewItemSet(core.ItemSet{
	Name: "Spider's Kiss",
	Bonuses: map[int32]core.ApplyEffect{
		// Chance on Hit: Immobilizes the target and lowers their armor by 100 for 10 sec.
		// Unsure about exlusive effects with this aura also looks like it might be lowering the characters armor instead of the enemy?
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			procAura := character.NewTemporaryStatsAura("Spider's Kiss", core.ActionID{SpellID: 17333}, stats.Stats{stats.Armor: -100}, time.Second*10)
			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				ActionID:   core.ActionID{SpellID: 17333},
				Name:       "Spider's Kiss",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.05,
				Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
					procAura.Activate(sim)
				},
			})
		},
	},
})

var ItemSetDalRendsArms = core.NewItemSet(core.ItemSet{
	Name: "Dal'Rend's Arms",
	Bonuses: map[int32]core.ApplyEffect{
		// +50 Attack Power.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 50)
			character.AddStat(stats.RangedAttackPower, 50)
		},
	},
})

var ItemSetShardOfTheGods = core.NewItemSet(core.ItemSet{
	Name: "Shard of the Gods",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// +27 to all attributes.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStats(stats.Stats{stats.Strength: 27, stats.Agility: 27, stats.Stamina: 27, stats.Intellect: 27, stats.Spirit: 27})
		},
	},
})

var ItemSetSpiritOfEskhandar = core.NewItemSet(core.ItemSet{
	Name: "Spirit of Eskhandar",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Increases your chance to dodge an attack by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Dodge, 1*core.DodgeRatingPerDodgeChance)
		},
		// Improves your chance to get a critical strike by 1%.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// 12% chance on a melee critical hit to call forth the spirit of Eskhandar to protect you in battle for 2 min. Can only occur once every 2 min.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				Name:       "Call of Eskhandar Trigger",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeCrit,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.12,
				ICD:        time.Minute * 2,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					for _, petAgent := range character.PetAgents {
						if eskhandar, ok := petAgent.(*guardians.Eskhandar); ok {
							eskhandar.EnableWithTimeout(sim, eskhandar, time.Minute*2)
							break
						}
					}
				},
			})
		},
	},
})

var ItemSetMajorMojoInfusion = core.NewItemSet(core.ItemSet{
	Name: "Major Mojo Infusion",
	Bonuses: map[int32]core.ApplyEffect{
		// +30 Attack Power.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStats(stats.Stats{
				stats.AttackPower:       30,
				stats.RangedAttackPower: 30,
			})
		},
	},
})

var ItemSetOverlordsResolution = core.NewItemSet(core.ItemSet{
	Name: "Overlord's Resolution",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Increased Defense +10.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Defense, 10)
		},
	},
})

var ItemSetPrayerOfThePrimal = core.NewItemSet(core.ItemSet{
	Name: "Prayer of the Primal",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases healing done by up to 33
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.HealingPower, 33)
		},
	},
})

var ItemSetPrimalBlessing = core.NewItemSet(core.ItemSet{
	Name: "Primal Blessing",
	Bonuses: map[int32]core.ApplyEffect{
		// Grants a small chance when ranged or melee damage is dealt to infuse the wielder with a blessing from the Primal Gods.
		// Ranged and melee attack power increased by 300 for 12 sec.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()

			aura := character.RegisterAura(core.Aura{
				ActionID: core.ActionID{SpellID: 24255},
				Label:    "Primal Blessing",
				Duration: time.Second * 12,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					character.AddStatsDynamic(sim, stats.Stats{
						stats.AttackPower:       300,
						stats.RangedAttackPower: 300,
					})
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					character.AddStatsDynamic(sim, stats.Stats{
						stats.AttackPower:       -300,
						stats.RangedAttackPower: -300,
					})
				},
			})

			core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
				Name:       "Primal Blessing Trigger",
				Callback:   core.CallbackOnSpellHitDealt,
				ProcMask:   core.ProcMaskMeleeOrRanged,
				Outcome:    core.OutcomeLanded,
				ProcChance: 0.05,
				ICD:        time.Second * 72,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					aura.Activate(sim)
				},
			})
		},
	},
})

var ItemSetTwinBladesofHakkari = core.NewItemSet(core.ItemSet{
	Name: "The Twin Blades of Hakkari",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases Swords +6
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.PseudoStats.SwordsSkill += 6
		},
	},
})

var ItemSetZanzilsConcentration = core.NewItemSet(core.ItemSet{
	Name: "Zanzil's Concentration",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Increases damage and healing done by magical spells and effects by up to 18.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellPower, 18)
		},
	},
})
