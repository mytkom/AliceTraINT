package jalien

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AODFile struct {
	Name      string
	Path      string
	Size      uint64
	LHCPeriod string
	RunNumber uint64
	AODNumber uint64
}

var aodFilename = "AO2D.root"

// getStringField retrieves a string field from a JAliEn result map.
func getStringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case fmt.Stringer:
			return val.String()
		case float64:
			// JSON numbers are float64; but for IDs or sizes we should not call this helper.
			return strconv.FormatFloat(val, 'f', -1, 64)
		default:
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// getUint64Field attempts to extract an uint64 from a JAliEn result map.
func getUint64Field(m map[string]any, key string) (uint64, error) {
	if m == nil {
		return 0, fmt.Errorf("missing field %q", key)
	}
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}

	switch val := v.(type) {
	case float64:
		if val < 0 {
			return 0, fmt.Errorf("negative value for %q", key)
		}
		return uint64(val), nil
	case string:
		if val == "" {
			return 0, fmt.Errorf("empty string for %q", key)
		}
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid uint value %q for %s: %w", val, key, err)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unsupported type %T for %q", v, key)
	}
}

// listDirectory returns the raw JAliEn ls results for a directory.
func (c *Client) listDirectory(ctx context.Context, path string) ([]map[string]any, error) {
	if path == "" {
		path = "/"
	}

	// Use -nomsg to avoid server-side formatted messages; we only care about JSON.
	opts := []string{"-nomsg", "-a", "-F", "-l", path}

	resp, err := c.send(ctx, "ls", opts)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// findFiles returns raw JAliEn find results for a directory and pattern.
func (c *Client) findFiles(ctx context.Context, dir, pattern string) ([]map[string]any, error) {
	if dir == "" {
		dir = "/"
	}
	if pattern == "" {
		return nil, errors.New("jalien: empty search pattern")
	}

	opts := []string{"-nomsg", "-f", "-a", "-s", dir, pattern}

	resp, err := c.send(ctx, "find", opts)
	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// FindAODFiles finds AO2D.root files under the specified path using the
// provided JAliEn client.
func (client *Client) FindAODFiles(path string) ([]AODFile, error) {
	minimal_path_reg := regexp.MustCompile(`^/alice/sim/.+/.+|/alice/data/.+/.+`)
	if !minimal_path_reg.MatchString(path) {
		return nil, errors.New("path must start with `/alice/sim/` or `/alice/data/` and has at least one more level")
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()

	rawResults, err := client.findFiles(ctx, path, aodFilename)
	if err != nil {
		return nil, err
	}

  aods := make([]AODFile, 0, len(rawResults))

  if strings.Contains(path, "/alice/data") {
    // process data
    matcher := newDataAODMatcher()

    for i, r := range rawResults {
    	// Skip directories if type information is present.
    	if t := getStringField(r, "type"); t != "" {
    		if strings.ToLower(t) != "f" {
    			continue
    		}
    	}

    	aodPath := getStringField(r, "lfn")

    	pathVariables, err := matcher.MatchAO2DPath(aodPath)
    	if err != nil {
    		print("Skipping file: %s, error: %s", aodPath, err.Error())
    		continue
    	}
    	if pathVariables == nil {
    		continue
    	}

    	size, err := getUint64Field(r, "size")
    	if err != nil {
    		return nil, err
    	}

    	aods = append(aods, AODFile{
    		Name:      aodFilename,
    		Path:      aodPath,
    		Size:      size,
    		LHCPeriod: pathVariables.LHCPeriod,
    		RunNumber: pathVariables.RunNumber,
    		AODNumber: uint64(i),
    	})
    }
  } else {
    // process MC
    matcher := newAODMatcher()

    for _, r := range rawResults {
    	// Skip directories if type information is present.
    	if t := getStringField(r, "type"); t != "" {
    		if strings.ToLower(t) != "f" {
    			continue
    		}
    	}

    	aodPath := getStringField(r, "lfn")

    	pathVariables, err := matcher.MatchAO2DPath(aodPath)
    	if err != nil {
    		print("Skipping file: %s, error: %s", aodPath, err.Error())
    		continue
    	}
    	if pathVariables == nil {
    		continue
    	}

    	size, err := getUint64Field(r, "size")
    	if err != nil {
    		return nil, err
    	}

    	aods = append(aods, AODFile{
    		Name:      aodFilename,
    		Path:      aodPath,
    		Size:      size,
    		LHCPeriod: pathVariables.LHCPeriod,
    		RunNumber: pathVariables.RunNumber,
    		AODNumber: pathVariables.AODNumber,
    	})
    }
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

// ListAndParseDirectory lists the contents of a directory using the provided
// JAliEn client and parses them into DirectoryContents.
func (client *Client) ListAndParseDirectory(path string) (*DirectoryContents, error) {
	// Ensure path has a trailing slash for consistent path handling below.
	if path == "" {
		path = "/"
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()

	entries, err := client.listDirectory(ctx, path)
	if err != nil {
		return nil, err
	}

	matcher := newAODMatcher()
	dirContents := &DirectoryContents{}

	for _, e := range entries {
		name := getStringField(e, "name")
		entryPath := getStringField(e, "path")
		entryPath = strings.TrimSuffix(entryPath, "/")

		// Normalize to always use trailing slash for directories, as before.
		isDir := false
		if t := getStringField(e, "permissions"); t != "" {
			if strings.HasPrefix(strings.ToLower(t), "d") || strings.Contains(strings.ToLower(t), "dir") {
				isDir = true
			}
		}

		size, err := getUint64Field(e, "size")
		if err != nil && !isDir {
			return nil, err
		}

		if isDir {
			dirContents.Subdirs = append(dirContents.Subdirs, Dir{
				Name: strings.TrimSuffix(name, "/"),
				Path: entryPath,
			})
			continue
		}

		if strings.TrimSuffix(name, "/") == aodFilename {
			pathVariables, err := matcher.MatchAO2DPath(entryPath)
			if err != nil {
				print("Skipping file: %s, error: %s", entryPath, err.Error())
				continue
			}
			if pathVariables == nil {
				// Not in the expected AO2D layout, treat as regular file.
				dirContents.OtherFiles = append(dirContents.OtherFiles, File{
					Name: strings.TrimSuffix(name, "/"),
					Path: entryPath,
					Size: size,
				})
				continue
			}

			dirContents.AODFiles = append(dirContents.AODFiles, AODFile{
				Name:      strings.TrimSuffix(name, "/"),
				Path:      entryPath,
				Size:      size,
				LHCPeriod: pathVariables.LHCPeriod,
				RunNumber: pathVariables.RunNumber,
				AODNumber: pathVariables.AODNumber,
			})
		} else {
			dirContents.OtherFiles = append(dirContents.OtherFiles, File{
				Name: strings.TrimSuffix(name, "/"),
				Path: entryPath,
				Size: size,
			})
		}
	}

	return dirContents, nil
}
