package mage

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The UI sends talent strings in Arcane / Fire / Frost order.
func TestTalentsStringTreeOrder(t *testing.T) {
	talents := &proto.MageTalents{}
	// Second tree (Fire): Improved Fireball is its second talent.
	core.FillTalentsProto(talents.ProtoReflect(), reorderTalentsString("-05"), TalentTreeSizes)
	if talents.ImprovedFireball != 5 || talents.WandSpecialization != 0 {
		t.Fatalf("Fire tree talents misread: ImprovedFireball=%d", talents.ImprovedFireball)
	}

	talents = &proto.MageTalents{}
	// First tree (Arcane): Wand Specialization is its first talent.
	core.FillTalentsProto(talents.ProtoReflect(), reorderTalentsString("5"), TalentTreeSizes)
	if talents.WandSpecialization != 5 || talents.ImprovedFireball != 0 {
		t.Fatalf("Arcane tree talents misread: WandSpecialization=%d", talents.WandSpecialization)
	}
}
