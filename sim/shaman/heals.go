package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shaman heals with the server's values (CSV's/Spell.csv: heal range, mana cost, level). Cast times follow the server's cast time index (16 = 1.5,
// 5 = 2.0, 20 = 2.5, 14 = 3.0, 22 = 3.5 sec, matched against spells with known classic cast times; the index table
// itself is not exported, see TODO.md question 13). Spell power coefficient: cast time / 3.5, with the classic
// penalty for ranks learned below level 20.

type healRank struct {
	level    int
	spellID  int32
	min      float64
	max      float64
	mana     float64
	castTime time.Duration
}

func lowLevelPenalty(level int) float64 {
	if level >= 20 {
		return 1
	}
	return 1 - float64(20-level)*0.0375
}

func healCoefficient(castTime time.Duration, level int) float64 {
	return max(castTime.Seconds(), 1.5) / 3.5 * lowLevelPenalty(level)
}

// The members of the target's party, or just the target when it is not in a party.
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

func (shaman *Shaman) RegisterHealingSpells() {
	// Purification (DBC 16178-16213): +2% healing per rank.
	healMultiplier := 1 + shaman.purificationHealingModifier()

	// Healing Way (29202-29206 / 29206-34288): each Healing Wave has a 20% chance per rank to make later Healing Waves
	// on that target heal 3% more for 30 sec, up to 10 stacks.
	healingWayChance := 0.2 * float64(shaman.Talents.HealingWay)
	healingWay := shaman.NewRaidAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:     "Healing Way",
			ActionID:  core.ActionID{SpellID: 29203},
			Duration:  30 * time.Second,
			MaxStacks: 10,
		})
	})

	// Improved Healing Wave: -0.1 sec cast time per rank.
	improvedHealingWave := 100 * time.Millisecond * time.Duration(shaman.Talents.ImprovedHealingWave)

	hwRanks := []healRank{
		{1, 331, 34, 44, 25, 1500 * time.Millisecond}, {6, 332, 64, 78, 50, 2 * time.Second},
		{12, 547, 129, 155, 90, 2500 * time.Millisecond}, {18, 913, 268, 316, 170, 3 * time.Second},
		{24, 939, 376, 440, 230, 3 * time.Second}, {32, 959, 536, 622, 300, 3 * time.Second},
		{40, 8005, 740, 854, 390, 3 * time.Second}, {48, 10395, 1017, 1167, 500, 3 * time.Second},
		{56, 10396, 1367, 1561, 640, 3 * time.Second}, {60, 25357, 1620, 1850, 710, 3 * time.Second},
	}
	shaman.HealingWave = make([]*core.Spell, len(hwRanks)+1)
	for i, rank := range hwRanks {
		if rank.level > int(shaman.Level) {
			break
		}
		rank := rank
		shaman.HealingWave[i+1] = shaman.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_ShamanHealingWave,
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{FlatCost: rank.mana},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: rank.castTime - improvedHealingWave,
				},
			},

			BonusCoefficient: healCoefficient(rank.castTime, rank.level),
			DamageMultiplier: healMultiplier,
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				way := healingWay.Get(target)
				bonus := 1 + 0.03*float64(way.GetStacks())
				spell.DamageMultiplier *= bonus
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
				spell.DamageMultiplier /= bonus
				if healingWayChance > 0 && sim.RandomFloat("Healing Way") < healingWayChance {
					way.Activate(sim)
					way.AddStack(sim)
				}
			},
		})
	}

	lhwRanks := []healRank{
		{20, 8004, 162, 186, 110, 0}, {28, 8008, 247, 281, 150, 0}, {36, 8010, 337, 381, 180, 0},
		{44, 10466, 458, 514, 240, 0}, {52, 10467, 631, 705, 300, 0}, {60, 10468, 832, 928, 380, 0},
	}
	shaman.LesserHealingWave = make([]*core.Spell, len(lhwRanks)+1)
	for i, rank := range lhwRanks {
		if rank.level > int(shaman.Level) {
			break
		}
		rank := rank
		shaman.LesserHealingWave[i+1] = shaman.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_ShamanLesserHealingWave,
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{FlatCost: rank.mana},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: 1500 * time.Millisecond,
				},
			},

			BonusCoefficient: healCoefficient(1500*time.Millisecond, rank.level),
			DamageMultiplier: healMultiplier,
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
			},
		})
	}

	shaman.registerChainHeal(healMultiplier)
}

// Chain Heal (server: 3.5 sec cast index, heals up to 3 targets, each jump 70% of the previous; classic was 2.5 sec
// and 50%). This server's Concussion (+1%/rank), Convection (-5%/rank mana), Call of Thunder (+2%/rank crit),
// Lightning Mastery (-0.2 sec/rank) and Lightning Overlord (crit refund) also apply to Chain Heal.
func (shaman *Shaman) registerChainHeal(healMultiplier float64) {
	ranks := []healRank{
		{40, 1064, 448, 518, 400, 3500 * time.Millisecond},
		{46, 10622, 566, 651, 500, 3500 * time.Millisecond},
		{54, 10623, 771, 880, 650, 3500 * time.Millisecond},
	}
	const jumpMultiplier = 0.7
	shaman.ChainHeal = make([]*core.Spell, len(ranks)+1)
	for i, rank := range ranks {
		if rank.level > int(shaman.Level) {
			break
		}
		rank := rank
		shaman.ChainHeal[i+1] = shaman.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_ShamanChainHeal,
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.mana,
				Multiplier: 100 - 5*shaman.Talents.Convection,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: rank.castTime - 200*time.Millisecond*time.Duration(shaman.Talents.LightningMastery),
				},
			},

			BonusCritRating:          2 * float64(shaman.Talents.CallOfThunder) * core.SpellCritRatingPerCritChance,
			BonusCoefficient:         healCoefficient(rank.castTime, rank.level),
			DamageMultiplier:         healMultiplier,
			DamageMultiplierAdditive: 1 + 0.01*float64(shaman.Talents.Concussion),
			ThreatMultiplier:         0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				targets := []*core.Unit{target}
				for _, member := range partyOf(target) {
					if len(targets) >= 3 {
						break
					}
					if member != target {
						targets = append(targets, member)
					}
				}
				multiplier := 1.0
				for _, t := range targets {
					spell.DamageMultiplier *= multiplier
					spell.CalcAndDealHealing(sim, t, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
					spell.DamageMultiplier /= multiplier
					multiplier *= jumpMultiplier
				}
			},
		})
	}
}
