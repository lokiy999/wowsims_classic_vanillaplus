package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Death and Decay (talent, server spell 33926): instant, 5 min cooldown, 50% of base mana; for 30 sec, enemies in a
// 20 yd area take Shadow damage equal to the warlock's level every 2 sec (and 20% less healing). The damage is per
// level with no spell power coefficient in the data: none is assumed (confirm in game).
func (warlock *Warlock) registerDeathAndDecaySpell() {
	if !warlock.Talents.DeathAndDecay {
		return
	}

	tickDamage := float64(warlock.Level)

	warlock.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 33926},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagAffliction,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.50,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Minute * 5,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 0, // server: Generates No Threat

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "DeathAndDecay-" + warlock.Label,
			},
			NumberOfTicks: 15,
			TickLength:    time.Second * 2,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, tickDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})
}
