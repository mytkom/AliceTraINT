package service

import (
	"errors"
	"time"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/jalien"
)

func resolveDatasetTimestampRange(aodFiles []jalien.AODFile, monalisaService IMonalisaService, ccdbService ICCDBService) (time.Time, time.Time, error) {
	if len(aodFiles) == 0 {
		return time.Time{}, time.Time{}, errors.New("unexpected behaviour: empty training dataset")
	}

	distinctPeriods := make(map[string]struct{}, len(aodFiles))
	for _, aod := range aodFiles {
		if aod.LHCPeriod != "" {
			distinctPeriods[aod.LHCPeriod] = struct{}{}
		}
	}

	if len(distinctPeriods) == 0 {
		return time.Time{}, time.Time{}, errors.New("unexpected behaviour: empty training dataset periods")
	}

	runList, err := monalisaService.GetRunList()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	globalMinRun, globalMaxRun, err := findGlobalRunRange(runList.TagToRuns, distinctPeriods)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	firstRunInfo, lastRunInfo, err := getRunInfoRangeFromCCDB(ccdbService, globalMinRun, globalMaxRun)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return time.UnixMilli(int64(firstRunInfo.SOR)).UTC(), time.UnixMilli(int64(lastRunInfo.EOR)).UTC(), nil
}

func findGlobalRunRange(runMap map[string][]uint64, periods map[string]struct{}) (uint64, uint64, error) {
	var globalMinRun uint64
	var globalMaxRun uint64
	initialized := false

	for period := range periods {
		runs, ok := runMap[period]
		if !ok || len(runs) == 0 {
			continue
		}

		for _, run := range runs {
			if !initialized || run < globalMinRun {
				globalMinRun = run
			}
			if !initialized || run > globalMaxRun {
				globalMaxRun = run
			}
			initialized = true
		}
	}

	if !initialized {
		return 0, 0, errors.New("no runs found for dataset periods in Monalisa run list")
	}

	return globalMinRun, globalMaxRun, nil
}

func getRunInfoRangeFromCCDB(ccdbService ICCDBService, smallestRun, greatestRun uint64) (*ccdb.RunInformation, *ccdb.RunInformation, error) {
	firstRunInfo, err := ccdbService.GetRunInformation(smallestRun)
	if err != nil {
		return nil, nil, handleCCDBError(err)
	}

	lastRunInfo, err := ccdbService.GetRunInformation(greatestRun)
	if err != nil {
		return nil, nil, handleCCDBError(err)
	}

	return firstRunInfo, lastRunInfo, nil
}

func getDatasetAODRunRange(aodFiles []jalien.AODFile) (uint64, uint64, error) {
	var minRun, maxRun uint64
	initialized := false

	for _, aod := range aodFiles {
		if !initialized || aod.RunNumber < minRun {
			minRun = aod.RunNumber
		}
		if !initialized || aod.RunNumber > maxRun {
			maxRun = aod.RunNumber
		}
		initialized = true
	}

	if !initialized {
		return 0, 0, errors.New("unexpected behaviour: empty training dataset")
	}

	return minRun, maxRun, nil
}
