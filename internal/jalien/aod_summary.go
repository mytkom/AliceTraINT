package jalien

import "sort"

type AODPeriodSummary struct {
	LHCPeriod string
	FileCount int
	TotalSize uint64
	Runs      []AODRunSummary
}

type AODRunSummary struct {
	LHCPeriod string
	RunNumber uint64
	FileCount int
	TotalSize uint64
}

func SummarizeByPeriod(files []AODFile) []AODPeriodSummary {
	if len(files) == 0 {
		return nil
	}

	periodMap := make(map[string]*AODPeriodSummary)
	for _, file := range files {
		summary, ok := periodMap[file.LHCPeriod]
		if !ok {
			summary = &AODPeriodSummary{LHCPeriod: file.LHCPeriod}
			periodMap[file.LHCPeriod] = summary
		}
		summary.FileCount++
		summary.TotalSize += file.Size
	}

	summaries := make([]AODPeriodSummary, 0, len(periodMap))
	for _, summary := range periodMap {
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].LHCPeriod < summaries[j].LHCPeriod
	})

	return summaries
}

func SummarizeByPeriodAndRun(files []AODFile) []AODPeriodSummary {
	if len(files) == 0 {
		return nil
	}

	type runKey struct {
		period string
		run    uint64
	}

	runMap := make(map[runKey]*AODRunSummary)
	periodTotals := make(map[string]*AODPeriodSummary)

	for _, file := range files {
		key := runKey{period: file.LHCPeriod, run: file.RunNumber}

		runSummary, ok := runMap[key]
		if !ok {
			runSummary = &AODRunSummary{
				LHCPeriod: file.LHCPeriod,
				RunNumber: file.RunNumber,
			}
			runMap[key] = runSummary
		}
		runSummary.FileCount++
		runSummary.TotalSize += file.Size

		periodSummary, ok := periodTotals[file.LHCPeriod]
		if !ok {
			periodSummary = &AODPeriodSummary{LHCPeriod: file.LHCPeriod}
			periodTotals[file.LHCPeriod] = periodSummary
		}
		periodSummary.FileCount++
		periodSummary.TotalSize += file.Size
	}

	periodSummaries := make([]AODPeriodSummary, 0, len(periodTotals))
	for _, periodSummary := range periodTotals {
		runs := make([]AODRunSummary, 0)
		for key, runSummary := range runMap {
			if key.period != periodSummary.LHCPeriod {
				continue
			}
			runs = append(runs, *runSummary)
		}

		sort.Slice(runs, func(i, j int) bool {
			return runs[i].RunNumber < runs[j].RunNumber
		})

		periodSummary.Runs = runs
		periodSummaries = append(periodSummaries, *periodSummary)
	}

	sort.Slice(periodSummaries, func(i, j int) bool {
		return periodSummaries[i].LHCPeriod < periodSummaries[j].LHCPeriod
	})

	return periodSummaries
}
