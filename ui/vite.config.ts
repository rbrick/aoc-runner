import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// in development, the API is served by `aoc-runner http`
const api = "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/events": api,
      "/submit": api,
      "/auth": api,
    },
  },
});
