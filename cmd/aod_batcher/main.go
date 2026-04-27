package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/monalisa"
)

const MONALISA_DEFAULT_BASE_URL = "https://alimonitor.cern.ch"

func getDataPath(dataTag, passName string, run uint64) string {
	yearSuffix := dataTag[3:5]
	return fmt.Sprintf("/alice/data/20%s/%s/%d/%s", yearSuffix, dataTag, run, passName)
}

type config struct {
	path                 string
	runs                 int
	filesPerRun          int
	maxFilesPerBatch     int
	minSizeMB            float64
	outputDir            string
	monalisaBaseUrl      string
	jalienHost           string
	jalienPort           string
	clientCert           string
	clientKey            string
	caCertsDir           string
	jalienTimeoutSeconds uint
}

type metadata struct {
	Path             string  `json:"path"`
	Runs             int     `json:"runs"`
	FilesPerRun      int     `json:"files_per_run"`
	MaxFilesPerBatch int     `json:"max_files_per_batch"`
	MinSizeMB        float64 `json:"min_size_mb"`
	Timestamp        string  `json:"timestamp"`
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	if err := run(cfg); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func parseFlags() (*config, error) {
	cfg := &config{}

	flag.StringVar(&cfg.path, "path", "", "JAliEn path under which to search for AOD files (e.g. /alice/sim/2024/LHC24f3)")
	flag.IntVar(&cfg.runs, "runs", 0, "Total number of runs to select")
	flag.IntVar(&cfg.filesPerRun, "files-per-run", 0, "Number of AOD files to select for each chosen run")
	flag.IntVar(&cfg.maxFilesPerBatch, "max-files-per-batch", -1, "Maximum number of AOD files to include in a single batch")
	flag.Float64Var(&cfg.minSizeMB, "min-size-mb", 0, "Optional minimal AOD file size in megabytes; files smaller than this are excluded")
	flag.StringVar(&cfg.outputDir, "output-dir", "", "Directory where batch .txt files will be written")
	flag.UintVar(&cfg.jalienTimeoutSeconds, "jalien-timeout-seconds", 600, "JAliEn timeout in seconds")
	// JAliEn connectivity; defaults from environment if flags are not provided.
	flag.StringVar(&cfg.monalisaBaseUrl, "monalisa-base-url", MONALISA_DEFAULT_BASE_URL, "Monalisa base url used for querying anchor prod tag")
	flag.StringVar(&cfg.jalienHost, "jalien-host", os.Getenv("JALIEN_HOST"), "JAliEn host (default from $JALIEN_HOST or internal default)")
	flag.StringVar(&cfg.jalienPort, "jalien-port", os.Getenv("JALIEN_PORT"), "JAliEn port (default from $JALIEN_PORT or internal default)")
	flag.StringVar(&cfg.clientCert, "cert", os.Getenv("X509_USER_CERT"), "Path to user X.509 certificate (default from $X509_USER_CERT)")
	flag.StringVar(&cfg.clientKey, "key", os.Getenv("X509_USER_KEY"), "Path to user X.509 key (default from $X509_USER_KEY)")
	flag.StringVar(&cfg.caCertsDir, "cert-dir", os.Getenv("X509_CERT_DIR"), "Directory containing CA certificates (default from $X509_CERT_DIR, falls back to system / grid CAs)")

	flag.Parse()

	if cfg.path == "" {
		return nil, errors.New("flag --path is required")
	}
	if cfg.runs <= 0 {
		return nil, errors.New("flag --runs must be > 0")
	}
	if cfg.filesPerRun <= 0 {
		return nil, errors.New("flag --files-per-run must be > 0")
	}
	if cfg.maxFilesPerBatch <= 0 {
		return nil, errors.New("flag --max-files-per-batch must be > 0")
	}
	if cfg.outputDir == "" {
		return nil, errors.New("flag --output-dir is required")
	}
	if cfg.minSizeMB < 0 {
		return nil, errors.New("flag --min-size-mb must be >= 0")
	}

	// Require certificate and key explicitly. The CA directory is optional
	// because the JAliEn client has its own fallback logic.
	if cfg.clientCert == "" {
		return nil, errors.New("flag --cert (or $X509_USER_CERT) is required")
	}
	if cfg.clientKey == "" {
		return nil, errors.New("flag --key (or $X509_USER_KEY) is required")
	}
	if err := ensureOutputDirs(cfg.outputDir); err != nil {
		return nil, err
	}

	return cfg, nil
}

func run(cfg *config) error {
	client, err := jalien.NewClient(cfg.jalienHost, cfg.jalienPort, cfg.clientCert, cfg.clientKey, cfg.caCertsDir, cfg.jalienTimeoutSeconds)
	if err != nil {
		return fmt.Errorf("cannot create JAliEn client: %w", err)
	}

	allSimAodsByRun, err := getSimAodsByRun(cfg.path, cfg.minSizeMB, client)
	if err != nil {
		return err
	}

	simEligibleRuns, err := filterEligibleRuns(allSimAodsByRun, cfg.runs, cfg.filesPerRun)
	if err != nil {
		return err
	}

	// Uniformly select runs
	selectedRunIdx := uniformIndices(uint64(len(simEligibleRuns)), uint64(cfg.runs))
	selectedRuns := make([]uint64, 0, cfg.runs)
	for _, idx := range selectedRunIdx {
		selectedRuns = append(selectedRuns, simEligibleRuns[idx])
	}

	runToSimAods := make(map[uint64][]jalien.AODFile, len(selectedRuns))
	for _, rn := range selectedRuns {
		// At this point we know len(files) >= filesPerRun by construction.
		idxs := uniformIndices(uint64(len(allSimAodsByRun[rn])), uint64(cfg.filesPerRun))
		selected := make([]jalien.AODFile, 0, cfg.filesPerRun)
		for _, idx := range idxs {
			selected = append(selected, allSimAodsByRun[rn][idx])
		}
		runToSimAods[rn] = selected
	}

	// Get anchored data production (real data tag)
	monc := monalisa.NewMonalisaClient(cfg.clientKey, cfg.clientCert, cfg.monalisaBaseUrl)
	aodRow, err := monc.GetMCRow(allSimAodsByRun[selectedRuns[0]][0].LHCPeriod)
	if err != nil {
		return fmt.Errorf("error while getting anchor production tag: %s", err.Error())
	}

	runList, err := monc.GetRunList()
	if err != nil {
		return err
	}

	runToDataAods := make(map[uint64][]jalien.AODFile, len(selectedRuns))
	for _, sr := range selectedRuns {
		dataTag := ""
		for _, t := range runList.RunToTags[sr] {
			if !runList.TagToIsMC[t] {
				dataTag = t
				break
			}
		}
		if dataTag == "" {
			return fmt.Errorf("Cannot find a anchored production period for run %d", sr)
		}

		dataPath := getDataPath(dataTag, aodRow.PassName, sr)
		aods, err := getAods(dataPath, cfg.minSizeMB/2, client)
		if err != nil {
			return err
		}

		fCount := min(cfg.filesPerRun, len(aods))
		idxs := uniformIndices(uint64(len(aods)), uint64(fCount))
		selected := make([]jalien.AODFile, 0, fCount)
		for _, idx := range idxs {
			selected = append(selected, aods[idx])
		}

		runToDataAods[sr] = selected
	}

	if err := writeFilesInBatches(filepath.Join(cfg.outputDir, "sim"), flattenRunFiles(runToSimAods, selectedRuns), cfg.maxFilesPerBatch); err != nil {
		return err
	}
	if err := writeFilesInBatches(filepath.Join(cfg.outputDir, "data"), flattenRunFiles(runToDataAods, selectedRuns), cfg.maxFilesPerBatch); err != nil {
		return err
	}

	// Persist non-sensitive configuration metadata alongside the batch files.
	if err := writeMetadataFile(cfg); err != nil {
		return err
	}

	return nil
}

func getAods(path string, minSizeMB float64, client *jalien.Client) ([]jalien.AODFile, error) {
	aods, err := client.FindAODFiles(path)
	if err != nil {
		return nil, fmt.Errorf("failed to discover AOD files under %q: %w", path, err)
	}
	if len(aods) == 0 {
		return nil, fmt.Errorf("no AOD files found under path %q", path)
	}

	// Apply optional minimum size filter (in MB) before any selection logic.
	if minSizeMB > 0 {
		minBytes := uint64(minSizeMB * 1024 * 1024)
		filtered := make([]jalien.AODFile, 0, len(aods))
		for _, f := range aods {
			if f.Size >= minBytes {
				filtered = append(filtered, f)
			}
		}
		aods = filtered

		// sort for determinism
		sort.Slice(aods, func(i, j int) bool {
			if aods[i].AODNumber == aods[j].AODNumber {
				return aods[i].Path < aods[j].Path
			}

			return aods[i].AODNumber < aods[j].AODNumber
		})

		if len(aods) == 0 {
			return nil, fmt.Errorf("no AOD files >= %.2f MB found under path %q", minSizeMB, path)
		}
	}

	return aods, nil
}

func getSimAodsByRun(path string, minSizeMB float64, client *jalien.Client) (map[uint64][]jalien.AODFile, error) {
	aods, err := getAods(path, minSizeMB, client)
	if err != nil {
		return nil, err
	}

	// Group by run and sort.
	runsMap := make(map[uint64][]jalien.AODFile)
	for _, f := range aods {
		runsMap[f.RunNumber] = append(runsMap[f.RunNumber], f)
	}

	return runsMap, nil
}

func filterEligibleRuns(runsMap map[uint64][]jalien.AODFile, reqRunCount int, minFilesPerRun int) ([]uint64, error) {
	runNumbers := make([]uint64, 0, len(runsMap))
	for rn := range runsMap {
		runNumbers = append(runNumbers, rn)
	}
	sort.Slice(runNumbers, func(i, j int) bool { return runNumbers[i] < runNumbers[j] })

	// Filter to runs that have at least the requested number of AODs so we do
	// not fail later when selecting files-per-run.
	eligibleRuns := make([]uint64, 0, len(runNumbers))
	for _, rn := range runNumbers {
		if len(runsMap[rn]) >= minFilesPerRun {
			eligibleRuns = append(eligibleRuns, rn)
		}
	}

	if len(eligibleRuns) < reqRunCount {
		return nil, fmt.Errorf(
			"requested %d runs with at least %d AOD files each, but only %d runs satisfy this condition",
			reqRunCount, minFilesPerRun, len(eligibleRuns))
	}

	return eligibleRuns, nil
}

// uniformIndices returns k indices in [0, n) that are as uniformly spaced as
// possible over the interval when n >= k > 0. The result is strictly
// increasing and deterministic.
func uniformIndices(n, k uint64) []uint64 {
	if k <= 0 || n <= 0 {
		return []uint64{}
	}
	if k == 1 {
		return []uint64{n / 2}
	}

	step := float64(n) / float64(k)
	offset := step / 2.0

	indices := make([]uint64, 0, k)
	prev := n + 1
	for i := uint64(0); i < k; i++ {
		x := uint64(offset + float64(i)*step)
		if x >= n {
			x = n - 1
		}
		// Ensure strict monotonicity in the unlikely case rounding produces duplicates.
		if prev < n && x <= prev {
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

func flattenRunFiles(runToAods map[uint64][]jalien.AODFile, runOrder []uint64) []jalien.AODFile {
	flat := make([]jalien.AODFile, 0, len(runOrder))
	for _, rn := range runOrder {
		flat = append(flat, runToAods[rn]...)
	}

	return flat
}

func writeFilesInBatches(outputDir string, files []jalien.AODFile, maxFilesPerBatch int) error {
	batchCount := (len(files) + maxFilesPerBatch - 1) / maxFilesPerBatch
	padding := len(fmt.Sprintf("%d", batchCount))

	for start, batchIdx := 0, 1; start < len(files); start, batchIdx = start+maxFilesPerBatch, batchIdx+1 {
		end := min(start+maxFilesPerBatch, len(files))
		filename := fmt.Sprintf("batch_%0*d.txt", padding, batchIdx)
		if err := writeBatchFile(outputDir, filename, files[start:end]); err != nil {
			return err
		}
	}

	return nil
}

func writeBatchFile(output_dir, filename string, batchFiles []jalien.AODFile) error {
	path := filepath.Join(output_dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create batch file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	for _, a := range batchFiles {
		if _, err := fmt.Fprintln(f, a.Path); err != nil {
			return fmt.Errorf("failed to write to batch file %q: %w", path, err)
		}
	}

	return nil
}

func writeMetadataFile(cfg *config) error {
	meta := metadata{
		Path:             cfg.path,
		Runs:             cfg.runs,
		FilesPerRun:      cfg.filesPerRun,
		MaxFilesPerBatch: cfg.maxFilesPerBatch,
		MinSizeMB:        cfg.minSizeMB,
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	path := filepath.Join(cfg.outputDir, "metadata.json")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create metadata file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(meta); err != nil {
		return fmt.Errorf("cannot write metadata file %q: %w", path, err)
	}

	return nil
}

func ensureOutputDirs(outputDir string) error {
	// 0o750: owner rwx, group rx, others no access. Keeps output private while
	// still allowing read/execute traversal for shared project group.
	for _, dir := range []string{
		outputDir,
		filepath.Join(outputDir, "sim"),
		filepath.Join(outputDir, "data"),
	} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("cannot prepare output directory %q: %w", dir, err)
		}
	}

	metadataPath := filepath.Join(outputDir, "metadata.json")
	testFile, err := os.CreateTemp(outputDir, ".write-check-*")
	if err != nil {
		return fmt.Errorf("output directory %q not writable: %w", outputDir, err)
	}
	testPath := testFile.Name()
	_ = testFile.Close()
	_ = os.Remove(testPath)
	// Ensure we can replace/create metadata file later.
	if _, err := os.Stat(metadataPath); err == nil {
		f, err := os.OpenFile(metadataPath, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("metadata file %q not writable: %w", metadataPath, err)
		}
		_ = f.Close()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot stat metadata file %q: %w", metadataPath, err)
	}

	return nil
}
