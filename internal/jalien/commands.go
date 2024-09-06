package jalien

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func List(path string, longFormat bool) ([]byte, error) {
	var cmd *exec.Cmd
	if longFormat {
		cmd = exec.Command("alien_ls", "-l", path)
	} else {
		cmd = exec.Command("alien_ls", path)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func Find(path string, searchPattern string, longFormat bool) ([]byte, error) {
	var cmd *exec.Cmd
	if longFormat {
		cmd = exec.Command("alien_find", "-w", path, searchPattern)
	} else {
		cmd = exec.Command("alien_find", path, searchPattern)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

type longFormatParsed struct {
	Permissions string
	Owner       string
	Group       string
	Size        uint64
	Month       string
	Day         string
	Time        string
	Name        string
	IsDir       bool
}

func parseLongFormat(line string) (*longFormatParsed, error) {
	parts := strings.Fields(line)
	if len(parts) < 8 {
		return nil, fmt.Errorf("line does not have enough parts")
	}
	isDir := line[0] == 'd'

	size, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, err
	}

	return &longFormatParsed{
		Permissions: parts[0],
		Owner:       parts[1],
		Group:       parts[2],
		Size:        size,
		Month:       parts[4],
		Day:         parts[5],
		Time:        parts[6],
		Name:        strings.Join(parts[7:], " "),
		IsDir:       isDir,
	}, nil
}

func formatSizePretty(bytes uint64) string {
	const (
		_         = iota // ignore first value by assigning to blank identifier
		KB uint64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

type AODFile struct {
	Name       string
	Path       string
	Size       uint64
	PrettySize string
	LHCPeriod  string
	RunNumber  uint64
	AODNumber  uint64
}

var aodFilename = "AO2D.root"

func FindAODFiles(path string) ([]AODFile, error) {
	out, err := Find(path, aodFilename, true)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	matcher := NewAODMatcher()
	aods := make([]AODFile, 0, len(lines))

	for _, line := range lines {
		if line == "" || strings.TrimSpace(line) == "" {
			continue
		}

		lineParsed, err := parseLongFormat(line)
		if err != nil {
			return nil, err
		}

		if lineParsed.IsDir {
			return nil, errors.New("alien_find returned dir, but it shouldn't")
		}

		// find returns full path in parsed Name variable
		aodPath := lineParsed.Name
		pathVariables, err := matcher.MatchAO2DPath(aodPath)
		if err != nil {
			return nil, err
		}

		aods = append(aods, AODFile{
			Name:       aodFilename,
			Path:       aodPath,
			Size:       lineParsed.Size,
			PrettySize: formatSizePretty(lineParsed.Size),
			LHCPeriod:  pathVariables.LHCPeriod,
			RunNumber:  pathVariables.RunNumber,
			AODNumber:  pathVariables.AODNumber,
		})
	}

	return aods, nil
}

type Dir struct {
	Name string
	Path string
}

type File struct {
	Name string
	Path string
	Size uint64
}

type DirectoryContents struct {
	AODFiles   []AODFile
	OtherFiles []File
	Subdirs    []Dir
}

func ListAndParseDirectory(path string) (*DirectoryContents, error) {
	out, err := List(path, true)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	matcher := NewAODMatcher()
	dirContents := &DirectoryContents{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		lineParsed, err := parseLongFormat(line)
		if err != nil {
			return nil, err
		}

		linePath := path + lineParsed.Name

		if lineParsed.IsDir {
			dirContents.Subdirs = append(dirContents.Subdirs, Dir{
				Name: lineParsed.Name,
				Path: linePath,
			})
		} else if lineParsed.Name == aodFilename {
			pathVariables, err := matcher.MatchAO2DPath(linePath)
			if err != nil {
				return nil, err
			}

			dirContents.AODFiles = append(dirContents.AODFiles, AODFile{
				Name:      lineParsed.Name,
				Path:      linePath,
				Size:      lineParsed.Size,
				LHCPeriod: pathVariables.LHCPeriod,
				RunNumber: pathVariables.RunNumber,
				AODNumber: pathVariables.AODNumber,
			})
		} else {
			dirContents.OtherFiles = append(dirContents.OtherFiles, File{
				Name: lineParsed.Name,
				Path: linePath,
				Size: lineParsed.Size,
			})
		}
	}

	return dirContents, nil
}
