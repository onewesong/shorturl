import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

const backendTarget =
  process.env.ADMIN_DEV_PROXY_TARGET ||
  process.env.VITE_BACKEND_URL ||
  `http://localhost:${process.env.PORT || "8080"}`;

export default defineConfig({
  base: "/admin/",
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/admin/api": backendTarget,
      "/healthz": backendTarget,
      "/go": backendTarget,
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: "./src/setupTests.ts",
  },
});
