package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Returns the time to wait before the next action, or 0 if innervate is on CD
// or disabled.
func (druid *Druid) registerInnervateCD() {
	innervateTarget := druid.GetUnit(druid.SelfBuffs.InnervateTarget)
	if innervateTarget == nil {
		return
	}
	innervateTargetChar := druid.Env.Raid.GetPlayerFromUnit(innervateTarget).GetCharacter()

	actionID := core.ActionID{SpellID: 29166, Tag: druid.Index}

	innervateCD := core.InnervateCD

	innervateAura := core.InnervateAura(innervateTargetChar, actionID.Tag)

	// Catharsis (33790-33792): an Innervate on someone else also gives the druid 10/20/30 sec of Innervate.
	var catharsisAura *core.Aura
	if druid.Talents.Catharsis > 0 && innervateTarget != &druid.Unit {
		catharsisAura = druid.RegisterAura(core.Aura{
			Label:    "Catharsis Innervate",
			ActionID: core.ActionID{SpellID: []int32{0, 34242, 34243, 34244}[druid.Talents.Catharsis]},
			Duration: time.Second * 10 * time.Duration(druid.Talents.Catharsis),
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				druid.PseudoStats.SpiritRegenMultiplier += 3 // like Innervate: +300%
				druid.PseudoStats.ForceFullSpiritRegen = true
				druid.UpdateManaRegenRates()
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				druid.PseudoStats.SpiritRegenMultiplier -= 3
				druid.PseudoStats.ForceFullSpiritRegen = false
				druid.UpdateManaRegenRates()
			},
		})
	}
	var innervateManaThreshold float64

	druid.RegisterResetEffect(func(sim *core.Simulation) {
		if innervateTarget == &druid.Unit {
			if druid.StartingForm.Matches(Cat) {
				// double shift + innervate cost.
				// Prevents not having enough mana to shift back into form if more powershift are executed
				innervateManaThreshold = druid.CatForm.DefaultCast.Cost*2 + druid.Innervate.DefaultCast.Cost
			} else {
				// Threshold can be lower when casting on self because its never mid-cast.
				innervateManaThreshold = 500
			}
		} else {
			innervateManaThreshold = core.InnervateManaThreshold(innervateTargetChar)
		}
	})

	druid.Innervate = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID: actionID,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.05,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: innervateCD,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// If target already has another innervate, don't cast.
			return !innervateTarget.HasActiveAuraWithTag(core.InnervateAuraTag)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			innervateAura.Activate(sim)
			if catharsisAura != nil {
				catharsisAura.Activate(sim)
			}
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Innervate.Spell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Innervate needs to be activated as late as possible to maximize DPS. The issue is that
			// innervate gives so much mana that it can cause Super Mana Potion or Dark Rune usages
			// to be delayed, if they come off CD soon after innervate. This delay is minimized by
			// activating innervate from the smallest amount of mana possible.
			return innervateTarget.CurrentMana() <= innervateManaThreshold && !druid.form.Matches(Cat|Bear)
		},
	})
}
