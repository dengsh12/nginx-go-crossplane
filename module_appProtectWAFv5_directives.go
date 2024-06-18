// nginx app protect specific and global directives
// this is human-derived, since source code is not available
package crossplane

//nolint:gochecknoglobals
var appProtectWAFv5Directives = map[string][]uint{
	// https://docs.nginx.com/nginx-app-protect-waf/v5/configuration-guide/configuration/#global-directives
	"app_protect_physical_memory_util_thresholds": {
		ngxHTTPMainConf | ngxConfTake2,
	},
	"app_protect_cpu_thresholds": {
		ngxHTTPMainConf | ngxConfTake2,
	},
	"app_protect_failure_mode_action": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"app_protect_cookie_seed": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"app_protect_request_buffer_overflow_action": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"app_protect_reconnect_period_seconds": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	// https://docs.nginx.com/nginx-app-protect-waf/v5/configuration-guide/configuration/#app-protect-specific-directives
	"app_protect_enforcer_address": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"app_protect_enable": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfFlag,
	},
	"app_protect_policy_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"app_protect_security_log_enable": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfFlag,
	},
	"app_protect_security_log": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
	},
	"app_protect_custom_log_attribute": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
	},
}

// AppProtectWAFv4DirectivesMatchFn is a match function for parsing an NGINX config that contains the
// App Protect v5 module.
func AppProtectWAFv5DirectivesMatchFn(directive string) ([]uint, bool) {
	masks, matched := appProtectWAFv5Directives[directive]
	return masks, matched
}
