package gcp

import (
	"fmt"
	"log"
	"os/exec"
)

// Shell runs command on shell and get back output and error if get one
func Shell(format string, args ...interface{}) (string, error) {
	return sh(format, false, args...)
}

// ShellSilent runs command on shell without logging the exact command
// useful when command involves secrets
func ShellSilent(format string, args ...interface{}) (string, error) {
	return sh(format, true, args...)
}

// Runs command on shell and get back output and error if get one
func sh(format string, muted bool, args ...interface{}) (string, error) {
	command := fmt.Sprintf(format, args...)
	if !muted {
		log.Printf("Running command %s", command)
	}
	c := exec.Command("sh", "-c", command) // #nosec
	b, err := c.CombinedOutput()
	if !muted {
		log.Printf("Command output: \n%s", string(b))
	}
	if err != nil {
		return "", fmt.Errorf("command failed: %q %v", string(b), err)
	}
	return string(b), nil
}
