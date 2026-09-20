package druid

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// Every talent in the UI tree file must decode to the proto field with the same name, since a talent
// string is read by position (tab order and order inside each tab).
func TestTalentTreeMatchesProto(t *testing.T) {
	raw, err := os.ReadFile("../../ui/core/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var tabs []struct {
		Name    string `json:"name"`
		Talents []struct {
			FieldName string `json:"fieldName"`
			MaxPoints int    `json:"maxPoints"`
		} `json:"talents"`
	}
	if err := json.Unmarshal(raw, &tabs); err != nil {
		t.Fatal(err)
	}

	for tabIdx, tab := range tabs {
		for talentIdx, talent := range tab.Talents {
			digits := make([]string, len(tabs))
			digits[tabIdx] = strings.Repeat("0", talentIdx) + "1"
			talentsString := strings.Join(digits, "-")

			talents := &proto.DruidTalents{}
			core.FillTalentsProto(talents.ProtoReflect(), talentsString, TalentTreeSizes)

			msg := talents.ProtoReflect()
			fd := msg.Descriptor().Fields().ByJSONName(talent.FieldName)
			if fd == nil {
				t.Errorf("%s: no proto field %q", tab.Name, talent.FieldName)
				continue
			}
			var set bool
			if fd.Kind() == protoreflect.BoolKind {
				set = msg.Get(fd).Bool()
			} else {
				set = msg.Get(fd).Int() == 1
			}
			if !set {
				t.Errorf("%s talent #%d (%s) does not decode to its proto field", tab.Name, talentIdx, talent.FieldName)
			}
		}
	}
}
