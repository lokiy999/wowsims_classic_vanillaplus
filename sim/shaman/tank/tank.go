package tank

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

func RegisterTankShaman() {
	core.RegisterAgentFactory(
		proto.Player_TankShaman{},
		proto.Spec_SpecTankShaman,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewTankShaman(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_TankShaman)
			if !ok {
				panic("Invalid spec value for Tank Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

type TankShaman struct {
	*shaman.Shaman

	Options *proto.TankShaman_Options
}

func NewTankShaman(character *core.Character, options *proto.Player) *TankShaman {
	tank := &TankShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString),
	}

	// Enable Auto Attacks for this spec
	tank.EnableAutoAttacks(tank, core.AutoAttackOptions{
		MainHand:       tank.WeaponFromMainHand(),
		OffHand:        tank.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})

	return tank
}

func (tank *TankShaman) GetShaman() *shaman.Shaman {
	return tank.Shaman
}

func (tank *TankShaman) Initialize() {
	tank.Shaman.Initialize()
}

func (tank *TankShaman) Reset(sim *core.Simulation) {
	tank.Shaman.Reset(sim)
	tank.Shaman.PseudoStats.Stunned = false
}
