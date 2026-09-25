package item_sets

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Sets in the DB that had no code. Bonus texts from the server's item dump (CSV's/VPlusItemDB.lua).
// Not modeled: Soulforge Armor 4 (crit proc, rate unknown), Vestments of the Virtuous 4 (damage shield proc),
// Augur's Regalia 2/3 (Frost Shock duration, spell range), Primal Batskin / Primal Fury / Gurubashi Ceremonial
// Blades (movement, kill and heal procs), The Fists of Fury (its items have no set name in the DB).

func addResistancesAndArmor(character *core.Character, res, armor float64) {
	character.AddResistances(res)
	character.AddStat(stats.BonusArmor, armor)
}

var ItemSetAugursRegalia = core.NewItemSet(core.ItemSet{
	Name: "Augur's Regalia",
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage and healing done by magical spells and effects by up to 40.
		5: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.SpellPower, 40)
		},
	},
})

var ItemSetChainOfTheScarletCrusade = core.NewItemSet(core.ItemSet{
	Name: "Chain of the Scarlet Crusade",
	Bonuses: map[int32]core.ApplyEffect{
		// Increased Defense +15.
		2: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Defense, 15)
		},
		// +30 Shadow Resistance/+200 Armor.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.ShadowResistance, 30)
			character.AddStat(stats.BonusArmor, 200)
		},
		// Improves your chance to hit with all attacks and spells by 2%.
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeHit, 2*core.MeleeHitRatingPerHitChance)
			character.AddStat(stats.SpellHit, 2*core.SpellHitRatingPerHitChance)
		},
		// Increases your damage against undead by 3%.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			if character.CurrentTarget.MobType == proto.MobType_MobTypeUndead {
				character.PseudoStats.DamageDealtMultiplier *= 1.03
			}
		},
		// Increases your attack and casting speed by 5%.
		6: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.PseudoStats.MeleeSpeedMultiplier *= 1.05
			character.PseudoStats.RangedSpeedMultiplier *= 1.05
			character.MultiplyCastSpeed(1.05)
		},
	},
})

var ItemSetDefiasLeather = core.NewItemSet(core.ItemSet{
	Name: "Defias Leather",
	ID:   161,
	Bonuses: map[int32]core.ApplyEffect{
		// +10 Armor.
		2: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.BonusArmor, 10)
		},
		// +10 Arcane Resistance.
		3: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.ArcaneResistance, 10)
		},
		// Increased Daggers +1.
		4: func(agent core.Agent) {
			agent.GetCharacter().PseudoStats.DaggersSkill += 1
		},
		// +10 Attack Power.
		5: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 10)
			character.AddStat(stats.RangedAttackPower, 10)
		},
	},
})

var ItemSetEmbraceOfTheViper = core.NewItemSet(core.ItemSet{
	Name: "Embrace of the Viper",
	ID:   162,
	Bonuses: map[int32]core.ApplyEffect{
		// Increases damage done by Nature spells and effects by up to 7.
		2: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.NaturePower, 7)
		},
		// Increased Staves +2.
		3: func(agent core.Agent) {
			agent.GetCharacter().PseudoStats.StavesSkill += 2
		},
		// Increases healing done by spells and effects by up to 11.
		4: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.HealingPower, 11)
		},
		// +10 Intellect.
		5: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Intellect, 10)
		},
	},
})

var ItemSetHaruspexsGarb = core.NewItemSet(core.ItemSet{
	Name: "Haruspex's Garb",
	Bonuses: map[int32]core.ApplyEffect{
		// +12 Stamina.
		2: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Stamina, 12)
		},
		// Improves your chance to get a critical strike with Nature spells by 3%.
		3: func(agent core.Agent) {
			agent.GetCharacter().PseudoStats.SchoolBonusCritChance[stats.SchoolIndexNature] += 3 * core.SpellCritRatingPerCritChance
		},
		// Increases healing done by spells and effects by up to 30.
		5: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.HealingPower, 30)
		},
	},
})

var ItemSetSoulforgeArmor = core.NewItemSet(core.ItemSet{
	Name: "Soulforge Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// +10 Resistances/+200 Armor.
		2: func(agent core.Agent) {
			addResistancesAndArmor(agent.GetCharacter(), 10, 200)
		},
		// +40 Attack Power.
		6: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 40)
			character.AddStat(stats.RangedAttackPower, 40)
		},
		// +10 Resistances/+200 Armor.
		8: func(agent core.Agent) {
			addResistancesAndArmor(agent.GetCharacter(), 10, 200)
		},
	},
})

var ItemSetTheGladiator = core.NewItemSet(core.ItemSet{
	Name: "The Gladiator",
	Bonuses: map[int32]core.ApplyEffect{
		// +14 Attack Power.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.AttackPower, 14)
			character.AddStat(stats.RangedAttackPower, 14)
		},
		// Improves your chance to get a critical strike by 1%.
		3: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// Increases your chance to dodge an attack by 2%.
		4: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Dodge, 2*core.DodgeRatingPerDodgeChance)
		},
		// +300 Armor.
		5: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.BonusArmor, 300)
		},
	},
})

var ItemSetVestmentsOfTheVirtuous = core.NewItemSet(core.ItemSet{
	Name: "Vestments of the Virtuous",
	Bonuses: map[int32]core.ApplyEffect{
		// +10 Resistances/+200 Armor.
		2: func(agent core.Agent) {
			addResistancesAndArmor(agent.GetCharacter(), 10, 200)
		},
		// Increases healing done by spells and effects by up to 44.
		6: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.HealingPower, 44)
		},
		// +10 Resistances/+200 Armor.
		8: func(agent core.Agent) {
			addResistancesAndArmor(agent.GetCharacter(), 10, 200)
		},
	},
})
