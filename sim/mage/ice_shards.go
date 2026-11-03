package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Ice Shards (talent, server spell 35714): instant, 10 sec cooldown, 28% of base mana; launches four shards of ice,
// each dealing Frost damage equal to the mage's level. The damage is a server script (a Dummy effect), so the data has
// no spell power coefficient: none is assumed (confirm in game). Each shard rolls hit and crit on its own.
// Cold Grip (11185 / 12487): +15%/rank damage and -25%/rank cooldown.
func (mage *Mage) registerIceShardsSpell() {
	if !mage.Talents.IceShards {
		return
	}

	const numShards = 4
	coldGrip := float64(mage.Talents.ColdGrip)
	shardDamage := float64(mage.Level)

	mage.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 35714},
		SpellSchool: core.SpellSchoolFrost,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.28,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Duration(float64(time.Second*10) * (1 - 0.25*coldGrip)),
			},
		},

		DamageMultiplier: 1 + 0.15*coldGrip,
		ThreatMultiplier: 0, // server: Generates No Threat

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			result := spell.CalcDamage(sim, target, shardDamage, spell.OutcomeExpectedMagicHitAndCrit)
			result.Damage *= numShards
			return result
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for i := 0; i < numShards; i++ {
				spell.CalcAndDealDamage(sim, target, shardDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	})
}
