// Shared Vite config for every web/ app; run from the app's directory.
import { realpathSync } from "node:fs";
import { dirname, join } from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// Vite 8 bundles with Rolldown (native), which resolves symlinks itself: in Bazel's sandbox
// the inputs are symlinks into the execroot, so Rolldown sees their real paths. Give Vite
// the real directory of index.html as its root, and write the bundle to the sandbox path
// Bazel collects outputs from.
const cwd = process.cwd();

export default defineConfig({
  root: dirname(realpathSync(join(cwd, "index.html"))),
  plugins: [react()],
  logLevel: "warn",
  build: {
    outDir: join(cwd, "bundle"),
    emptyOutDir: true,
  },
});
