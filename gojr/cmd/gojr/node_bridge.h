#ifndef GOJR_CMD_GOJR_NODE_BRIDGE_H
#define GOJR_CMD_GOJR_NODE_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct gojr_node_runtime gojr_node_runtime;

gojr_node_runtime* gojr_node_new(const char* module_bundle_json, char** error_out);
char* gojr_node_eval(gojr_node_runtime* runtime, const char* source, char** error_out);
char* gojr_node_set_sheet(gojr_node_runtime* runtime, const char* json, char** error_out);
void gojr_node_free(gojr_node_runtime* runtime);
void gojr_string_free(char* value);

#ifdef __cplusplus
}
#endif

#endif
