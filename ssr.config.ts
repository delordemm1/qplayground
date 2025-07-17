import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import laravel from "laravel-vite-plugin";

export default defineConfig({
  plugins: [
    laravel({
      input: ["resources/src/app.ts", "resources/css/app.css"],
      ssr: "resources/src/ssr.ts", // Enable SSR
      publicDirectory: "public",
      buildDirectory: "ssr",
      refresh: true,
    }),
    svelte(),
  ],
  build: {
    ssr: true, // Enable SSR
    outDir: "public/ssr",
    rollupOptions: {
      input: "resources/src/ssr.ts",
      output: {
        entryFileNames: "assets/[name].js",
        chunkFileNames: "assets/[name].js",
        assetFileNames: "assets/[name][extname]",
        manualChunks: undefined, // Disable automatic chunk splitting
      },
      external: [
        "$lib/components/RootLayout.svelte",
        "$lib/components/PublicHeader.svelte",
        "$lib/utils/toast",
        "$lib/components/TermsModal.svelte",
        "$lib/components/PolicyModal.svelte",
        "$lib/components/ContestDetailsModal.svelte",
        "$lib/utils/date",
        "$lib/components/ProfileOnboardingModal.svelte",
        "$lib/components/OrganizationOnboardingModal.svelte",
        "$lib/components/UserDashboard.svelte"
      ]
    },
  },
});
