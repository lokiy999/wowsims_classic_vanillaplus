package paladin

import (
	"github.com/wowsims/classic/sim/core"
)

func (paladin *Paladin) registerRighteousFury() {
	if !paladin.Options.RighteousFury {
		return
	}
	actionID := core.ActionID{SpellID: 25780}

	// +30% threat, and Improved Righteous Fury adds +10% threat and +10% attack speed per rank (60% at 3/3), confirmed in game.
	rfThreatMultiplier := 1.3 + 0.1*float64(paladin.Talents.ImprovedRighteousFury)
	rfAttackSpeed := 1 + 0.1*float64(paladin.Talents.ImprovedRighteousFury)

	rfAura := paladin.RegisterAura(core.Aura{
		Label:    "Righteous Fury",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.ThreatMultiplier *= rfThreatMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.ThreatMultiplier /= rfThreatMultiplier
		},
	})
	if paladin.Talents.ImprovedRighteousFury > 0 {
		rfAura.AttachMultiplyAttackSpeed(&paladin.Unit, rfAttackSpeed)
	}
	core.MakePermanent(rfAura)
}
