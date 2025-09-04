package testutil

import (
	"os"
	"path/filepath"
	"runtime"
)

// SetRootPath changes the working directory to the project root.
// This is crucial for tests that need to access files like .env.
func SetRootPath() {
	// Get the path of the current file.
	_, b, _, _ := runtime.Caller(0)
	// Go up three levels from internal/testutil/helper.go to the project root.
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..")

	// Change the working directory to the project root.
	if err := os.Chdir(projectRoot); err != nil {
		panic("could not change to project root directory for testing: " + err.Error())
	}
}
