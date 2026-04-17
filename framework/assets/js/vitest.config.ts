import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "jsdom",
    include: ["htmxextensions/__tests__/**/*.test.ts"],
  },
});
