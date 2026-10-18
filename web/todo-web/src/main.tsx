import { ApiClient } from "@suite/api-client";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { App } from "./App.js";

const client = new ApiClient("/api", localStorage.getItem("token") ?? "");

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App client={client} />
  </StrictMode>,
);
