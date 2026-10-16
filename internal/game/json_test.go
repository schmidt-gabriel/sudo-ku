package game

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestJSONRoundTrip(t *testing.T) {
	g := newGame()
	g.Set(2, 4)
	g.ToggleNote(3, 1)
	g.Tick(65 * time.Second)

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var got Game
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Givens != g.Givens || got.Values != g.Values || got.Notes != g.Notes ||
		got.Solution != g.Solution || got.Difficulty != g.Difficulty || got.Elapsed != g.Elapsed {
		t.Errorf("round trip changed the game:\n got %+v\nwant %+v", got, *g)
	}
	if got.Undo() {
		t.Error("undo history should not be saved")
	}
}

func TestUnmarshalRejectsInvalidData(t *testing.T) {
	cases := map[string]func(m map[string]any){
		"short givens":      func(m map[string]any) { m["givens"] = "123" },
		"bad characters":    func(m map[string]any) { m["values"] = strings.Repeat("x", 81) },
		"given overwritten": func(m map[string]any) { m["values"] = "9" + m["values"].(string)[1:] },
		"unsolved solution": func(m map[string]any) { m["solution"] = m["givens"] },
		"bad difficulty":    func(m map[string]any) { m["difficulty"] = 9 },
		"bad notes":         func(m map[string]any) { m["notes"].([]any)[0] = 1 },
		"negative time":     func(m map[string]any) { m["elapsedMs"] = -1 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			data, _ := json.Marshal(newGame())
			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatal(err)
			}
			mutate(m)
			data, _ = json.Marshal(m)
			var g Game
			if err := json.Unmarshal(data, &g); err == nil {
				t.Error("invalid save was accepted")
			}
		})
	}
	for _, raw := range []string{"{not json", "{}"} {
		var g Game
		if err := json.Unmarshal([]byte(raw), &g); err == nil {
			t.Errorf("%q was accepted", raw)
		}
	}
}
