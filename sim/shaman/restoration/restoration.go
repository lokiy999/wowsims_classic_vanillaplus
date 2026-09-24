package restoration

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

func RegisterRestorationShaman() {
	core.RegisterAgentFactory(
		proto.Player_RestorationShaman{},
		proto.Spec_SpecRestorationShaman,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewRestorationShaman(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_RestorationShaman)
			if !ok {
				panic("Invalid spec value for Restoration Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewRestorationShaman(character *core.Character, options *proto.Player) *RestorationShaman {
	return &RestorationShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString),
	}
}

type RestorationShaman struct {
	*shaman.Shaman
}

func (resto *RestorationShaman) GetShaman() *shaman.Shaman {
	return resto.Shaman
}

func (resto *RestorationShaman) Reset(sim *core.Simulation) {
	resto.Shaman.Reset(sim)
}

// Heals go to the first target dummy (the healing sim's stand-in for a raid member), or to the shaman.
func (resto *RestorationShaman) GetMainTarget() *core.Unit {
	target := resto.Env.Raid.GetFirstTargetDummy()
	if target == nil {
		return &resto.Unit
	}
	return &target.Unit
}

func (resto *RestorationShaman) Initialize() {
	resto.CurrentTarget = resto.GetMainTarget()
	resto.Shaman.Initialize()
	resto.Shaman.RegisterHealingSpells()
}
