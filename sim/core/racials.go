package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func applyRaceEffects(agent Agent) {
	character := agent.GetCharacter()

	switch character.Race {
	case proto.Race_RaceDwarf:
		// Server (Spell.csv): Frost Resistance 20 (20596); Thunderer (20595): Maces, Two-Handed Maces and Guns +5.
		character.AddStat(stats.FrostResistance, 20)
		character.GunSpecializationAura()
		character.MaceSpecializationAura()

		actionID := ActionID{SpellID: 20594}

		// Server: physical damage taken -20% for 15 sec (Spell.csv).
		stoneFormAura := character.NewTemporaryStatsAuraWrapped("Stoneform", actionID, stats.Stats{}, time.Second*15, func(aura *Aura) {
			aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
				aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] *= 0.8
			})
			aura.ApplyOnExpire(func(aura *Aura, sim *Simulation) {
				aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical] /= 0.8
			})
		})

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 5 /* server (Spell.csv) */,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				stoneFormAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeSurvival,
			ShouldActivate: func(s *Simulation, c *Character) bool {
				// Only castable with manual APL Action
				return false
			},
		})
	case proto.Race_RaceGnome:
		// Server (Spell.csv): Arcane Resistance 20 (20592), Expansive Mind +10% Intellect (20591).
		character.AddStat(stats.ArcaneResistance, 20)
		character.MultiplyStat(stats.Intellect, 1.10)
	case proto.Race_RaceHuman:
		character.MultiplyStat(stats.Spirit, 1.10) // server: The Human Spirit (20598) +10%
		character.SwordSpecializationAura()
		character.MaceSpecializationAura()
	case proto.Race_RaceNightElf:
		// Server (Spell.csv): Nature Resistance 20 (20583); Quickness (20582): Agility, movement and casting
		// speed +5% (classic: +1% dodge).
		character.AddStat(stats.NatureResistance, 20)
		character.MultiplyStat(stats.Agility, 1.05)
		character.MultiplyCastSpeed(1.05)
	case proto.Race_RaceOrc:
		character.AxeSpecializationAura()

		if character.Class == proto.Class_ClassHunter || character.Class == proto.Class_ClassWarlock {
			// Command (server 20575): damage dealt by Hunter and Warlock pets increased by 10%.
			for _, pet := range character.Pets {
				if !pet.IsGuardian() {
					pet.PseudoStats.DamageDealtMultiplier *= 1.10
				}
			}
		}

		// Blood Fury
		actionID := ActionID{SpellID: 20572}
		var bloodFuryAP float64
		bloodFuryAura := character.RegisterAura(Aura{
			Label:    "Blood Fury",
			ActionID: actionID,
			Duration: time.Second * 20, // server (Spell.csv)
			// Tooltip is misleading; ap bonus is base AP plus AP from current strength, does not include +attackpower on items/buffs
			OnGain: func(aura *Aura, sim *Simulation) {
				bloodFuryAP = (character.GetBaseStats()[stats.AttackPower] + (character.GetStat(stats.Strength) * APPerStrength[character.Class]) + (character.GetStat(stats.Agility) * APPerAgility[character.Class])) * 0.25
				character.AddStatDynamic(sim, stats.AttackPower, bloodFuryAP)
			},

			OnExpire: func(aura *Aura, sim *Simulation) {
				character.AddStatDynamic(sim, stats.AttackPower, -bloodFuryAP)
			},
		})

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{
					GCD: GCDDefault,
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				bloodFuryAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeDPS,
		})
	case proto.Race_RaceTauren:
		character.AddStat(stats.NatureResistance, 20) // server (20551)
		character.MultiplyStat(stats.Health, 1.05)
	case proto.Race_RaceTroll:
		// Server (Spell.csv): Light Weapons Specialization (20558): one-handed axes, daggers and thrown +5;
		// Hunting Weapons Specialization (26290): bows and polearms +5.
		character.BowSpecializationAura()
		character.ThrownSpecializationAura()
		character.GetOrRegisterAura(Aura{
			Label:      "Troll Weapon Skill Specialization",
			BuildPhase: CharacterBuildPhaseGear,
			Duration:   NeverExpires,
			OnGain: func(aura *Aura, sim *Simulation) {
				character.PseudoStats.AxesSkill += 5
				character.PseudoStats.DaggersSkill += 5
				character.PseudoStats.PolearmsSkill += 5
			},
		})

		// Monster Slaying (server 20557): +5% damage against Beasts and Dragonkin.
		character.Env.RegisterPostFinalizeEffect(func() {
			for _, t := range character.Env.Encounter.Targets {
				if t.MobType == proto.MobType_MobTypeBeast || t.MobType == proto.MobType_MobTypeDragonkin {
					for _, at := range character.AttackTables[t.UnitIndex] {
						at.DamageDealtMultiplier *= 1.05
						at.CritMultiplier *= 1.05
					}
				}
			}
		})

		// Berserking
		berserkingTimer := character.NewTimer()
		// Baseline cooldown
		makeBerserkingCooldown(character, 0, berserkingTimer)
		// Hard-coded percentage cooldown options
		makeBerserkingCooldown(character, .1, berserkingTimer)
		makeBerserkingCooldown(character, .15, berserkingTimer)
		makeBerserkingCooldown(character, .2, berserkingTimer)
		makeBerserkingCooldown(character, .25, berserkingTimer)
		makeBerserkingCooldown(character, .3, berserkingTimer)
	case proto.Race_RaceUndead:
		character.AddStat(stats.ShadowResistance, 20)
	}
}

// If customPercentage is 0, use the baseline Berserking calculations from health missing
// otherwise create a cooldown hard-coded to the custom percentage.
func makeBerserkingCooldown(character *Character, customPercentage float64, timer *Timer) {
	actionID := ActionID{SpellID: 26297, Tag: int32(customPercentage * 20)}

	label := "Berserking"
	if customPercentage != 0 {
		label = fmt.Sprintf("%s (%d)", label, int(customPercentage*100))
	}

	calcBerserkingPct := func() float64 {
		if customPercentage != 0 {
			return customPercentage
		}
		// from 10% at full health to 30% at 40% or less health
		switch hp := character.CurrentHealthPercent(); {
		case hp >= 1:
			return 0.1
		case hp <= 0.4:
			return 0.3
		default:
			return 0.1 + (1-hp)/3
		}
	}

	// Server (Spell.csv 26635): melee attack speed +10% at full health up to +30% when badly hurt, ranged attack
	// speed and casting speed +5%, for 20 sec.
	var berserkingHaste float64
	berserkingAura := character.RegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Duration: time.Second * 20,
		OnGain: func(aura *Aura, sim *Simulation) {
			berserkingHaste = 1 + calcBerserkingPct()
			character.MultiplyMeleeSpeed(sim, berserkingHaste)
			character.MultiplyRangedSpeed(sim, 1.05)
			character.MultiplyCastSpeed(1.05)

			if sim.Log != nil {
				character.Log(sim, "Berserking increased melee attack speed by %.2f%% (%.2f%% hp)", berserkingHaste*100-100, character.CurrentHealthPercent()*100)
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			character.MultiplyMeleeSpeed(sim, 1/berserkingHaste)
			character.MultiplyRangedSpeed(sim, 1/1.05)
			character.MultiplyCastSpeed(1 / 1.05)
		},
	})

	config := SpellConfig{
		ActionID: actionID,

		Cast: CastConfig{
			CD: Cooldown{
				Timer:    timer,
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			berserkingAura.Activate(sim)
		},
	}

	switch {
	case character.HasManaBar():
		config.ManaCost = ManaCostOptions{BaseCost: 0.07}
	case character.HasRageBar():
		config.RageCost = RageCostOptions{Cost: 5}
	case character.HasEnergyBar():
		config.EnergyCost = EnergyCostOptions{Cost: 10}
	}

	berserkingSpell := character.RegisterSpell(config)

	character.AddMajorCooldown(MajorCooldown{
		Spell: berserkingSpell,
		Type:  CooldownTypeDPS,
	})
}

func (character *Character) GetFaction() proto.Faction {
	if slices.Contains([]proto.Race{proto.Race_RaceHuman, proto.Race_RaceDwarf, proto.Race_RaceGnome, proto.Race_RaceNightElf}, character.Race) {
		return proto.Faction_Alliance
	} else if slices.Contains([]proto.Race{proto.Race_RaceOrc, proto.Race_RaceTroll, proto.Race_RaceTauren, proto.Race_RaceUndead}, character.Race) {
		return proto.Faction_Horde
	} else {
		return proto.Faction_Unknown
	}
}
