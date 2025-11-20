/*
 * Copyright 2025 - IBM Corporation. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package info

var (
	// gitCommit is the git commit hash, set via ldflags during build
	gitCommit = "unknown"
	// version is the version string, set via ldflags during build
	version = "unknown"
)

// GetGitCommit returns the git commit hash
func GetGitCommit() string {
	return gitCommit
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}

// Made with Bob
