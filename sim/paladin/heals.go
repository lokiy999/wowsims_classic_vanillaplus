package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Paladin heals with the server's values (CSV's/Spell.csv: heal range, mana cost, level; Holy Light cast time index
// 20 = 2.5 sec, Flash of Light 16 = 1.5 sec). Spell power coefficient: cast time / 3.5, classic penalty below level 20.

type healRank struct {
	level   int32
	spellID int32
	min     float64
	max     float64
	mana    float64
}

func lowLevelPenalty(level int32) float64 {
	if level >= 20 {
		return 1
	}
	return 1 - float64(20-level)*0.0375
}

// Healing Light: +4% healing per rank. Searing Light: -4% healing per rank.
func (paladin *Paladin) healMultiplier() float64 {
	return (1 + 0.04*float64(paladin.Talents.HealingLight)) * (1 - 0.04*float64(paladin.Talents.SearingLight))
}

func (paladin *Paladin) RegisterHealingSpells() {
	// Holy Power: +2% crit per rank for Holy spells (the damage side is a school bonus; heals use their own crit).
	holyPowerCrit := 2 * float64(paladin.Talents.HolyPower) * core.SpellCritRatingPerCritChance

	// Light's Mercy: each Flash of Light has a 20% chance per rank to add a Spark of Light stack; each stack makes the
	// next Holy Light cast 20% faster (up to 5 stacks).
	sparkChance := 0.2 * float64(paladin.Talents.LightsMercy)
	var sparkOfLight *core.Aura
	if paladin.Talents.LightsMercy > 0 {
		sparkOfLight = paladin.RegisterAura(core.Aura{
			Label:     "Spark of Light",
			ActionID:  core.ActionID{SpellID: 35739},
			Duration:  30 * time.Second,
			MaxStacks: 5,
		})
	}

	// Codex of the Silver Hand: Holy Light -10% mana and +0.3 sec cast time per rank (the regen part is in talents.go).
	codex := paladin.Talents.CodexOfTheSilverHand

	holyLight := []healRank{
		{1, 635, 39, 47, 35}, {6, 639, 76, 90, 60}, {14, 647, 159, 187, 110}, {22, 1026, 310, 356, 190},
		{30, 1042, 491, 553, 275}, {38, 3472, 698, 780, 365}, {46, 10328, 945, 1053, 465}, {54, 10329, 1246, 1388, 580},
		{60, 25292, 1590, 1770, 660},
	}
	for i, rank := range holyLight {
		if rank.level > paladin.Level {
			break
		}
		rank := rank
		paladin.HolyLight = append(paladin.HolyLight, paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_PaladinHolyLight,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.mana,
				Multiplier: 100 - 10*codex,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: 2500*time.Millisecond + 300*time.Millisecond*time.Duration(codex),
				},
				ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
					if sparkOfLight != nil && sparkOfLight.IsActive() {
						cast.CastTime = time.Duration(float64(cast.CastTime) * max(0, 1-0.2*float64(sparkOfLight.GetStacks())))
						sparkOfLight.Deactivate(sim)
					}
				},
			},

			BonusCritRating:  holyPowerCrit,
			BonusCoefficient: 2.5 / 3.5 * lowLevelPenalty(rank.level),
			DamageMultiplier: paladin.healMultiplier(),
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
			},
		}))
	}

	flashOfLight := []healRank{
		{20, 19750, 62, 72, 40}, {26, 19939, 96, 110, 60}, {34, 19940, 145, 163, 80}, {42, 19941, 197, 221, 105},
		{50, 19942, 267, 299, 135}, {58, 19943, 343, 383, 170},
	}
	for i, rank := range flashOfLight {
		if rank.level > paladin.Level {
			break
		}
		rank := rank
		paladin.FlashOfLight = append(paladin.FlashOfLight, paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_PaladinFlashOfLight,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{FlatCost: rank.mana},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: 1500 * time.Millisecond,
				},
			},

			BonusCritRating:  holyPowerCrit,
			BonusCoefficient: 1.5 / 3.5 * lowLevelPenalty(rank.level),
			DamageMultiplier: paladin.healMultiplier(),
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
				if sparkOfLight != nil && sim.RandomFloat("Light's Mercy") < sparkChance {
					sparkOfLight.Activate(sim)
					sparkOfLight.AddStack(sim)
				}
			},
		}))
	}

	paladin.registerHolyShockHeal(holyPowerCrit)
}

// Holy Shock on an ally (talent): heals for the same amount as the damage version (25914/25913/25903). It uses its
// own 10 sec cooldown here (the damage version is not used by a healer).
func (paladin *Paladin) registerHolyShockHeal(holyPowerCrit float64) {
	if !paladin.Talents.HolyShock {
		return
	}
	ranks := []healRank{
		{30, 25914, 134, 150, 225}, {44, 25913, 242, 264, 275}, {58, 25903, 405, 435, 325},
	}
	timer := paladin.NewTimer()
	for i, rank := range ranks {
		if rank.level > paladin.Level {
			break
		}
		rank := rank
		paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_PaladinHolyShockHeal,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{FlatCost: rank.mana},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    timer,
					Duration: 10 * time.Second,
				},
			},

			BonusCritRating:  holyPowerCrit,
			BonusCoefficient: 0.429,
			DamageMultiplier: paladin.healMultiplier(),
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
			},
		})
	}
}
