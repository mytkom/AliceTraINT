package jalien

import "testing"

func TestAODMatcher_SimPaths(t *testing.T) {
	cases := []struct {
		path      string
		period    string
		runNumber uint64
		aodNumber uint64
	}{
		{"/alice/sim/2023/LHC23c1/302004/AOD/010/AO2D.root", "LHC23c1", 302004, 10},
		{"/alice/sim/2025/LHC25f3_gp/0/544116/AOD/001/AO2D.root", "LHC25f3_gp", 544116, 1},
		{"/alice/sim/2025/LHC25h1_0.2/0/544116/AOD/001/AO2D.root", "LHC25h1_0.2", 544116, 1},
	}

	matcher := newAODMatcher()
	for _, c := range cases {
		res, err := matcher.MatchAO2DPath(c.path)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.path, err)
		}
		if res == nil {
			t.Fatalf("%s: expected match, got nil", c.path)
		}
		if res.LHCPeriod != c.period || res.RunNumber != c.runNumber || res.AODNumber != c.aodNumber {
			t.Errorf("%s: got %+v, want {%s %d %d}", c.path, *res, c.period, c.runNumber, c.aodNumber)
		}
	}
}
