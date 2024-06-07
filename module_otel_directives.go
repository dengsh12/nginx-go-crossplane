package crossplane

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