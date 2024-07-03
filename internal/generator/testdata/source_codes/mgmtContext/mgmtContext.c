static ngx_command_t ngx_mgmt_block_commands[] = {

    { ngx_string("auto_append_mgmt_context"),
      NGX_CONF_TAKE2,
      0,
      0,
      0,
      NULL },

    ngx_null_command
};