/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	"io"
)

// Generate receives a string sourcePath and an io.Writer writer. It will
// extract all the directives definitions from the .c and .cpp files in
// sourcePath and its subdirectories, then output the corresponding directive
// definition map and matchFunc via writer.
func Generate(sourcePath string, writer io.Writer) error {
	return genFromSrcCode(sourcePath, "directives", "Match", writer)
}
