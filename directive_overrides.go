/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

// directiveOverrides is a map of directive names to masks used to override any
// masks that are provided to tell the parser how to validate directives and their
// arguments.
//
//nolint:gochecknoglobals
var directiveOverrides = map[string][]uint{
	// if contains a bitmask (ngxConfExpr) that does not exist in NGINX that indicates
	// the directive name is followed by an expression in parantheses.
	"if": {
		ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfBlock | ngxConfExpr | ngxConf1More,
	},
}
