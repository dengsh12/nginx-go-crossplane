/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

// A map used to overwrite generated directive definitions.
// When we have a human-defined definition for a directive, which is different
// from the definition in source code, put it here.
//
//nolint:gochecknoglobals
var priorityMap = map[string][]uint{
	"if": {
		ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfBlock | ngxConfExpr | ngxConf1More,
	},
}
