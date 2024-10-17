package jalien

import (
	"regexp"
	"strconv"
)

type filepathAODMatcher struct {
	regexpCompiled *regexp.Regexp
}

type filepathAODMatcherResult struct {
	LHCPeriod string
	RunNumber uint64
	AODNumber uint64
}

func newAODMatcher() *filepathAODMatcher {
	regexpPattern := `/(?P<LHCPeriod>LHC[a-zA-Z0-9]+)/\d+/(?P<RunNumber>\d+)/AOD/(?P<AODNumber>\d+)/AO2D\.root`

	return &filepathAODMatcher{
		regexpCompiled: regexp.MustCompile(regexpPattern),
	}
}

func (m *filepathAODMatcher) MatchAO2DPath(path string) (*filepathAODMatcherResult, error) {
	match := m.regexpCompiled.FindStringSubmatch(path)

	runNumber, err := strconv.ParseUint(match[2], 10, 64)
	if err != nil {
		return nil, err
	}

	aodNumber, err := strconv.ParseUint(match[3], 10, 64)
	if err != nil {
		return nil, err
	}

	return &filepathAODMatcherResult{
		LHCPeriod: match[1],
		RunNumber: runNumber,
		AODNumber: aodNumber,
	}, nil
}
