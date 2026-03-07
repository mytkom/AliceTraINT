package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/mytkom/AliceTraINT/internal/jalien"
)

type config struct {
	path            string
	runs            int
	filesPerRun     int
	maxRunsPerBatch int
	minSizeMB       float64
	outputDir       string

	jalienHost           string
	jalienPort           string
	clientCert           string
	clientKey            string
	caCertsDir           string
	jalienTimeoutSeconds uint
}

type metadata struct {
	Path            string  `json:"path"`
	Runs            int     `json:"runs"`
	FilesPerRun     int     `json:"files_per_run"`
	MaxRunsPerBatch int     `json:"max_runs_per_batch"`
	MinSizeMB       float64 `json:"min_size_mb"`
	Timestamp       string  `json:"timestamp"`
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
	flag.IntVar(&cfg.maxRunsPerBatch, "max-runs-per-batch", 0, "Maximum number of runs to include in a single batch")
	flag.Float64Var(&cfg.minSizeMB, "min-size-mb", 0, "Optional minimal AOD file size in megabytes; files smaller than this are excluded")
	flag.StringVar(&cfg.outputDir, "output-dir", "", "Directory where batch .txt files will be written")
	flag.UintVar(&cfg.jalienTimeoutSeconds, "jalien-timeout-seconds", 600, "JAliEn timeout in seconds")
	// JAliEn connectivity; defaults from environment if flags are not provided.
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
	if cfg.maxRunsPerBatch <= 0 {
		return nil, errors.New("flag --max-runs-per-batch must be > 0")
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

	return cfg, nil
}

func run(cfg *config) error {
	// Ensure output directory exists and is writable.
	if err := os.MkdirAll(cfg.outputDir, 0o755); err != nil {
		return fmt.Errorf("cannot create output directory %q: %w", cfg.outputDir, err)
	}

	client, err := jalien.NewClient(cfg.jalienHost, cfg.jalienPort, cfg.clientCert, cfg.clientKey, cfg.caCertsDir, cfg.jalienTimeoutSeconds)
	if err != nil {
		return fmt.Errorf("cannot create JAliEn client: %w", err)
	}

	aods, err := client.FindAODFiles(cfg.path)
	if err != nil {
		return fmt.Errorf("failed to discover AOD files under %q: %w", cfg.path, err)
	}
	if len(aods) == 0 {
		return fmt.Errorf("no AOD files found under path %q", cfg.path)
	}

	// Apply optional minimum size filter (in MB) before any selection logic.
	if cfg.minSizeMB > 0 {
		minBytes := uint64(cfg.minSizeMB * 1024 * 1024)
		filtered := make([]jalien.AODFile, 0, len(aods))
		for _, f := range aods {
			if f.Size >= minBytes {
				filtered = append(filtered, f)
			}
		}
		aods = filtered
		if len(aods) == 0 {
			return fmt.Errorf("no AOD files >= %.2f MB found under path %q", cfg.minSizeMB, cfg.path)
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
		if len(runsMap[rn]) >= cfg.filesPerRun {
			eligibleRuns = append(eligibleRuns, rn)
		}
	}

	if len(eligibleRuns) < cfg.runs {
		return fmt.Errorf(
			"requested %d runs with at least %d AOD files each, but only %d runs satisfy this under %q",
			cfg.runs, cfg.filesPerRun, len(eligibleRuns), cfg.path,
		)
	}

	selectedRunIdx := uniformIndices(len(eligibleRuns), cfg.runs)
	selectedRuns := make([]uint64, 0, cfg.runs)
	for _, idx := range selectedRunIdx {
		selectedRuns = append(selectedRuns, eligibleRuns[idx])
	}

	// For determinism, within each selected run sort AODs by AODNumber then by path.
	selectedFilesByRun := make(map[uint64][]jalien.AODFile, len(selectedRuns))
	for _, rn := range selectedRuns {
		files := append([]jalien.AODFile(nil), runsMap[rn]...)
		sort.Slice(files, func(i, j int) bool {
			if files[i].AODNumber == files[j].AODNumber {
				return files[i].Path < files[j].Path
			}
			return files[i].AODNumber < files[j].AODNumber
		})

		// At this point we know len(files) >= filesPerRun by construction.
		idxs := uniformIndices(len(files), cfg.filesPerRun)
		selected := make([]jalien.AODFile, 0, cfg.filesPerRun)
		for _, idx := range idxs {
			selected = append(selected, files[idx])
		}
		selectedFilesByRun[rn] = selected
	}

	// Build batches of runs.
	if cfg.maxRunsPerBatch <= 0 {
		return errors.New("--max-runs-per-batch must be greater than 0")
	}

	batchCount := int(math.Ceil(float64(len(selectedRuns)) / float64(cfg.maxRunsPerBatch)))
	padding := len(fmt.Sprintf("%d", batchCount)) // number of digits for zero-padding

	batchIdx := 0
	for start := 0; start < len(selectedRuns); start += cfg.maxRunsPerBatch {
		end := start + cfg.maxRunsPerBatch
		if end > len(selectedRuns) {
			end = len(selectedRuns)
		}
		batchRuns := selectedRuns[start:end]

		batchIdx++
		filename := fmt.Sprintf("batch_%0*d.txt", padding, batchIdx)
		fullPath := filepath.Join(cfg.outputDir, filename)

		if err := writeBatchFile(fullPath, batchRuns, selectedFilesByRun); err != nil {
			return err
		}
	}

	// Persist non-sensitive configuration metadata alongside the batch files.
	if err := writeMetadataFile(cfg); err != nil {
		return err
	}

	return nil
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

func writeBatchFile(path string, runs []uint64, filesByRun map[uint64][]jalien.AODFile) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create batch file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	for _, rn := range runs {
		files := filesByRun[rn]
		for _, a := range files {
			if _, err := fmt.Fprintln(f, a.Path); err != nil {
				return fmt.Errorf("failed to write to batch file %q: %w", path, err)
			}
		}
	}

	return nil
}

func writeMetadataFile(cfg *config) error {
	meta := metadata{
		Path:            cfg.path,
		Runs:            cfg.runs,
		FilesPerRun:     cfg.filesPerRun,
		MaxRunsPerBatch: cfg.maxRunsPerBatch,
		MinSizeMB:       cfg.minSizeMB,
		Timestamp:       time.Now().Format(time.RFC3339),
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
