package warrior

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents added in the second audit round (see docs/CHANGES.md Part BB).

// Inner Rage (DBC): -25%/rank cooldown of Bloodrage, Berserker Rage and Recklessness.
func (warrior *Warrior) innerRageFactor() float64 {
	return 1 - 0.25*float64(warrior.Talents.InnerRage)
}

// bleedDamageMultiplier is the Two-Handed Weapon Specialization bonus to Bleed effects (+5%/rank, with a two-hander).
func (warrior *Warrior) bleedDamageMultiplier() float64 {
	if warrior.MainHand().HandType != proto.HandType_HandTypeTwoHand {
		return 1
	}
	return 1 + 0.05*float64(warrior.Talents.TwoHandedWeaponSpecialization)
}

func (warrior *Warrior) applyAuditTalents() {
	// Constitution (DBC 34127): +Strength, Agility and Spirit equal to 1% of the current health (taken as the maximum health).
	if warrior.Talents.Constitution {
		// Health is not a base stat, so a stat dependency is not allowed; apply it when the fight starts (not in the sidebar).
		warrior.RegisterResetEffect(func(sim *core.Simulation) {
			bonus := 0.01 * warrior.MaxHealth()
			warrior.AddStatsDynamic(sim, stats.Stats{stats.Strength: bonus, stats.Agility: bonus, stats.Spirit: bonus})
		})
	}

	// Butterfly Style (DBC): +2%/rank dodge and crit. The rage on dodge or parry part is not modeled.
	if warrior.Talents.ButterflyStyle > 0 {
		bonus := 2 * float64(warrior.Talents.ButterflyStyle)
		warrior.AddStat(stats.Dodge, bonus*core.DodgeRatingPerDodgeChance)
		warrior.AddStat(stats.MeleeCrit, bonus*core.CritRatingPerCritChance)
	}
}
