package shaman

import (
	"math"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const RockbiterWeaponRanks = 7

var RockbiterWeaponEnchantId = [RockbiterWeaponRanks + 1]int32{0, 29, 6, 1, 503, 1663, 683, 1664}
// Server data: Rockbiter Weapon gives Strength and healing, not attack power.
var RockbiterWeaponBonusStrength = [RockbiterWeaponRanks + 1]float64{0, 15, 30, 40, 60, 110, 200, 280}
var RockbiterWeaponBonusHealing = [RockbiterWeaponRanks + 1]float64{0, 7, 15, 22, 33, 51, 78, 100} // Spell.csv 34118-34124
var RockbiterWeaponBonusTPS = [RockbiterWeaponRanks + 1]float64{0, 6, 10, 16, 27, 41, 55, 72}
var RockbiterWeaponLevel = [RockbiterWeaponRanks + 1]int32{0, 1, 8, 16, 24, 34, 44, 54}

var RockbiterWeaponRankByLevel = map[int32]int32{
	25: 4,
	40: 5,
	50: 6,
	60: 7,
}

func (shaman *Shaman) RegisterRockbiterImbue(procMask core.ProcMask) {
	if procMask == core.ProcMaskUnknown {
		return
	}

	rank := RockbiterWeaponRankByLevel[shaman.Level]
	enchantId := RockbiterWeaponEnchantId[rank]
	bonusThreat := RockbiterWeaponBonusTPS[rank]

	duration := time.Minute * 5

	hasMHImbue := procMask.Matches(core.ProcMaskMeleeMH)
	hasOHImbue := procMask.Matches(core.ProcMaskMeleeOH)

	if hasMHImbue {
		shaman.MainHand().TempEnchant = enchantId
		shaman.AutoAttacks.MHConfig().FlatThreatBonus += bonusThreat * shaman.AutoAttacks.MH().SwingSpeed
	}
	if hasOHImbue {
		shaman.OffHand().TempEnchant = enchantId
		shaman.AutoAttacks.MHConfig().FlatThreatBonus += bonusThreat * shaman.AutoAttacks.OH().SwingSpeed
	}

	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(procMask) {
			spell.FlatThreatBonus += bonusThreat
		}
	})

	aura := shaman.RegisterAura(core.Aura{
		Label:    "Rockbiter Imbue",
		Duration: duration,
	})

	shaman.RegisterOnItemSwapWithImbue(enchantId, &procMask, aura)
}

func (shaman *Shaman) ApplyRockbiterImbue(procMask core.ProcMask) {
	if procMask.Matches(core.ProcMaskMeleeMH) && shaman.HasMHWeapon() {
		shaman.ApplyRockbiterImbueToItem(shaman.MainHand())
	}

	if procMask.Matches(core.ProcMaskMeleeOH) && shaman.HasOHWeapon() {
		shaman.ApplyRockbiterImbueToItem(shaman.OffHand())
	}
}

func (shaman *Shaman) ApplyRockbiterImbueToItem(item *core.Item) {
	if item == nil {
		return
	}

	rank := RockbiterWeaponRankByLevel[shaman.Level]
	enchantId := RockbiterWeaponEnchantId[rank]

	effect := []float64{1, 1.10, 1.20, 1.30}[shaman.Talents.ElementalWeapons] * (1 + shaman.ElementalWeaponEnchantEffectivenessBonus)

	newStats := stats.Stats{
		stats.Strength:     math.Floor(RockbiterWeaponBonusStrength[rank] * effect),
		stats.HealingPower: math.Floor(RockbiterWeaponBonusHealing[rank] * effect),
	}

	item.Stats = item.Stats.Add(newStats)
	item.TempEnchant = enchantId
}
