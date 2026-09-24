package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Priest heals with the server's values (CSV's/Spell.csv: EffectBasePoints+1 to +EffectDieSides, mana cost, level).
// Cast times are the classic ones (the server data only has a cast time index; see TODO.md question 13). Spell power
// coefficients use the classic rule: cast time / 3.5 (min 1.5 sec), HoTs duration / 15, with the classic penalty for
// ranks learned below level 20.

type healRank struct {
	level   int
	spellID int32
	min     float64
	max     float64
	mana    float64
}

// Classic penalty for spells learned below level 20.
func lowLevelPenalty(level int) float64 {
	if level >= 20 {
		return 1
	}
	return 1 - float64(20-level)*0.0375
}

func (priest *Priest) healMultiplier() float64 {
	// Spiritual Healing (DBC 14898-15356): +3% healing per rank.
	return 1 + 0.03*float64(priest.Talents.SpiritualHealing)
}

type directHealConfig struct {
	spellCode    int32
	ranks        []healRank
	castTime     time.Duration
	coefficient  float64
	costModifier int32 // percent, added to the cost multiplier
	targets      func(sim *core.Simulation, target *core.Unit) []*core.Unit
}

func (priest *Priest) registerDirectHeal(cfg directHealConfig) []*core.Spell {
	spells := make([]*core.Spell, len(cfg.ranks)+1)
	for i, rank := range cfg.ranks {
		if rank.level > int(priest.Level) {
			break
		}
		rank := rank
		spells[i+1] = priest.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   cfg.spellCode,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.mana,
				Multiplier: 100 + cfg.costModifier,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: cfg.castTime,
				},
			},

			BonusCoefficient: cfg.coefficient * lowLevelPenalty(rank.level),
			DamageMultiplier: priest.healMultiplier(),
			ThreatMultiplier: 0.5, // healing threat is half the amount healed

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				targets := []*core.Unit{target}
				if cfg.targets != nil {
					targets = cfg.targets(sim, target)
				}
				for _, t := range targets {
					spell.CalcAndDealHealing(sim, t, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
				}
			},
		})
	}
	return spells
}

// The party of the target (all players and pets in it), or just the target when it is not in a party.
func partyOf(target *core.Unit) []*core.Unit {
	agent := target.Env.Raid.GetPlayerFromUnit(target)
	if agent == nil || agent.GetCharacter().Party == nil {
		return []*core.Unit{target}
	}
	units := []*core.Unit{}
	for _, member := range agent.GetCharacter().Party.PlayersAndPets {
		units = append(units, &member.GetCharacter().Unit)
	}
	return units
}

func (priest *Priest) registerHealingSpells() {
	divineFury := time.Millisecond * 100 * time.Duration(priest.Talents.DivineFury)
	improvedHealing := -5 * priest.Talents.ImprovedHealing // DBC: -5% mana per rank

	priest.Heal = priest.registerDirectHeal(directHealConfig{
		spellCode: SpellCode_PriestHeal,
		ranks: []healRank{
			{16, 2054, 295, 341, 170}, {22, 2055, 429, 491, 230}, {28, 6063, 566, 642, 290}, {34, 6064, 712, 804, 350},
		},
		castTime:     3*time.Second - divineFury,
		coefficient:  3.0 / 3.5,
		costModifier: improvedHealing,
	})

	priest.GreaterHeal = priest.registerDirectHeal(directHealConfig{
		spellCode: SpellCode_PriestGreaterHeal,
		ranks: []healRank{
			{40, 2060, 1263, 1423, 500}, {46, 10963, 1611, 1807, 660}, {52, 10964, 2011, 2252, 800},
			{58, 10965, 2516, 2808, 960}, {60, 25314, 2752, 3071, 1040},
		},
		castTime:     3*time.Second - divineFury,
		coefficient:  3.0 / 3.5,
		costModifier: improvedHealing,
	})

	priest.FlashHeal = priest.registerDirectHeal(directHealConfig{
		spellCode: SpellCode_PriestFlashHeal,
		ranks: []healRank{
			{20, 2061, 193, 237, 140}, {26, 9472, 258, 314, 180}, {32, 9473, 327, 393, 210}, {38, 9474, 400, 478, 250},
			{44, 10915, 518, 616, 300}, {50, 10916, 644, 764, 360}, {56, 10917, 812, 958, 440},
		},
		castTime:    1500 * time.Millisecond,
		coefficient: 1.5 / 3.5,
	})

	// Prayer of Healing heals every member of the target's party. Classic coefficient: (3 / 3.5) / 3.
	priest.PrayerOfHealing = priest.registerDirectHeal(directHealConfig{
		ranks: []healRank{
			{30, 596, 301, 321, 460}, {40, 996, 444, 472, 640}, {50, 10960, 657, 695, 885}, {60, 10961, 939, 991, 1180},
			{60, 25316, 1041, 1099, 1230},
		},
		castTime:     3 * time.Second,
		coefficient:  3.0 / 3.5 / 3,
		costModifier: -15 * priest.Talents.ImprovedPrayerOfHealing, // DBC: -15% mana per rank
		targets: func(sim *core.Simulation, target *core.Unit) []*core.Unit {
			return partyOf(target)
		},
	})

	priest.registerRenewSpell()
	priest.registerPowerWordShieldSpell()
}

// Renew: heal over 15 sec in 5 ticks. Improved Renew +5% per rank.
func (priest *Priest) registerRenewSpell() {
	ranks := []healRank{
		{8, 139, 9, 9, 30}, {14, 6074, 20, 20, 65}, {20, 6075, 35, 35, 105}, {26, 6076, 49, 49, 140},
		{32, 6077, 63, 63, 170}, {38, 6078, 80, 80, 205}, {44, 10927, 102, 102, 250}, {50, 10928, 130, 130, 305},
		{56, 10929, 162, 162, 365}, {60, 25315, 194, 194, 410},
	}
	priest.Renew = make([]*core.Spell, len(ranks)+1)
	for i, rank := range ranks {
		if rank.level > int(priest.Level) {
			break
		}
		rank := rank
		priest.Renew[i+1] = priest.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_PriestRenew,
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.mana,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			DamageMultiplier: priest.healMultiplier() * (1 + 0.05*float64(priest.Talents.ImprovedRenew)),
			ThreatMultiplier: 0.5,

			Hot: core.DotConfig{
				Aura: core.Aura{
					Label: "Renew",
				},
				NumberOfTicks:    5,
				TickLength:       3 * time.Second,
				BonusCoefficient: 0.2 * lowLevelPenalty(rank.level), // 15 sec / 15 = 1.0 over 5 ticks
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.SnapshotHeal(target, rank.min, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Hot(target).Apply(sim)
			},
		})
	}
}

// Power Word: Shield: absorbs damage for 30 sec, then Weakened Soul (15 sec) blocks another shield. Improved Power
// Word: Shield: +10% absorb and -2 sec Weakened Soul per rank. Classic spell power coefficient 0.1.
func (priest *Priest) registerPowerWordShieldSpell() {
	ranks := []healRank{
		{6, 17, 44, 44, 45}, {12, 592, 88, 88, 80}, {18, 600, 158, 158, 130}, {24, 3747, 234, 234, 175},
		{30, 6065, 301, 301, 210}, {36, 6066, 381, 381, 250}, {42, 10898, 484, 484, 300}, {48, 10899, 605, 605, 355},
		{54, 10900, 763, 763, 425}, {60, 10901, 942, 942, 500},
	}

	priest.WeakenedSouls = priest.NewRaidAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Weakened Soul",
			ActionID: core.ActionID{SpellID: 6788},
			Duration: 15*time.Second - 2*time.Second*time.Duration(priest.Talents.ImprovedPowerWordShield),
		})
	})

	priest.PowerWordShield = make([]*core.Spell, len(ranks)+1)
	for i, rank := range ranks {
		if rank.level > int(priest.Level) {
			break
		}
		rank := rank
		priest.PowerWordShield[i+1] = priest.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_PriestPowerWordShield,
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.mana,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return !priest.WeakenedSouls.Get(target).IsActive()
			},

			DamageMultiplier: 1 + 0.1*float64(priest.Talents.ImprovedPowerWordShield),
			ThreatMultiplier: 0.5,

			Shield: core.ShieldConfig{
				Aura: core.Aura{
					Label:    "Power Word: Shield",
					Duration: 30 * time.Second,
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Shield(target).Apply(sim, rank.min+0.1*lowLevelPenalty(rank.level)*spell.HealingPower(target))
				priest.WeakenedSouls.Get(target).Activate(sim)
			},
		})
	}
}
