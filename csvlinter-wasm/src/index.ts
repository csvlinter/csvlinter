export interface LintOptions {
  filename?: string;
  delimiter?: string;
  failFast?: boolean;
  inferSchema?: boolean;
  inferSchemaMaxRows?: number;
  schemaContent?: Uint8Array;
}

export interface LintError {
  line_number: number;
  field?: string;
  message: string;
  value?: string;
  type: string;
}

export type LintWarning = LintError;

export interface LintResult {
  file: string;
  total_rows: number;
  errors: LintError[] | null;
  warnings: LintWarning[] | null;
  duration: string;
  valid: boolean;
  schema_used: boolean;
  schema_inferred?: boolean;
  error?: string;
}

export interface CsvlinterInstance {
  validate(csvContent: Uint8Array, options?: LintOptions): LintResult;
}

async function loadWasmExec(url: string): Promise<void> {
  if (typeof (globalThis as Record<string, unknown>)['Go'] !== 'undefined') return;

  if (typeof document === 'undefined') {
    throw new Error('csvlinter-wasm requires a browser environment');
  }

  await new Promise<void>((resolve, reject) => {
    const script = document.createElement('script');
    script.src = url;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`Failed to load wasm_exec.js from ${url}`));
    document.head.appendChild(script);
  });
}

let _instance: CsvlinterInstance | null = null;
let _initPromise: Promise<CsvlinterInstance> | null = null;

export async function createCsvlinter(config?: {
  wasmUrl?: string;
  wasmExecUrl?: string;
}): Promise<CsvlinterInstance> {
  if (_instance) return _instance;
  if (_initPromise) return _initPromise;

  _initPromise = (async () => {
    const wasmUrl = config?.wasmUrl ?? '/csvlinter.wasm';
    const wasmExecUrl = config?.wasmExecUrl ?? '/wasm_exec.js';

    await loadWasmExec(wasmExecUrl);

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const Go = (globalThis as any).Go;
    const go = new Go();

    const result = await WebAssembly.instantiateStreaming(fetch(wasmUrl), go.importObject);

    // go.run() runs Go's main() synchronously until it hits select{}, then returns.
    // By the time this line completes, csvlinterValidate is already registered on globalThis.
    go.run(result.instance);

    _instance = {
      validate(csvContent: Uint8Array, options: LintOptions = {}): LintResult {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return (globalThis as any).csvlinterValidate(csvContent, options) as LintResult;
      },
    };

    return _instance;
  })();

  return _initPromise;
}
