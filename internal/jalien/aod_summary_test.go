package jalien_test

import (
	"testing"

	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
)

func TestSummarizeByPeriod(t *testing.T) {
	files := []jalien.AODFile{
		{LHCPeriod: "LHC24f3", RunNumber: 1, Size: 100},
		{LHCPeriod: "LHC24f3", RunNumber: 2, Size: 200},
		{LHCPeriod: "LHC24b1b", RunNumber: 3, Size: 300},
	}

	summaries := jalien.SummarizeByPeriod(files)

	assert.Len(t, summaries, 2)
	assert.Equal(t, "LHC24b1b", summaries[0].LHCPeriod)
	assert.Equal(t, 1, summaries[0].FileCount)
	assert.Equal(t, uint64(300), summaries[0].TotalSize)
	assert.Equal(t, "LHC24f3", summaries[1].LHCPeriod)
	assert.Equal(t, 2, summaries[1].FileCount)
	assert.Equal(t, uint64(300), summaries[1].TotalSize)
}

func TestSummarizeByPeriodAndRun(t *testing.T) {
	files := []jalien.AODFile{
		{LHCPeriod: "LHC24f3", RunNumber: 2, Size: 100},
		{LHCPeriod: "LHC24f3", RunNumber: 1, Size: 200},
		{LHCPeriod: "LHC24f3", RunNumber: 1, Size: 50},
		{LHCPeriod: "LHC24b1b", RunNumber: 9, Size: 300},
	}

	summaries := jalien.SummarizeByPeriodAndRun(files)

	assert.Len(t, summaries, 2)
	assert.Equal(t, "LHC24b1b", summaries[0].LHCPeriod)
	assert.Equal(t, 1, summaries[0].FileCount)
	assert.Len(t, summaries[0].Runs, 1)
	assert.Equal(t, uint64(9), summaries[0].Runs[0].RunNumber)

	assert.Equal(t, "LHC24f3", summaries[1].LHCPeriod)
	assert.Equal(t, 3, summaries[1].FileCount)
	assert.Equal(t, uint64(350), summaries[1].TotalSize)
	assert.Len(t, summaries[1].Runs, 2)
	assert.Equal(t, uint64(1), summaries[1].Runs[0].RunNumber)
	assert.Equal(t, 2, summaries[1].Runs[0].FileCount)
	assert.Equal(t, uint64(250), summaries[1].Runs[0].TotalSize)
	assert.Equal(t, uint64(2), summaries[1].Runs[1].RunNumber)
}
