import { defineConfig } from "vite";
import laravel from "laravel-vite-plugin";
import { svelte, vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
    plugins: [
        laravel({
            input: "resources/src/app.ts",
            publicDirectory: "public",
            buildDirectory: "build",
            refresh: true,
        }),
        tsconfigPaths(),
        svelte({
            preprocess: [vitePreprocess()],
            compilerOptions: {
                runes: true,
            },
            dynamicCompileOptions({ filename }) {
                if (
                    filename.includes("svelte-french-toast") ||
                    // filename.includes("@tanstack/svelte-query") ||
                    filename.includes("@inertiajs")
                ) {
                    return { runes: undefined }; // or false, check what works
                }
            },
        }),
        tailwindcss(),
    ],
    build: {
        manifest: true, // Generate manifest.json file
        outDir: "public/build",
        emptyOutDir: true,

        // rollupOptions: {
        //   input: "resources/js/app.js",
        //   output: {
        //     entryFileNames: "assets/[name].js",
        //     chunkFileNames: "assets/[name].js",
        //     assetFileNames: "assets/[name].[ext]",
        //     manualChunks: undefined, // Disable automatic chunk splitting
        //   },
        // },
    },
    server: {
        hmr: {
            host: "localhost",
        },
        host: "localhost",
        port: 3200,
    },
});
