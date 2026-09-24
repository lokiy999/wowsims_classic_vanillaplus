package holy

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/paladin"
)

func RegisterHolyPaladin() {
	core.RegisterAgentFactory(
		proto.Player_HolyPaladin{},
		proto.Spec_SpecHolyPaladin,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewHolyPaladin(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_HolyPaladin)
			if !ok {
				panic("Invalid spec value for Holy Paladin!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewHolyPaladin(character *core.Character, options *proto.Player) *HolyPaladin {
	holyOptions := options.GetHolyPaladin().Options
	if holyOptions == nil {
		holyOptions = &proto.PaladinOptions{}
	}
	return &HolyPaladin{
		Paladin: paladin.NewPaladin(character, options, holyOptions),
	}
}

type HolyPaladin struct {
	*paladin.Paladin
}

func (holy *HolyPaladin) GetPaladin() *paladin.Paladin {
	return holy.Paladin
}

// Heals go to the first target dummy (the healing sim's stand-in for a raid member), or to the paladin.
func (holy *HolyPaladin) GetMainTarget() *core.Unit {
	target := holy.Env.Raid.GetFirstTargetDummy()
	if target == nil {
		return &holy.Unit
	}
	return &target.Unit
}

func (holy *HolyPaladin) Initialize() {
	holy.CurrentTarget = holy.GetMainTarget()
	holy.Paladin.Initialize()
	holy.Paladin.RegisterHealingSpells()
}

func (holy *HolyPaladin) Reset(sim *core.Simulation) {
	holy.Paladin.Reset(sim)
}
