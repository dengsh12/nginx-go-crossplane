/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	"os"
)

func Generate(sourcePath string) error {
	return genSupFromSrcCode(sourcePath, "directives", "Match", os.Stdout)
}
