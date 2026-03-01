// bindings.ts — Deno-side runtime for Molt JS/TS functions.
// Embedded in the Go binary via //go:embed. Runs inside each Deno worker process.
//
// Protocol: JSON lines over stdin/stdout.
// - Go sends "invoke" messages with a request
// - JS sends "host_call" messages for env.fetch/storage/crawl (Go executes, returns "host_result")
// - JS sends "response" message with the handler's Response

const encoder = new TextEncoder();
const decoder = new TextDecoder();

let callId = 0;
const pending = new Map<
  number,
  { resolve: (value: unknown) => void; reject: (reason: unknown) => void }
>();

// Buffered stdin reader
const stdinBuf: string[] = [];
let stdinRemainder = "";

async function readLine(): Promise<string> {
  if (stdinBuf.length > 0) {
    return stdinBuf.shift()!;
  }

  // Read from stdin until we get a complete line
  const buf = new Uint8Array(65536);
  while (true) {
    const n = await Deno.stdin.read(buf);
    if (n === null) {
      throw new Error("stdin closed");
    }

    stdinRemainder += decoder.decode(buf.subarray(0, n));
    const lines = stdinRemainder.split("\n");

    if (lines.length > 1) {
      stdinRemainder = lines.pop()!;
      stdinBuf.push(...lines.slice(1));
      return lines[0];
    }
  }
}

async function readMessage(): Promise<Record<string, unknown>> {
  const line = await readLine();
  return JSON.parse(line);
}

function writeMessage(msg: Record<string, unknown>): void {
  const json = JSON.stringify(msg) + "\n";
  Deno.stdout.writeSync(encoder.encode(json));
}

// Host call: send request to Go, await response
async function hostCall(fn: string, args: Record<string, unknown>): Promise<unknown> {
  const id = ++callId;
  return new Promise((resolve, reject) => {
    pending.set(id, { resolve, reject });
    writeMessage({ type: "host_call", id, fn, args });
  });
}

// Declare env type for TypeScript
interface MoltEnv {
  fetch(url: string, options?: RequestInit): Promise<Response>;
  storage: {
    put(bucket: string, key: string, data: string, contentType?: string): Promise<unknown>;
    get(bucket: string, key: string): Promise<{ data: string; content_type: string; size: number }>;
    delete(bucket: string, key: string): Promise<void>;
    list(bucket: string, prefix?: string): Promise<unknown>;
  };
  crawl: {
    page(url: string, opts?: { selectors?: string[]; screenshot?: boolean; js?: boolean }): Promise<unknown>;
  };
}

// Build the env global
const env: MoltEnv = {
  async fetch(url: string, options?: RequestInit): Promise<Response> {
    const method = options?.method ?? "GET";
    const headers: Record<string, string> = {};
    if (options?.headers) {
      if (options.headers instanceof Headers) {
        options.headers.forEach((v, k) => { headers[k] = v; });
      } else if (Array.isArray(options.headers)) {
        for (const [k, v] of options.headers) { headers[k] = v; }
      } else {
        Object.assign(headers, options.headers);
      }
    }

    let body: string | undefined;
    if (options?.body) {
      if (typeof options.body === "string") {
        body = btoa(options.body);
      } else if (options.body instanceof ArrayBuffer || options.body instanceof Uint8Array) {
        const bytes = options.body instanceof Uint8Array ? options.body : new Uint8Array(options.body);
        body = btoa(String.fromCharCode(...bytes));
      } else {
        body = btoa(String(options.body));
      }
    }

    const result = (await hostCall("http_request", {
      method,
      url,
      headers,
      body,
    })) as { status: number; headers: Record<string, string>; body: string };

    const respBody = result.body ? Uint8Array.from(atob(result.body), (c) => c.charCodeAt(0)) : null;
    return new Response(respBody, {
      status: result.status,
      headers: result.headers,
    });
  },

  storage: {
    async put(bucket: string, key: string, data: string, contentType?: string): Promise<unknown> {
      return await hostCall("storage_put", {
        bucket,
        key,
        data: btoa(data),
        content_type: contentType ?? "application/octet-stream",
      });
    },
    async get(bucket: string, key: string): Promise<{ data: string; content_type: string; size: number }> {
      return (await hostCall("storage_get", { bucket, key })) as {
        data: string;
        content_type: string;
        size: number;
      };
    },
    async delete(bucket: string, key: string): Promise<void> {
      await hostCall("storage_delete", { bucket, key });
    },
    async list(bucket: string, prefix?: string): Promise<unknown> {
      return await hostCall("storage_list", { bucket, prefix });
    },
  },

  crawl: {
    async page(url: string, opts?: { selectors?: string[]; screenshot?: boolean; js?: boolean }): Promise<unknown> {
      return await hostCall("crawl_page", { url, ...opts });
    },
  },
};

// Install env as a global
(globalThis as unknown as { env: MoltEnv }).env = env;

// Module cache: import user scripts only once
const moduleCache = new Map<string, { default: (req: Request) => Promise<Response> | Response }>();

async function loadModule(scriptPath: string) {
  if (moduleCache.has(scriptPath)) {
    return moduleCache.get(scriptPath)!;
  }
  const mod = await import(scriptPath);
  moduleCache.set(scriptPath, mod);
  return mod;
}

// Main loop
async function main(): Promise<void> {
  while (true) {
    let msg: Record<string, unknown>;
    try {
      msg = await readMessage();
    } catch {
      // stdin closed — exit gracefully
      break;
    }

    if (msg.type === "shutdown") {
      break;
    }

    if (msg.type === "host_result") {
      // Resolve a pending host_call promise
      const id = msg.id as number;
      const p = pending.get(id);
      if (p) {
        pending.delete(id);
        if (msg.error) {
          p.reject(new Error(msg.error as string));
        } else {
          p.resolve(msg.data);
        }
      }
      continue;
    }

    if (msg.type === "invoke") {
      const data = msg.data as { script_path: string; request: { method: string; url: string; headers?: Record<string, string>; body?: string } };

      try {
        const userModule = await loadModule(data.script_path);
        const handler = userModule.default;

        if (typeof handler !== "function") {
          writeMessage({
            type: "response",
            status: 500,
            body: "module does not export a default function handler",
          });
          continue;
        }

        // Build Request object
        let reqBody: BodyInit | null = null;
        if (data.request.body) {
          reqBody = Uint8Array.from(atob(data.request.body), (c) => c.charCodeAt(0));
        }

        const request = new Request(data.request.url || "http://localhost/", {
          method: data.request.method || "GET",
          headers: data.request.headers,
          body: reqBody,
        });

        const response = await handler(request);

        // Read response
        const body = await response.text();
        const headers: Record<string, string> = {};
        response.headers.forEach((v: string, k: string) => {
          headers[k] = v;
        });

        writeMessage({
          type: "response",
          status: response.status,
          headers,
          body,
        });
      } catch (e: unknown) {
        const errMsg = e instanceof Error ? e.message : String(e);
        writeMessage({
          type: "response",
          status: 500,
          body: errMsg,
          error: errMsg,
        });
      }
      continue;
    }

    if (msg.type === "ping") {
      writeMessage({ type: "pong" });
      continue;
    }
  }
}

main();
