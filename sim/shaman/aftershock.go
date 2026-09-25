package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Aftershock (talent, Spell.csv 33648): 20 sec cooldown. Consumes your shock spells on the target; a consumed
// Flame Shock instantly deals damage equal to 15 sec of Flame Shock (the whole server DoT). The Frost Shock stun and
// the Earth Shock taunt do nothing in a DPS sim. The mana cost is a percentage of base mana that the spell export
// does not have, so it is free here (TODO question 34).
func (shaman *Shaman) registerAftershockSpell() {
	if !shaman.Talents.Aftershock {
		return
	}

	activeFlameShock := func(target *core.Unit) *core.Dot {
		for _, spell := range shaman.FlameShock {
			if spell != nil && spell.Dot(target).IsActive() {
				return spell.Dot(target)
			}
		}
		return nil
	}

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 33648},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		// The damage comes from the Flame Shock snapshot, which already has the caster's modifiers.
		Flags: core.SpellFlagAPL | SpellFlagShaman | core.SpellFlagIgnoreAttackerModifiers,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 20,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return activeFlameShock(target) != nil
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dot := activeFlameShock(target)
			if dot == nil {
				return
			}
			damage := dot.SnapshotBaseDamage * dot.SnapshotAttackerMultiplier * float64(dot.NumberOfTicks)
			dot.Deactivate(sim)
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
		},
	})
}
