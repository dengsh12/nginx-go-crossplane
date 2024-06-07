package crossplane

var moduleHeardersMoreDirectives = map[string][]uint{
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

func MatchHeardersMore(directive string) ([]uint, bool) {
	masks, matched := moduleHeardersMoreDirectives[directive]
	return masks, matched
}