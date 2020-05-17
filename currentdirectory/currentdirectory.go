package currentdirectory

import (
	"os"
	"path/filepath"
	"strings"
)

func GetCurrentDirectory() (*string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	s := strings.Replace(dir, "\\", "/", -1)
	return &s, nil
}
