import type { GoJuniorPackageExportData } from "../build.js";
import type { Diagnostic } from "../diagnostics.js";

export interface EmitterContextOptions {
  artifact: GoJuniorPackageExportData;
}

export class EmitterContext {
  public readonly diagnostics: Diagnostic[] = [];
  private nextSymbolID = 0;

  public constructor(public readonly options: EmitterContextOptions) {}

  public symbol(prefix: string): string {
    const clean = prefix.replace(/[^A-Za-z0-9_$]/g, "_").replace(/^[^A-Za-z_$]/, "_");
    this.nextSymbolID += 1;
    return `__gojr_${clean || "sym"}_${this.nextSymbolID}`;
  }

  public emitError(message: string): void {
    this.diagnostics.push({
      filename: this.options.artifact.sources[0]?.filename ?? "gojr-emitter.go",
      code: "GOJR_EMIT001",
      severity: "error",
      message
    });
  }
}
