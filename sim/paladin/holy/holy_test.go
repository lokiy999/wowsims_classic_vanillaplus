package holy

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterHolyPaladin()
}

func TestHolyPaladin(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:    proto.Class_ClassPaladin,
			Phase:    1,
			Race:     proto.Race_RaceHuman,
			IsHealer: true,

			Talents:     StandardTalents,
			GearSet:     core.GetGearSet("../../../ui/holy_paladin/gear_sets", "placeholder"),
			Rotation:    core.GetAplRotation("../../../ui/holy_paladin/apls", "default"),
			Buffs:       core.FullBuffs,
			Consumes:    FullConsumes,
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: PlayerOptionsStandard},

			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeOffHand,
					proto.WeaponType_WeaponTypeShield,
				},
				ArmorType: proto.ArmorType_ArmorTypePlate,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeLibram,
				},
			},
			EPReferenceStat: proto.Stat_StatSpellPower,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatIntellect,
				proto.Stat_StatSpirit,
				proto.Stat_StatSpellPower,
				proto.Stat_StatSpellCrit,
				proto.Stat_StatMP5,
			},
		},
	}))
}

var StandardTalents = ""

var PlayerOptionsStandard = &proto.Player_HolyPaladin{
	HolyPaladin: &proto.HolyPaladin{
		Options: &proto.PaladinOptions{},
	},
}

var FullConsumes = core.ConsumesCombo{
	Label: "Consumes",
	Consumes: &proto.Consumes{
		Flask:         proto.Flask_FlaskOfDistilledWisdom,
		DefaultPotion: proto.Potions_MajorManaPotion,
	},
}
