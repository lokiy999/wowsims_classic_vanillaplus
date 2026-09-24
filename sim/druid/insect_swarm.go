package druid

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const InsectSwarmRanks = 5

var InsectSwarmSpellId = [InsectSwarmRanks + 1]int32{0, 5570, 24974, 24975, 24976, 24977}
var InsectSwarmBaseDamage = [InsectSwarmRanks + 1]float64{0, 80, 160, 240, 320, 400} // server: over 16 sec
var InsectSwarmManaCost = [InsectSwarmRanks + 1]float64{0, 45, 85, 100, 140, 160}
var InsectSwarmLevel = [InsectSwarmRanks + 1]int{0, 20, 30, 40, 50, 60}

func (druid *Druid) registerInsectSwarmSpell() {
	druid.InsectSwarm = make([]*DruidSpell, InsectSwarmRanks+1)

	druid.InsectSwarmAuras = druid.NewEnemyAuraArray(core.InsectSwarmAura)
	cdTimer := druid.NewTimer() // shared by all ranks

	for rank := 1; rank <= InsectSwarmRanks; rank++ {
		level := InsectSwarmLevel[rank]
		if int32(level) <= druid.Level {
			// Power of Nature (DBC 33736/33737): +25%/50% duration of Insect Swarm.
			// Server: 8 ticks every 2 sec (16 sec). Power of Nature adds ticks (and their damage).
			baseTicks := 8.0
			numTicks := int32(math.Round(baseTicks * (1 + 0.25*float64(druid.Talents.PowerOfNature))))
			tickLength := time.Second * 2

			spellID := InsectSwarmSpellId[rank]
			baseDamage := InsectSwarmBaseDamage[rank] / baseTicks
			manaCost := InsectSwarmManaCost[rank]
			spellCoef := .158 * 6 / baseTicks // classic total, spread over the server ticks

			druid.InsectSwarm[rank] = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
				SpellCode:   SpellCode_DruidInsectSwarm,
				ActionID:    core.ActionID{SpellID: spellID},
				SpellSchool: core.SpellSchoolNature,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskSpellDamage,
				Flags:       SpellFlagOmen | core.SpellFlagAPL | core.SpellFlagBinary,

				ManaCost: core.ManaCostOptions{
					FlatCost: manaCost,
				},
				Cast: core.CastConfig{
					DefaultCast: core.Cast{
						GCD: core.GCDDefault,
					},
					CD: core.Cooldown{
						Timer:    cdTimer,
						Duration: time.Second * 8, // DBC: 8s cooldown on every rank
					},
				},

				DamageMultiplier: 1,
				ThreatMultiplier: 1,

				Dot: core.DotConfig{
					Aura: core.Aura{
						Label: fmt.Sprintf("Insect Swarm (Rank %d)", rank),
						OnGain: func(aura *core.Aura, sim *core.Simulation) {
							druid.InsectSwarmAuras.Get(aura.Unit).Activate(sim)
						},
						OnExpire: func(aura *core.Aura, sim *core.Simulation) {
							insectSwarmAura := druid.InsectSwarmAuras.Get(aura.Unit)
							if !insectSwarmAura.IsPermanent() {
								insectSwarmAura.Deactivate(sim)
							}
						},
					},

					NumberOfTicks:    numTicks,
					TickLength:       tickLength,
					BonusCoefficient: spellCoef,

					OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
						dot.Snapshot(target, baseDamage, isRollover)
						if !druid.form.Matches(Moonkin) {
							dot.SnapshotCritChance = 0
						}
					},
					OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
						dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)
					},
				},

				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
					if result.Landed() {
						spell.Dot(target).Apply(sim)
					}
					spell.DealOutcome(sim, result)
				},

				RelatedAuras: []core.AuraArray{druid.InsectSwarmAuras},
			})
		}
	}
}
