package sim

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The custom Vanilla+ enchants use the game's SpellItemEnchantment ids (Part IG); the old 900xxx ids still load.
func TestLegacyEnchantIDs(t *testing.T) {
	for old, id := range core.LegacyEnchantIDs {
		if _, ok := core.EnchantsByEffectID[id]; !ok {
			t.Errorf("enchant %d (was %d) is not in the sim DB", id, old)
			continue
		}
		if _, ok := core.EnchantsByEffectID[old]; ok {
			t.Errorf("old id %d is still in the sim DB", old)
		}
		var itemID int32 // any item: the Go side doesn't check the slot
		for id := range core.ItemsByID {
			itemID = id
			break
		}
		for _, spec := range []int32{old, id} {
			item := core.NewItem(core.ItemSpec{ID: itemID, Enchant: spec})
			if item.Enchant.EffectID != id {
				t.Errorf("item %d with enchant %d: got enchant %d, want %d", itemID, spec, item.Enchant.EffectID, id)
			}
		}
	}
}

// Enchant Gloves - Arcane Power and Sigil of Speed: Boots (Part IH).
func TestCustomEnchantsIH(t *testing.T) {
	for id, want := range map[int32]stats.Stats{2723: {stats.ArcanePower: 20}, 2784: {stats.Stamina: 5}} {
		enchant, ok := core.EnchantsByEffectID[id]
		if !ok {
			t.Errorf("enchant %d is not in the sim DB", id)
			continue
		}
		if enchant.Stats != want {
			t.Errorf("enchant %d: stats %v, want %v", id, enchant.Stats, want)
		}
	}
}
