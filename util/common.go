package util

import "os"

// FileExists is simple wrapper for check file existance.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
