package configfiles

import (
	"os"
	"path/filepath"
)

// EnsureInDir agents.json, config.json, operaagent.json dosyalarını
// hedef klasörde bulundurur. Yoksa .example'dan veya üst klasörlerden kopyalar.
func EnsureInDir(targetDir string) {
	type filePair struct{ name, example string }
	files := []filePair{
		{"config.json", "config.example.json"},
		{"agents.json", "agents.json.example"},
		{"operaagent.json", "operaagent.json.example"},
	}
	sourceDirs := []string{targetDir}
	sourceDirs = append(sourceDirs,
		filepath.Dir(targetDir),
		filepath.Dir(filepath.Dir(targetDir)))
	if wd, err := os.Getwd(); err == nil {
		sourceDirs = append(sourceDirs, wd, filepath.Dir(wd))
	}

	for _, f := range files {
		dst := filepath.Join(targetDir, f.name)
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		// Önce .example dosyasından kopyala (güvenli mock veri)
		for _, dir := range sourceDirs {
			ex := filepath.Join(dir, f.example)
			if data, err := os.ReadFile(ex); err == nil && len(data) > 0 {
				_ = os.WriteFile(dst, data, 0644)
				break
			}
		}
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		// .example yoksa orijinal dosyadan kopyala
		for _, dir := range sourceDirs {
			src := filepath.Join(dir, f.name)
			if data, err := os.ReadFile(src); err == nil && len(data) > 0 {
				_ = os.WriteFile(dst, data, 0644)
				break
			}
		}
	}
}
