#!/bin/bash
set -eu
cd "$(dirname "$0")/../.."
files=(network.go network_aig.go hierarchy.go abc.go network_api.go abcFanio.go abcObj.go abcRefs.go abcAig.go abcBarBuf.go abcBlifMv.go abcCheck.go mv_table.go mv_evaluate.go abcDfs.go dfs_iso.go abcFanOrder.go abcFunc.go abcHie.go abcHieCec.go scl_adapter.go fxu_adapter.go abcSop.go abcNtk.go ntk_core.go induction_model.go abcSop_reference_test.go abcSop_test.go aig_manager_test.go barbuf_test.go blifmv_test.go check_test.go dfs_test.go fanio_test.go fanorder_test.go function_test.go fxu_adapter_test.go hie_test.go hiecec_test.go induction_model_test.go network_api_test.go network_test.go ntk_c_test.go ntk_core_test.go object_test.go refs_test.go scl_adapter_test.go)
for i in "${!files[@]}"; do files[i]="./base/abc/${files[i]}"; done
GOCACHE=/tmp/abcgo-cache go test "$@" "${files[@]}"
