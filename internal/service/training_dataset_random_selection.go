package service

import (
	"fmt"
	"sort"

	"github.com/mytkom/AliceTraINT/internal/jalien"
)

func (s *TrainingDatasetService) SelectRandomAODSubset(path string, runCount, filesPerRun int, minSizeBytes uint64) ([]jalien.AODFile, error) {
	if path == "" {
		return nil, &ErrHandlerValidation{
			Field: "path",
			Msg:   "is required",
		}
	}
	if runCount <= 0 {
		return nil, &ErrHandlerValidation{
			Field: "runCount",
			Msg:   "must be > 0",
		}
	}
	if filesPerRun <= 0 {
		return nil, &ErrHandlerValidation{
			Field: "filesPerRun",
			Msg:   "must be > 0",
		}
	}

	aods, err := s.JAliEn.FindAODFiles(path)
	if err != nil {
		return nil, handleJAlienError(err)
	}
	if len(aods) == 0 {
		return nil, &ErrHandlerValidation{
			Field: "path",
			Msg:   fmt.Sprintf("no AOD files found under %q", path),
		}
	}

	if minSizeBytes > 0 {
		filtered := make([]jalien.AODFile, 0, len(aods))
		for _, f := range aods {
			if f.Size >= minSizeBytes {
				filtered = append(filtered, f)
			}
		}
		aods = filtered
		if len(aods) == 0 {
			return nil, &ErrHandlerValidation{
				Field: "minFileSizeMB",
				Msg:   fmt.Sprintf("no AOD files >= %.0f MB found under path %q", float64(minSizeBytes)/(1024*1024), path),
			}
		}
	}

	// Group by run and sort.
	runsMap := make(map[uint64][]jalien.AODFile)
	for _, f := range aods {
		runsMap[f.RunNumber] = append(runsMap[f.RunNumber], f)
	}

	runNumbers := make([]uint64, 0, len(runsMap))
	for rn := range runsMap {
		runNumbers = append(runNumbers, rn)
	}
	sort.Slice(runNumbers, func(i, j int) bool { return runNumbers[i] < runNumbers[j] })

	// Filter to runs that have at least the requested number of AODs so we do
	// not fail later when selecting files-per-run.
	eligibleRuns := make([]uint64, 0, len(runNumbers))
	for _, rn := range runNumbers {
		if len(runsMap[rn]) >= filesPerRun {
			eligibleRuns = append(eligibleRuns, rn)
		}
	}

	if len(eligibleRuns) < runCount {
		return nil, &ErrHandlerValidation{
			Field: "runCount",
			Msg: fmt.Sprintf(
				"requested %d runs with at least %d AOD files each, but only %d runs satisfy this under path %q",
				runCount, filesPerRun, len(eligibleRuns), path,
			),
		}
	}

	selectedRunIdx := uniformIndices(len(eligibleRuns), runCount)
	selectedRuns := make([]uint64, 0, runCount)
	for _, idx := range selectedRunIdx {
		selectedRuns = append(selectedRuns, eligibleRuns[idx])
	}

	// For determinism, within each selected run sort AODs by AODNumber then by path.
	selected := make([]jalien.AODFile, 0, runCount*filesPerRun)
	for _, rn := range selectedRuns {
		files := append([]jalien.AODFile(nil), runsMap[rn]...)
		sort.Slice(files, func(i, j int) bool {
			if files[i].AODNumber == files[j].AODNumber {
				return files[i].Path < files[j].Path
			}
			return files[i].AODNumber < files[j].AODNumber
		})

		idxs := uniformIndices(len(files), filesPerRun)
		for _, idx := range idxs {
			selected = append(selected, files[idx])
		}
	}

	return selected, nil
}

// uniformIndices returns k indices in [0, n) that are as uniformly spaced as
// possible over the interval when n >= k > 0. The result is strictly
// increasing and deterministic.
func uniformIndices(n, k int) []int {
	if k <= 0 || n <= 0 {
		return []int{}
	}
	if k == 1 {
		return []int{n / 2}
	}

	step := float64(n) / float64(k)
	offset := step / 2.0

	indices := make([]int, 0, k)
	prev := -1
	for i := 0; i < k; i++ {
		x := int(offset + float64(i)*step)
		if x >= n {
			x = n - 1
		}
		// Ensure strict monotonicity in the unlikely case rounding produces duplicates.
		if x <= prev {
			x = prev + 1
			if x >= n {
				x = n - 1
			}
		}
		indices = append(indices, x)
		prev = x
	}
	return indices
}
