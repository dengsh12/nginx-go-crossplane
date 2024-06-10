package crossplane

//nolint:gochecknoglobals
var moduleLuaDirectives = map[string][]uint{
	"access_by_lua": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"access_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfBlock | ngxConfNoArgs,
	},
	"access_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"access_by_lua_no_postpone": {
		ngxHTTPMainConf | ngxConfFlag,
	},
	"balancer_by_lua_block": {
		ngxHTTPUpsConf | ngxConfBlock | ngxConfNoArgs,
	},
	"balancer_by_lua_file": {
		ngxHTTPUpsConf | ngxConfTake1,
	},
	"body_filter_by_lua": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"body_filter_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfBlock | ngxConfNoArgs,
	},
	"body_filter_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"content_by_lua": {
		ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"content_by_lua_block": {
		ngxHTTPLocConf | ngxHTTPLifConf | ngxConfBlock | ngxConfNoArgs,
	},
	"content_by_lua_file": {
		ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"exit_worker_by_lua_block": {
		ngxHTTPMainConf | ngxConfBlock | ngxConfNoArgs,
	},
	"exit_worker_by_lua_file": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"header_filter_by_lua": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"header_filter_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfBlock | ngxConfNoArgs,
	},
	"header_filter_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"init_by_lua": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"init_by_lua_block": {
		ngxHTTPMainConf | ngxConfBlock | ngxConfNoArgs,
	},
	"init_by_lua_file": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"init_worker_by_lua": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"init_worker_by_lua_block": {
		ngxHTTPMainConf | ngxConfBlock | ngxConfNoArgs,
	},
	"init_worker_by_lua_file": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"log_by_lua": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"log_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfBlock | ngxConfNoArgs,
	},
	"log_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_capture_error_log": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_check_client_abort": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_code_cache": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_http10_buffering": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_load_resty_core": {
		ngxHTTPMainConf | ngxConfFlag,
	},
	"lua_malloc_trim": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_max_pending_timers": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_max_running_timers": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_need_request_body": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_package_cpath": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_package_path": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_regex_cache_max_entries": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_regex_match_limit": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_sa_restart": {
		ngxHTTPMainConf | ngxConfFlag,
	},
	"lua_shared_dict": {
		ngxHTTPMainConf | ngxConfTake2,
	},
	"lua_socket_buffer_size": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_socket_connect_timeout": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_socket_keepalive_timeout": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_socket_log_errors": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_socket_pool_size": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_socket_read_timeout": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_socket_send_lowat": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_socket_send_timeout": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"lua_ssl_certificate": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"lua_ssl_certificate_key": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"lua_ssl_ciphers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"lua_ssl_conf_command": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
	},
	"lua_ssl_crl": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"lua_ssl_protocols": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConf1More,
	},
	"lua_ssl_trusted_certificate": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"lua_ssl_verify_depth": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
	},
	"lua_thread_cache_max_entries": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"lua_transform_underscores_in_response_headers": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_use_default_type": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfFlag,
	},
	"lua_worker_thread_vm_pool_size": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"rewrite_by_lua": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"rewrite_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfBlock | ngxConfNoArgs,
	},
	"rewrite_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1,
	},
	"rewrite_by_lua_no_postpone": {
		ngxHTTPMainConf | ngxConfFlag,
	},
	"server_rewrite_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxConfBlock | ngxConfNoArgs,
	},
	"server_rewrite_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxConfTake1,
	},
	"set_by_lua": {
		ngxHTTPSrvConf | ngxHTTPSifConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf2More,
	},
	"set_by_lua_block": {
		ngxHTTPSrvConf | ngxHTTPSifConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake1 | ngxConfBlock,
	},
	"set_by_lua_file": {
		ngxHTTPSrvConf | ngxHTTPSifConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf2More,
	},
	"ssl_certificate_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxConfBlock | ngxConfNoArgs,
	},
	"ssl_certificate_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxConfTake1,
	},
	"ssl_client_hello_by_lua_block": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxConfBlock | ngxConfNoArgs,
	},
	"ssl_client_hello_by_lua_file": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxConfTake1,
	},
	"ssl_session_fetch_by_lua_block": {
		ngxHTTPMainConf | ngxConfBlock | ngxConfNoArgs,
	},
	"ssl_session_fetch_by_lua_file": {
		ngxHTTPMainConf | ngxConfTake1,
	},
	"ssl_session_store_by_lua_block": {
		ngxHTTPMainConf | ngxConfBlock | ngxConfNoArgs,
	},
	"ssl_session_store_by_lua_file": {
		ngxHTTPMainConf | ngxConfTake1,
	},
}

func MatchLua(directive string) ([]uint, bool) {
	masks, matched := moduleLuaDirectives[directive]
	return masks, matched
}
