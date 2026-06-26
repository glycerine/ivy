export class EmitterContext {
    options;
    diagnostics = [];
    nextSymbolID = 0;
    constructor(options) {
        this.options = options;
    }
    symbol(prefix) {
        const clean = prefix.replace(/[^A-Za-z0-9_$]/g, "_").replace(/^[^A-Za-z_$]/, "_");
        this.nextSymbolID += 1;
        return `__gojr_${clean || "sym"}_${this.nextSymbolID}`;
    }
    emitError(message) {
        this.diagnostics.push({
            filename: this.options.artifact.sources[0]?.filename ?? "gojr-emitter.go",
            code: "GOJR_EMIT001",
            severity: "error",
            message
        });
    }
}
