package druid

import (
	"math"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Druid heals with the server's values (CSV's/Spell.csv: heal range, mana cost, level, cast time index; 16 = 1.5,
// 5 = 2.0, 19 = 2.5 (assumed), 14 = 3.0, 22 = 3.5 sec). Spell power coefficient: cast time / 3.5 for direct heals,
// duration / 15 for HoTs (Regrowth split like classic: 28.6% direct, 70% over time), classic penalty below level 20.

type healRank struct {
	level    int
	spellID  int32
	min      float64
	max      float64
	mana     float64
	castTime time.Duration
	tick     float64 // HoT amount per tick
}

func lowLevelPenalty(level int) float64 {
	if level >= 20 {
		return 1
	}
	return 1 - float64(20-level)*0.0375
}

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

// Gift of Nature (17104, 24943-24946): healing +2% per rank.
func (druid *Druid) healMultiplier() float64 {
	return 1 + 0.02*float64(druid.Talents.GiftOfNature)
}

func (druid *Druid) RegisterHealingSpells() {
	druid.registerHealingTouchSpell()
	druid.registerRejuvenationSpell()
	druid.registerRegrowthSpell()
	druid.registerTranquilitySpell()
	druid.registerSwiftmendSpell()
	druid.registerNaturesSwiftnessCD()
}

// Healing Touch: the server changed the cast times of the low ranks. Naturalist -0.1 sec per rank, Tranquil Spirit
// -3% mana per rank, Moonglow -5% mana per rank.
func (druid *Druid) registerHealingTouchSpell() {
	ranks := []healRank{
		{1, 5185, 37, 51, 25, 1500 * time.Millisecond, 0}, {8, 5186, 65, 83, 40, 1500 * time.Millisecond, 0},
		{14, 5187, 117, 146, 65, 1500 * time.Millisecond, 0}, {20, 5188, 180, 221, 90, 1500 * time.Millisecond, 0},
		{26, 5189, 216, 268, 115, 1500 * time.Millisecond, 0}, {32, 6778, 463, 550, 190, 2 * time.Second, 0},
		{38, 8903, 733, 863, 285, 2500 * time.Millisecond, 0}, {44, 9758, 1128, 1324, 420, 3 * time.Second, 0},
		{50, 9888, 1670, 1951, 600, 3500 * time.Millisecond, 0}, {56, 9889, 2077, 2417, 720, 3500 * time.Millisecond, 0},
		{60, 25297, 2494, 2904, 800, 3500 * time.Millisecond, 0},
	}
	naturalist := 100 * time.Millisecond * time.Duration(druid.Talents.Naturalist)

	druid.HealingTouch = make([]*DruidSpell, len(ranks)+1)
	for i, rank := range ranks {
		if rank.level > int(druid.Level) {
			break
		}
		rank := rank
		druid.HealingTouch[i+1] = druid.RegisterSpell(Humanoid, core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.mana,
				Multiplier: 100 - 3*druid.Talents.TranquilSpirit - 5*druid.Talents.Moonglow,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: rank.castTime - naturalist,
				},
			},

			BonusCoefficient: rank.castTime.Seconds() / 3.5 * lowLevelPenalty(rank.level),
			DamageMultiplier: druid.healMultiplier(),
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
			},
		})
	}
}

// Rejuvenation: 4 ticks over 12 sec. Improved Rejuvenation +5% per rank, Power of Nature +25% duration per rank.
func (druid *Druid) registerRejuvenationSpell() {
	ranks := []healRank{
		{4, 774, 0, 0, 20, 0, 10}, {10, 1058, 0, 0, 30, 0, 20}, {16, 1430, 0, 0, 55, 0, 30}, {22, 2090, 0, 0, 75, 0, 50},
		{28, 2091, 0, 0, 100, 0, 60}, {34, 3627, 0, 0, 120, 0, 80}, {40, 8910, 0, 0, 150, 0, 100},
		{46, 9839, 0, 0, 175, 0, 120}, {52, 9840, 0, 0, 210, 0, 150}, {58, 9841, 0, 0, 250, 0, 190},
		{60, 25299, 0, 0, 270, 0, 230},
	}
	numTicks := int32(math.Round(4 * (1 + 0.25*float64(druid.Talents.PowerOfNature)))) // Power of Nature: +25% duration per rank

	druid.Rejuvenation = make([]*DruidSpell, len(ranks)+1)
	for i, rank := range ranks {
		if rank.level > int(druid.Level) {
			break
		}
		rank := rank
		druid.Rejuvenation[i+1] = druid.RegisterSpell(Humanoid, core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_DruidRejuvenation,
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.mana,
				Multiplier: 100 - 5*druid.Talents.Moonglow,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			DamageMultiplier: druid.healMultiplier() * (1 + 0.05*float64(druid.Talents.ImprovedRejuvenation)),
			ThreatMultiplier: 0.5,

			Hot: core.DotConfig{
				Aura: core.Aura{
					Label: "Rejuvenation",
				},
				NumberOfTicks:    numTicks,
				TickLength:       3 * time.Second,
				BonusCoefficient: 0.2 * lowLevelPenalty(rank.level), // 12 sec / 15 = 0.8 over 4 ticks
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.SnapshotHeal(target, rank.tick, isRollover)
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

// Regrowth: 2.0 sec cast, direct heal plus a HoT over 18 sec (6 ticks). Improved Regrowth +15% crit per rank (direct
// part), Moonglow -5% mana per rank, Power of Nature +25% HoT duration per rank.
func (druid *Druid) registerRegrowthSpell() {
	ranks := []healRank{
		{12, 8936, 84, 98, 120, 0, 16}, {18, 8938, 164, 188, 205, 0, 30}, {24, 8939, 240, 274, 280, 0, 45},
		{30, 8940, 318, 360, 350, 0, 60}, {36, 8941, 405, 457, 420, 0, 70}, {42, 9750, 511, 575, 510, 0, 90},
		{48, 9856, 646, 724, 615, 0, 115}, {54, 9857, 809, 905, 740, 0, 140}, {60, 9858, 1003, 1119, 880, 0, 175},
	}
	numTicks := int32(math.Round(6 * (1 + 0.25*float64(druid.Talents.PowerOfNature))))

	druid.Regrowth = make([]*DruidSpell, len(ranks)+1)
	for i, rank := range ranks {
		if rank.level > int(druid.Level) {
			break
		}
		rank := rank
		penalty := lowLevelPenalty(rank.level)
		druid.Regrowth[i+1] = druid.RegisterSpell(Humanoid, core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellCode:   SpellCode_DruidRegrowth,
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: rank.level,
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost:   rank.mana,
				Multiplier: 100 - 5*druid.Talents.Moonglow,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: 2 * time.Second,
				},
			},

			BonusCritRating:  15 * float64(druid.Talents.ImprovedRegrowth) * core.SpellCritRatingPerCritChance,
			BonusCoefficient: 0.286 * penalty,
			DamageMultiplier: druid.healMultiplier(),
			ThreatMultiplier: 0.5,

			Hot: core.DotConfig{
				Aura: core.Aura{
					Label: "Regrowth",
				},
				NumberOfTicks:    numTicks,
				TickLength:       3 * time.Second,
				BonusCoefficient: 0.7 / 6 * penalty,
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.SnapshotHeal(target, rank.tick, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(rank.min, rank.max), spell.OutcomeHealingCrit)
				spell.Hot(target).Apply(sim)
			},
		})
	}
}

// Tranquility (server): channeled for 10 sec, heals every party member each second (100/150/250/400 per rank, spells
// 35768-35771), 2 min cooldown. Tranquil Spirit: +20% effect and -3% mana per rank. The coefficient is a guess:
// (10 sec / 3.5) / 3 for an area heal, spread over the 10 ticks.
func (druid *Druid) registerTranquilitySpell() {
	ranks := []healRank{
		{30, 740, 0, 0, 375, 0, 100}, {40, 8918, 0, 0, 505, 0, 150}, {50, 9862, 0, 0, 695, 0, 250}, {60, 9863, 0, 0, 925, 0, 400},
	}
	var rank *healRank
	for i := range ranks {
		if ranks[i].level <= int(druid.Level) {
			rank = &ranks[i]
		}
	}
	if rank == nil {
		return
	}
	tick := rank.tick
	coefficient := 10.0 / 3.5 / 3 / 10

	druid.Tranquility = druid.RegisterSpell(Humanoid, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: rank.spellID},
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagChanneled | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost:   rank.mana,
			Multiplier: 100 - 3*druid.Talents.TranquilSpirit,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 2 * time.Minute,
			},
		},

		DamageMultiplier: druid.healMultiplier() * (1 + 0.2*float64(druid.Talents.TranquilSpirit)),
		ThreatMultiplier: 0.5 * (1 - 0.2*float64(druid.Talents.TranquilSpirit)),

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Tranquility",
			},
			NumberOfTicks: 10,
			TickLength:    time.Second,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, member := range partyOf(target) {
					dot.Spell.CalcAndDealHealing(sim, member, tick+coefficient*dot.Spell.HealingPower(member), dot.Spell.OutcomeHealingCrit)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
}

// Swiftmend (talent, 18562): consumes Rejuvenation or Regrowth on the target to heal for 12 sec of Rejuvenation
// (4 ticks) or 18 sec of Regrowth (6 ticks). The server data has no mana cost for it; the sim uses the classic 16%
// of base mana (unverified). Cooldown 10 sec (server data; classic 15 sec).
func (druid *Druid) registerSwiftmendSpell() {
	if !druid.Talents.Swiftmend {
		return
	}
	druid.Swiftmend = druid.RegisterSpell(Humanoid, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 18562},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.16,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 10 * time.Second,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return druid.activeHot(druid.Rejuvenation, target) != nil || druid.activeHot(druid.Regrowth, target) != nil
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 0.5,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			amount := 0.0
			if hot := druid.activeHot(druid.Rejuvenation, target); hot != nil {
				amount = hot.SnapshotBaseDamage * hot.SnapshotAttackerMultiplier * 4
				hot.Deactivate(sim)
			} else if hot := druid.activeHot(druid.Regrowth, target); hot != nil {
				amount = hot.SnapshotBaseDamage * hot.SnapshotAttackerMultiplier * 6
				hot.Deactivate(sim)
			}
			spell.CalcAndDealHealing(sim, target, amount, spell.OutcomeHealingCrit)
		},
	})
}

func (druid *Druid) activeHot(spells []*DruidSpell, target *core.Unit) *core.Dot {
	for _, spell := range spells {
		if spell != nil {
			if hot := spell.Hot(target); hot != nil && hot.IsActive() {
				return hot
			}
		}
	}
	return nil
}

// Nature's Swiftness (talent, 17116): the next Nature spell with a cast time is instant, 3 min cooldown.
func (druid *Druid) registerNaturesSwiftnessCD() {
	if !druid.Talents.NaturesSwiftness {
		return
	}
	actionID := core.ActionID{SpellID: 17116}
	var nsSpell *DruidSpell
	var affected []*core.Spell

	nsAura := druid.RegisterAura(core.Aura{
		Label:    "Nature's Swiftness",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affected = core.FilterSlice(druid.Spellbook, func(spell *core.Spell) bool {
				return spell.SpellSchool.Matches(core.SpellSchoolNature) && spell.DefaultCast.CastTime > 0
			})
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affected, func(spell *core.Spell) { spell.CastTimeMultiplier -= 1 })
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affected, func(spell *core.Spell) { spell.CastTimeMultiplier += 1 })
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolNature) && spell.DefaultCast.CastTime > 0 {
				aura.Deactivate(sim)
				nsSpell.CD.Use(sim)
				druid.UpdateMajorCooldowns()
			}
		},
	})

	nsSpell = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 3 * time.Minute,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			nsAura.Activate(sim)
		},
	})
}
