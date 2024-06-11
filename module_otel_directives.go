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
var moduleOtelDirectives = map[string][]uint{
	"batch_count": {
		ngxConfTake1,
	},
	"batch_size": {
		ngxConfTake1,
	},
	"endpoint": {
		ngxConfTake1,
	},
	"interval": {
		ngxConfTake1,
	},
	"otel_exporter": {
		ngxHTTPMainConf | ngxConfBlock | ngxConfNoArgs,
	},
	"otel_service_name": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"otel_span_attr": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
	},
	"otel_span_name": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"otel_trace": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"otel_trace_context": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
}

func MatchOtel(directive string) ([]uint, bool) {
	masks, matched := moduleOtelDirectives[directive]
	return masks, matched
}