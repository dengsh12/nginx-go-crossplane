// This is generated code, don't modify it.
// If you want to overwrite any directive's definition, please modify forced_directives_map.go
// All the definitions are generated from the source code you provided
// Each bit mask describes these behaviors:
//   - how many arguments the directive can take
//   - whether or not it is a block directive
//   - whether this is a flag (takes one argument that's either "on" or "off")
//   - which contexts it's allowed to be in

package crossplane

//nolint:gochecknoglobals
var moduleHeadersMoreDirectives = map[string][]uint{
	"more_clear_headers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
	},
	"more_clear_input_headers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
	},
	"more_set_headers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
	},
	"more_set_input_headers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
	},
}

func MatchHeadersMore(directive string) ([]uint, bool) {
	masks, matched := moduleHeadersMoreDirectives[directive]
	return masks, matched
}
