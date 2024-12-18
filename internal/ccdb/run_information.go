package ccdb

import (
	"fmt"
	"strconv"
)

type CCDBApi struct {
	baseURL string
}

func NewCCDBApi(baseUrl string) *CCDBApi {
	return &CCDBApi{
		baseURL: baseUrl,
	}
}

type RunInformation struct {
	RunNumber uint64
	SOR       uint64
	EOR       uint64
}

const (
	RCT_ENDPOINT string = "RCT/Info/RunInformation"
	AGENT        string = "AliceTraINT_Agent/1.0"
)

func (c *CCDBApi) GetRunInformation(runNumber uint64) (*RunInformation, error) {
	url := fmt.Sprintf("%s/%s/%d", c.baseURL, RCT_ENDPOINT, runNumber)

	headers, err := doRemoteHeaderCall(url, AGENT, -1)
	if err != nil {
		return nil, err
	}

	sorStr, sorOk := headers["Sor"]
	if !sorOk {
		return nil, fmt.Errorf("CCDB: SOR not present for run")
	}

	eorStr, eorOk := headers["Eor"]
	if !eorOk {
		return nil, fmt.Errorf("CCDB: EOR not present for run")
	}

	sor, err := parseUint64(sorStr)
	if err != nil {
		return nil, err
	}

	eor, err := parseUint64(eorStr)
	if err != nil {
		return nil, err
	}

	return &RunInformation{
		RunNumber: runNumber,
		SOR:       sor,
		EOR:       eor,
	}, nil
}

func parseUint64(val string) (uint64, error) {
	valUint, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return valUint, nil
}
