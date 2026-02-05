package rman

// DiffResult contains the differences between two manifests.
type DiffResult struct {
	Added    []FileEntry
	Removed  []FileEntry
	Modified []FileEntry
}

// Diff compares two manifests and returns the differences.
// Files are matched by path. A file is considered modified if its
// size or chunk composition has changed.
func Diff(old, new *Manifest) DiffResult {
	oldFiles := make(map[string]FileEntry, len(old.Files))
	for _, f := range old.Files {
		oldFiles[f.Name] = f
	}

	newFiles := make(map[string]FileEntry, len(new.Files))
	for _, f := range new.Files {
		newFiles[f.Name] = f
	}

	var result DiffResult

	for _, nf := range new.Files {
		of, exists := oldFiles[nf.Name]
		if !exists {
			result.Added = append(result.Added, nf)
			continue
		}
		if filesModified(of, nf) {
			result.Modified = append(result.Modified, nf)
		}
	}

	for _, of := range old.Files {
		if _, exists := newFiles[of.Name]; !exists {
			result.Removed = append(result.Removed, of)
		}
	}

	return result
}

func filesModified(old, new FileEntry) bool {
	if old.FileSize != new.FileSize {
		return true
	}
	if len(old.Chunks) != len(new.Chunks) {
		return true
	}
	for i := range old.Chunks {
		if old.Chunks[i].ChunkID != new.Chunks[i].ChunkID {
			return true
		}
	}
	return false
}
