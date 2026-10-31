package sim

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// Trinket and item effects with the server's values (Part HY): cooldowns from the item cache, stacks and durations
// from the spell data.

var trinketTestSpecs = map[proto.Class]func(p *proto.Player){
	proto.Class_ClassWarrior: func(p *proto.Player) {
		p.Race = proto.Race_RaceHuman
		p.Spec = &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}}
	},
	proto.Class_ClassMage: func(p *proto.Player) {
		p.Race = proto.Race_RaceGnome
		p.Spec = &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}}
	},
	proto.Class_ClassRogue: func(p *proto.Player) {
		p.Race = proto.Race_RaceHuman
		p.Spec = &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}}
	},
	proto.Class_ClassHunter: func(p *proto.Player) {
		p.Race = proto.Race_RaceDwarf
		p.Spec = &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{}}}
	},
	proto.Class_ClassShaman: func(p *proto.Player) {
		p.Race = proto.Race_RaceOrc
		p.Spec = &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{Options: &proto.ElementalShaman_Options{}}}
	},
	proto.Class_ClassWarlock: func(p *proto.Player) {
		p.Race = proto.Race_RaceHuman
		p.Spec = &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}}
	},
	proto.Class_ClassPaladin: func(p *proto.Player) {
		p.Race = proto.Race_RaceHuman
		p.Spec = &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{Options: &proto.PaladinOptions{}}}
	},
}

// newTrinketTestSim makes a level 60 character of `class` wearing `items` (slot -> item id), against one target of
// `mobType`, and resets the sim so auras and spells can be used directly.
func newTrinketTestSim(t *testing.T, class proto.Class, items map[proto.ItemSlot]int32, mobType proto.MobType) (*core.Simulation, *core.Character) {
	t.Helper()
	for _, id := range items {
		if _, ok := core.ItemsByID[id]; !ok {
			t.Skipf("item %d is not in the sim DB", id)
		}
	}
	p := &proto.Player{Name: "p", Class: class, Equipment: &proto.EquipmentSpec{Items: make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)}}
	for i := range p.Equipment.Items {
		p.Equipment.Items[i] = &proto.ItemSpec{}
	}
	for slot, id := range items {
		p.Equipment.Items[slot] = &proto.ItemSpec{Id: id}
	}
	trinketTestSpecs[class](p)
	p.Rotation = &proto.APLRotation{}
	encounter := &proto.Encounter{Duration: 300, Targets: []*proto.Target{{
		Stats:   stats.Stats{stats.Armor: 3000}.ToFloatArray(),
		MobType: mobType,
		Level:   63,
	}}}
	rsr := &proto.RaidSimRequest{
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{p}}}},
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Iterations: 1, IsTest: true, RandomSeed: 101},
	}
	sim := core.NewSim(rsr, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func trinketSpell(t *testing.T, char *core.Character, actionID core.ActionID) *core.Spell {
	t.Helper()
	spell := char.GetSpell(actionID)
	if spell == nil {
		t.Fatalf("%s: no spell %s", char.Label, actionID)
	}
	return spell
}

func trinketAura(t *testing.T, unit *core.Unit, label string) *core.Aura {
	t.Helper()
	aura := unit.GetAura(label)
	if aura == nil {
		t.Fatalf("no aura %q", label)
	}
	return aura
}

func TestTrinketCooldownsFromServer(t *testing.T) {
	trinket := proto.ItemSlot_ItemSlotTrinket1
	for _, tc := range []struct {
		item     int32
		class    proto.Class
		actionID core.ActionID
		cd       time.Duration
	}{
		{19930, proto.Class_ClassMage, core.ActionID{ItemID: 19930}, time.Minute * 2},     // Mar'li's Eye
		{19947, proto.Class_ClassMage, core.ActionID{ItemID: 19947}, time.Minute},         // Nat Pagle's Broken Reel
		{22268, proto.Class_ClassMage, core.ActionID{ItemID: 22268}, time.Second * 75},    // Draconic Infused Emblem
		{20036, proto.Class_ClassMage, core.ActionID{ItemID: 20036}, time.Minute * 2},     // Fire Ruby
		{19959, proto.Class_ClassMage, core.ActionID{ItemID: 19959}, time.Minute * 2},     // Hazza'rah's Charm of Magic
		{20512, proto.Class_ClassPaladin, core.ActionID{ItemID: 20512}, time.Minute * 6},  // Sanctified Orb
		{19951, proto.Class_ClassWarrior, core.ActionID{ItemID: 19951}, time.Minute * 2},  // Gri'lek's Charm of Might
		{19953, proto.Class_ClassHunter, core.ActionID{ItemID: 19953}, time.Minute * 2},   // Renataki's Charm of Beasts
		{19954, proto.Class_ClassRogue, core.ActionID{ItemID: 19954}, time.Minute * 2},    // Renataki's Charm of Trickery
		{19342, proto.Class_ClassRogue, core.ActionID{ItemID: 19342}, time.Minute * 5},    // Venomous Totem
		{19956, proto.Class_ClassShaman, core.ActionID{ItemID: 19956}, time.Minute * 2},   // Wushoolay's Charm of Spirits
		{19957, proto.Class_ClassWarlock, core.ActionID{ItemID: 19957}, time.Minute * 2},  // Hazza'rah's Charm of Destruction
		{19336, proto.Class_ClassHunter, core.ActionID{ItemID: 19336}, time.Minute * 2},   // Arcane Infused Gem
		{19343, proto.Class_ClassPaladin, core.ActionID{ItemID: 19343}, time.Minute * 5},  // Scrolls of Blinding Light
		{80006, proto.Class_ClassRogue, core.ActionID{ItemID: 80006}, time.Minute * 5},    // Cuergo's Trusty Flask
		{26386, proto.Class_ClassWarrior, core.ActionID{SpellID: 36231}, time.Minute * 3}, // Holy Water Flask
		{20503, proto.Class_ClassMage, core.ActionID{ItemID: 20503}, time.Minute * 3},     // Enamored Water Spirit
		{26313, proto.Class_ClassMage, core.ActionID{ItemID: 26313}, time.Second * 80},    // Mark of Primal Vitality
		{19024, proto.Class_ClassWarrior, core.ActionID{ItemID: 19024}, time.Minute * 30}, // Arena Grand Master
		{20071, proto.Class_ClassWarrior, core.ActionID{ItemID: 20071}, time.Minute * 3},  // Talisman of Arathor
		{21120, proto.Class_ClassWarrior, core.ActionID{ItemID: 21120}, time.Minute * 3},  // Defiler's Talisman
		{26224, proto.Class_ClassWarrior, core.ActionID{ItemID: 26224}, time.Minute * 2},  // Personal Harm Prevention Field Emitter
		{26365, proto.Class_ClassMage, core.ActionID{ItemID: 26365}, time.Minute * 2},     // Leyguard Charm
	} {
		t.Run(tc.actionID.String(), func(t *testing.T) {
			_, char := newTrinketTestSim(t, tc.class, map[proto.ItemSlot]int32{trinket: tc.item}, proto.MobType_MobTypeHumanoid)
			if got := trinketSpell(t, char, tc.actionID).CD.Duration; got != tc.cd {
				t.Errorf("item %d: cooldown %s, want %s", tc.item, got, tc.cd)
			}
		})
	}
}

// Hookfang Shanker (Corrosive Poison 13526) and The Cruel Hand of Timmy (Curse of Timmy 17505) stack up to 3 times.
func TestWeaponProcStacks(t *testing.T) {
	sim, char := newTrinketTestSim(t, proto.Class_ClassWarrior, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 11635}, proto.MobType_MobTypeHumanoid)
	target := sim.Encounter.TargetUnits[0]
	armor := target.GetStat(stats.Armor)
	poison := trinketSpell(t, char, core.ActionID{SpellID: 13526})
	for i := 0; i < 30; i++ {
		poison.Cast(sim, target)
	}
	if stacks := poison.Dot(target).GetStacks(); stacks != 3 {
		t.Errorf("Corrosive Poison: %d stacks, want 3", stacks)
	}
	if got := armor - target.GetStat(stats.Armor); got != 150 {
		t.Errorf("Corrosive Poison: armor -%v, want -150", got)
	}
	if got := poison.Dot(target).SnapshotBaseDamage; got != 30 {
		t.Errorf("Corrosive Poison: %v damage per tick before modifiers, want 30", got)
	}

	sim, _ = newTrinketTestSim(t, proto.Class_ClassWarrior, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 13401}, proto.MobType_MobTypeHumanoid)
	target = sim.Encounter.TargetUnits[0]
	strength := target.GetStat(stats.Strength)
	curse := trinketAura(t, target, "Curse of Timmy")
	for i := 0; i < 4; i++ {
		curse.Activate(sim)
		curse.AddStack(sim)
	}
	if curse.GetStacks() != 3 || strength-target.GetStat(stats.Strength) != 45 {
		t.Errorf("Curse of Timmy: %d stacks, strength -%v; want 3 and -45", curse.GetStacks(), strength-target.GetStat(stats.Strength))
	}
	curse.Deactivate(sim)
	if target.GetStat(stats.Strength) != strength {
		t.Errorf("Curse of Timmy: strength %v after it ends, want %v", target.GetStat(stats.Strength), strength)
	}
}

// Venomous Totem: poisons are guaranteed to apply for 30 sec.
func TestVenomousTotem(t *testing.T) {
	_, char := newTrinketTestSim(t, proto.Class_ClassRogue, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 19342}, proto.MobType_MobTypeHumanoid)
	if d := trinketAura(t, &char.Unit, "Venomous Totem").Duration; d != time.Second*30 {
		t.Errorf("Venomous Totem lasts %s, want 30s", d)
	}
}

// Wushoolay's Charm of Spirits: the next spell is a sure crit (30 sec, one charge).
func TestWushoolaysCharmOfSpirits(t *testing.T) {
	sim, char := newTrinketTestSim(t, proto.Class_ClassShaman, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 19956}, proto.MobType_MobTypeHumanoid)
	aura := trinketAura(t, &char.Unit, "Wushoolay's Charm of Spirits")
	if aura.Duration != time.Second*30 {
		t.Errorf("lasts %s, want 30s", aura.Duration)
	}
	var bolt *core.Spell
	for _, spell := range char.Spellbook {
		if spell.Flags.Matches(core.SpellFlagAPL) && spell.ProcMask.Matches(core.ProcMaskSpellDamage) && spell.SpellSchool == core.SpellSchoolNature && spell.DefaultCast.CastTime > 0 {
			bolt = spell
		}
	}
	if bolt == nil {
		t.Fatal("no Nature spell with a cast time (Lightning Bolt) in the shaman's spellbook")
	}
	before := bolt.BonusCritRating
	aura.Activate(sim)
	if got := bolt.BonusCritRating - before; got != 100*core.SpellCritRatingPerCritChance {
		t.Errorf("%s: +%v crit rating while active, want +%v", bolt.ActionID, got, 100*core.SpellCritRatingPerCritChance)
	}
	aura.OnCastComplete(aura, sim, bolt)
	if aura.IsActive() || bolt.BonusCritRating != before {
		t.Errorf("the charge isn't used up by %s (active %v, crit rating %v -> %v)", bolt.ActionID, aura.IsActive(), before, bolt.BonusCritRating)
	}
}

// Holy Water Flask: 222-263 Holy damage, triple against Undead and Demons.
func TestHolyWaterFlask(t *testing.T) {
	for _, tc := range []struct {
		mobType  proto.MobType
		min, max float64
	}{
		{proto.MobType_MobTypeHumanoid, 222, 263 * 2},
		{proto.MobType_MobTypeUndead, 3 * 222, 3 * 263 * 2},
	} {
		sim, char := newTrinketTestSim(t, proto.Class_ClassWarrior, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 26386}, tc.mobType)
		spell := trinketSpell(t, char, core.ActionID{SpellID: 36231})
		for i := 0; i < 20; i++ {
			spell.CD.Reset()
			spell.Cast(sim, sim.Encounter.TargetUnits[0])
		}
		m := spell.SpellMetrics[0]
		if m.Hits == 0 {
			t.Fatalf("%s: no hits in 20 casts", tc.mobType)
		}
		// partial resists can lower a hit, so only the average is checked: it is between min and max (crits x2 at most)
		if avg := m.TotalDamage / float64(m.Hits+m.Crits); avg < tc.min*0.75 || avg > tc.max {
			t.Errorf("%s: average %0.1f per hit, want %v-%v", tc.mobType, avg, tc.min, tc.max)
		}
	}
}

// Absorb shields: Talisman of Arathor absorbs 495-605 damage; Leyguard Charm absorbs up to 1200 for 0.5 mana each.
func TestAbsorbTrinkets(t *testing.T) {
	sim, char := newTrinketTestSim(t, proto.Class_ClassWarrior, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 20071}, proto.MobType_MobTypeHumanoid)
	shield := trinketAura(t, &char.Unit, "Damage Absorb")
	shield.Activate(sim)
	result := &core.SpellResult{Damage: 1000}
	for _, mod := range char.DynamicDamageTakenModifiers {
		mod(sim, nil, result)
	}
	if result.Damage < 1000-605 || result.Damage > 1000-495 || shield.IsActive() {
		t.Errorf("Talisman of Arathor: 1000 damage -> %v (active %v), want 395-505 and the shield gone", result.Damage, shield.IsActive())
	}

	sim, char = newTrinketTestSim(t, proto.Class_ClassMage, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 26365}, proto.MobType_MobTypeHumanoid)
	shield = trinketAura(t, &char.Unit, "Leyguard")
	shield.Activate(sim)
	mana := char.CurrentMana()
	result = &core.SpellResult{Damage: 400}
	for _, mod := range char.DynamicDamageTakenModifiers {
		mod(sim, nil, result)
	}
	if result.Damage != 0 || mana-char.CurrentMana() != 200 || !shield.IsActive() {
		t.Errorf("Leyguard Charm: 400 damage -> %v, mana -%v (active %v), want 0, -200 and still active", result.Damage, mana-char.CurrentMana(), shield.IsActive())
	}
}
