/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	g := NewWithT(t)

	// Test GetNonEmptyLines
	lines := GetNonEmptyLines("line1\n\nline2\n\nline3\n")
	g.Expect(lines).To(Equal([]string{"line1", "line2", "line3"}))

	emptyLines := GetNonEmptyLines("")
	g.Expect(emptyLines).To(BeEmpty())

	// Test GetProjectDir
	dir, err := GetProjectDir()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(dir).NotTo(BeEmpty())

	// Test warnError
	warnError(errors.New("test warning"))

	// Test UncommentCode
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	initialContent := "line 1\n# line 2\n# line 3\nline 4\n"
	err = os.WriteFile(testFile, []byte(initialContent), 0o644)
	g.Expect(err).NotTo(HaveOccurred())

	err = UncommentCode(testFile, "# line 2\n# line 3", "# ")
	g.Expect(err).NotTo(HaveOccurred())

	uncommentedContent, err := os.ReadFile(testFile)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(uncommentedContent)).To(Equal("line 1\nline 2\nline 3\nline 4\n"))

	// Test UncommentCode errors
	err = UncommentCode("non-existent-file.txt", "target", "#")
	g.Expect(err).To(HaveOccurred())

	err = UncommentCode(testFile, "non-existent-target", "#")
	g.Expect(err).To(HaveOccurred())
}
