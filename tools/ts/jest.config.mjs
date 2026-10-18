// Shared Jest config for every ts_package test. Tests run on tsc's ESM output (dist/).
import * as path from "path";

const runfile = (p) => path.join(process.env.JS_BINARY__RUNFILES, process.env.JS_BINARY__WORKSPACE, p);

export default {
  testEnvironment: "jsdom",
  testMatch: ["**/dist/**/*.test.js"],
  setupFilesAfterEnv: [runfile("tools/ts/jest.setup.mjs")],
};
