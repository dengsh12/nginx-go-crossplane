package crossplane

var moduleNjsDirectives = map[string][]uint{
	"js_access": {
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_body_filter": {
		ngxHTTPLocConf | ngxHTTPLifConf | ngxHTTPLmtConf | ngxConfTake12,
	},
	"js_content": {
		ngxHTTPLocConf | ngxHTTPLifConf | ngxHTTPLmtConf | ngxConfTake1,
	},
	"js_fetch_buffer_size": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_fetch_ciphers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_fetch_max_response_buffer_size": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_fetch_protocols": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConf1More,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConf1More,
	},
	"js_fetch_timeout": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_fetch_trusted_certificate": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_fetch_verify": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfFlag,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfFlag,
	},
	"js_fetch_verify_depth": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_filter": {
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_header_filter": {
		ngxHTTPLocConf | ngxHTTPLifConf | ngxHTTPLmtConf | ngxConfTake1,
	},
	"js_import": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake13,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake13,
	},
	"js_path": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_periodic": {
		ngxHTTPLocConf | ngxConfAny,
		ngxStreamSrvConf | ngxConfAny,
	},
	"js_preload_object": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake13,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake13,
	},
	"js_preread": {
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake1,
	},
	"js_set": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake2,
	},
	"js_shared_dict_zone": {
		ngxHTTPMainConf | ngxConf1More,
		ngxStreamMainConf | ngxConf1More,
	},
	"js_var": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake12,
		ngxStreamMainConf | ngxStreamSrvConf | ngxConfTake12,
	},
}

func MatchNjs(directive string) ([]uint, bool) {
	masks, matched := moduleNjsDirectives[directive]
	return masks, matched
}