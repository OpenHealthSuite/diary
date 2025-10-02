import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
// This is annoying, we can just ignore it though
// @ts-ignore ts(1259)
import path from "node:path";
import autoPreprocess from "svelte-preprocess";
const proxy = {
  "/api": {
    target: "http://localhost:8080",
    changeOrigin: true,
    headers: {
      "X-OpenFoodDiary-UserId":
        process.env.OPENFOODDIARY_USERID ??
        "f1750ac3-d6cc-4981-9466-f1de2ebbad33"
    },
    secure: false,
    ws: true
  }
};

// https://vitejs.dev/config/
export default defineConfig({
  preview: {
    port: 3000
  },
  server: {
    port: 3000,
    proxy
  },
  resolve: {
    alias: {
      src: path.resolve("./src"),
    }
  },

  optimizeDeps: {
    esbuildOptions: {
      plugins: [
      ]
    }
  },
  plugins: [svelte({
    preprocess: autoPreprocess()
  })]
});
