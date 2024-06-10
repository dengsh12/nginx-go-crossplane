package crossplane

var forced_directives_map = map[string][]uint{
	"if": {
		ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfBlock | ngxConfExpr | ngxConf1More,
	},
}
