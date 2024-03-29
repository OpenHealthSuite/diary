import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
// This is annoying, we can just ignore it though
// @ts-ignore ts(1259)
import path from "node:path";
import rollupNodePolyFill from "rollup-plugin-node-polyfills";
import autoPreprocess from "svelte-preprocess";
import { NodeGlobalsPolyfillPlugin } from "@esbuild-plugins/node-globals-polyfill";
import { NodeModulesPolyfillPlugin } from "@esbuild-plugins/node-modules-polyfill";
const proxy = {
  "/api": {
    target: "http://localhost:3012",
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
      stream: "rollup-plugin-node-polyfills/polyfills/stream",
      buffer: "rollup-plugin-node-polyfills/polyfills/buffer-es6"
    }
  },

  optimizeDeps: {
    esbuildOptions: {
      plugins: [
        // @ts-ignore
        NodeGlobalsPolyfillPlugin({
          buffer: true
        }),
        // @ts-ignore
        NodeModulesPolyfillPlugin()

      ]
    }
  },
  plugins: [rollupNodePolyFill(), svelte({
    preprocess: autoPreprocess()
  })]
});
