package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                            Phase 1 Item Sets - Molten Core
///////////////////////////////////////////////////////////////////////////

var ItemSetNightslayerArmor = core.NewItemSet(core.ItemSet{
	Name: "Nightslayer Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases your maximum Energy by 10.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			if c.HasEnergyBar() {
				c.EnableEnergyBar(c.MaxEnergy() + 10)
			}
		},
		// Increases the effectiveness of your finishing moves by 10%.
		4: func(agent core.Agent) {
			c := agent.(RogueAgent).GetRogue()
			c.RegisterAura(core.Aura{
				Label: "Improved Finishing Moves",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					c.Eviscerate.DamageMultiplier *= 1.10
					c.Rupture.DamageMultiplier *= 1.10
				},
			})
		},
		// Reduces the cooldown of Vanish and Cloak of Shadows by 1 min.
		6: func(agent core.Agent) {
			c := agent.(RogueAgent).GetRogue()
			c.RegisterAura(core.Aura{
				Label: "Improved Vanish",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					c.Vanish.CD.Duration -= time.Minute
					// Cloak of Shadows is not implemented in the sim.
				},
			})
		},
		// Heals the rogue for 50% of his health when Vanish is performed.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			healthMetrics := c.NewHealthMetrics(core.ActionID{SpellID: 23582})

			core.MakePermanent(c.RegisterAura(core.Aura{
				Label: "Clean Escape",
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
					if spell.SpellCode == SpellCode_RogueVanish {
						c.GainHealth(sim, c.MaxHealth()*0.5, healthMetrics)
					}
				},
			}))
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Phase 2 Item Sets - Dire Maul
///////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////
//                            Phase 3 Item Sets - BWL
///////////////////////////////////////////////////////////////////////////

var ItemSetBloodfangArmor = core.NewItemSet(core.ItemSet{
	Name: "Bloodfang Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases the chance to apply poisons to your target by 5%.
		2: func(agent core.Agent) {
			c := agent.(RogueAgent).GetRogue()
			c.RegisterAura(core.Aura{
				Label: "Improved Poisons",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					c.additivePoisonBonusChance += .05
				},
			})
		},
		// Improves the threat reduction of Feint by 50%.
		4: func(agent core.Agent) {
			// Feint threat reduction not currently implemented in feint.go
		},
		// Decreases the cost of your finishing moves by 10 Energy.
		6: func(agent core.Agent) {
			c := agent.(RogueAgent).GetRogue()

			core.MakePermanent(c.RegisterAura(core.Aura{
				Label: "Improved Finishing Moves Energy",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					for _, finisher := range c.Finishers {
						finisher.Cost.FlatModifier -= 10
					}
				},
			}))
		},
		// Gives the Rogue a chance to inflict 283 to 317 damage on the target and heal the Rogue for 50 health every 1 sec. for 12 sec. on a melee hit.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()

			bloodfangHeal := c.GetOrRegisterSpell(core.SpellConfig{
				ActionID:    core.ActionID{SpellID: 23580},
				SpellSchool: core.SpellSchoolPhysical,
				DefenseType: core.DefenseTypeMelee,
				ProcMask:    core.ProcMaskEmpty,
				Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
				Hot: core.DotConfig{
					Aura: core.Aura{
						Label: "Bloodfang",
					},
					NumberOfTicks: 12,
					TickLength:    time.Second,
					OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
						dot.SnapshotBaseDamage = 50
					},
					OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
						dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
					},
				},
				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.Hot(&c.Unit).Apply(sim)
				},
			})

			procSpell := c.GetOrRegisterSpell(core.SpellConfig{
				ActionID:    core.ActionID{SpellID: 23581},
				SpellSchool: core.SpellSchoolPhysical,
				DefenseType: core.DefenseTypeMelee,
				ProcMask:    core.ProcMaskEmpty,
				Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

				DamageMultiplier: 1,
				ThreatMultiplier: 1,

				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.CalcAndDealDamage(sim, target, sim.Roll(283, 317), spell.OutcomeMagicCrit)
				},
			})

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				Name:              "Bloodfang",
				Callback:          core.CallbackOnSpellHitDealt,
				Outcome:           core.OutcomeLanded,
				ProcMask:          core.ProcMaskMelee,
				SpellFlagsExclude: core.SpellFlagSuppressWeaponProcs,
				PPM:               1,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					procSpell.Cast(sim, result.Target)
					bloodfangHeal.Cast(sim, result.Target)
				},
			})
		},
	},
})

var ItemSetMadcapsOutfit = core.NewItemSet(core.ItemSet{
	Name: "Madcap's Outfit",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Attack Power.
		2: func(agent core.Agent) {
			c := agent.(RogueAgent).GetRogue()
			c.AddStats(stats.Stats{
				stats.AttackPower:       20,
				stats.RangedAttackPower: 20,
			})
		},
		// Decrease the energy cost of Eviscerate and Rupture by 5.
		3: func(agent core.Agent) {
			c := agent.(RogueAgent).GetRogue()

			core.MakePermanent(c.RegisterAura(core.Aura{
				Label: "Improved Eviscerate and Rupture",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					c.Eviscerate.Cost.FlatModifier -= 5
					c.Rupture.Cost.FlatModifier -= 5
				},
			}))
		},
		// Decreases the cooldown of Blind and Cloak of Shadows abilities by 60 sec.
		5: func(agent core.Agent) {
			// Blind and Cloak of Shadows are not implemented in the sim.
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Phase 4 Item Sets - AQ
///////////////////////////////////////////////////////////////////////////

// https://www.wowhead.com/classic/item-set=512/darkmantle-armor
var ItemSetDarkmantleArmor = core.NewItemSet(core.ItemSet{
	Name: "Darkmantle Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// +10 Resistances/+200 Armor.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(10)
			c.AddStat(stats.Armor, 200)
		},
		// Chance on melee attack to restore 35 energy.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 27787}
			energyMetrics := c.NewEnergyMetrics(actionID)

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID: actionID,
				Name:     "Rogue Armor Energize",
				Callback: core.CallbackOnSpellHitDealt,
				Outcome:  core.OutcomeLanded,
				ProcMask: core.ProcMaskMeleeWhiteHit,
				PPM:      1,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if c.HasEnergyBar() {
						c.AddEnergy(sim, 35, energyMetrics)
					}
				},
			})
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
			})
		},
		// +10 Resistances/+200 Armor.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(10)
			c.AddStat(stats.Armor, 200)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Phase 5 Item Sets - Naxx
///////////////////////////////////////////////////////////////////////////
