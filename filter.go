package rman

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Filter defines criteria for filtering manifest files.
// Multiple criteria are ANDed together — a file must match all specified criteria.
type Filter struct {
	// Glob pattern to match against file paths (e.g., "*.dll", "DATA/**")
	Glob string
	// Regex pattern to match against file paths
	Regex string
	// FlagNames filters files that have ALL of the specified flags
	FlagNames []string
	// MinSize filters files with size >= MinSize bytes (0 means no minimum)
	MinSize uint32
	// MaxSize filters files with size <= MaxSize bytes (0 means no maximum)
	MaxSize uint32
}

// FilterFiles returns the subset of files that match all filter criteria.
func FilterFiles(files []FileEntry, f Filter) ([]FileEntry, error) {
	var re *regexp.Regexp
	if f.Regex != "" {
		var err error
		re, err = regexp.Compile(f.Regex)
		if err != nil {
			return nil, err
		}
	}

	result := make([]FileEntry, 0)
	for _, file := range files {
		if !matchesFilter(file, f, re) {
			continue
		}
		result = append(result, file)
	}
	return result, nil
}

func matchesFilter(file FileEntry, f Filter, re *regexp.Regexp) bool {
	if f.Glob != "" {
		matched, err := filepath.Match(f.Glob, filepath.Base(file.Name))
		if err != nil || !matched {
			// Also try matching against full path
			matched2, _ := filepath.Match(f.Glob, file.Name)
			if !matched2 {
				return false
			}
		}
	}

	if re != nil && !re.MatchString(file.Name) {
		return false
	}

	if len(f.FlagNames) > 0 {
		fileFlags := make(map[string]bool, len(file.Flags))
		for _, fl := range file.Flags {
			fileFlags[strings.ToLower(fl.Name)] = true
		}
		for _, name := range f.FlagNames {
			if !fileFlags[strings.ToLower(name)] {
				return false
			}
		}
	}

	if f.MinSize > 0 && file.FileSize < f.MinSize {
		return false
	}

	if f.MaxSize > 0 && file.FileSize > f.MaxSize {
		return false
	}

	return true
}
