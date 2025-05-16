package sevenzip

import (
	"errors"
	"fmt"
	"os/exec"
)

type SevenZipCompressor struct{}

func (c *SevenZipCompressor) Compress(inputFile string, outputFile string, params map[string]interface{}) error {
	// Basic implementation using the external 7z command-line tool.
	// Assumes 7z is installed and in the system's PATH.
	// This is a simplified example; a robust implementation would handle
	// variations in 7z availability, different operating systems, and error parsing.

	// Construct the 7z command. You might add more options based on 'params' later.
	// For example, 'params' could include compression level, method, etc.
	cmd := exec.Command("7z", "a", "-y", outputFile, inputFile)

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("7z compression failed: %w", err)
	}

	// Check for execution errors of the 7z command itself (e.g., command not found)
	if !cmd.ProcessState.Success() {
		return errors.New("7z command did not run successfully")
	}

	return nil
}