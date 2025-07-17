import RootLayout from "$lib/components/RootLayout.svelte";
import { createInertiaApp } from "@inertiajs/svelte";
import createServer from "@inertiajs/svelte/server";
import { mount } from "svelte";
import ErrorPage from "./routes/+error.svelte"; // A fallback error page

createServer((page) =>
  createInertiaApp({
    /**
     * Resolves the page component and its layout based on the page name,
     * mimicking SvelteKit's file-based routing structure.
     *
     * @param name The name of the page from Inertia (e.g., "Users/Index").
     */
    resolve: (name) => {
      // Use eager importing for synchronous resolution.
      // @ts-expect-error
      const pages = import.meta.glob("./routes/**/+page.svelte", {
        eager: true,
      });
      // @ts-expect-error
      const layouts = import.meta.glob("./routes/**/+layout.svelte", {
        eager: true,
      });

      // Normalize name: remove leading/trailing slashes to handle cases like "", "/", or "/Users/Index/"
      const normalizedName = name.replace(/^\/+|\/+$/g, "");

      // Construct the expected path for the page component.
      // If normalizedName is empty, it correctly resolves to the root page: './routes/+page.svelte'
      const pagePath = `./routes/${
        normalizedName ? normalizedName + "/" : ""
      }+page.svelte`;
      const pageModule = pages[pagePath] as any;

      // If the page component doesn't exist, return a dedicated error page.
      if (!pageModule) {
        console.error(`Page component not found for name: ${name}`);
        return { default: ErrorPage, layout: RootLayout };
      }

      // --- Layout Resolution ---
      // Start searching for the nearest `+layout.svelte` file from the page's
      // directory upwards to the root of the `routes` directory.
      let layoutComponent: any = null;
      const pathSegments = name.split("/").filter(Boolean);

      // Iterate from the most specific path to the most general.
      // e.g., for "Users/Profile/Edit", it checks:
      // 1. ./routes/Users/Profile/Edit/+layout.svelte
      // 2. ./routes/Users/Profile/+layout.svelte
      // 3. ./routes/Users/+layout.svelte
      // 4. ./routes/+layout.svelte
      for (let i = pathSegments.length; i >= 0; i--) {
        const subPath = pathSegments.slice(0, i).join("/");
        const layoutPath = `./routes/${
          subPath ? subPath + "/" : ""
        }+layout.svelte`;

        if (layouts[layoutPath]) {
          const layoutModule = layouts[layoutPath] as any;
          layoutComponent = layoutModule.default;
          break; // Found the nearest layout, so we stop searching.
        }
      }

      // Determine the final layout with the following priority:
      // 1. A layout exported from the page itself (`export const layout = ...`).
      // 2. The nearest `+layout.svelte` found in the directory structure.
      // 3. The global `RootLayout` as a final fallback.
      const finalLayout = pageModule.layout || layoutComponent || RootLayout;

      return { default: pageModule.default, layout: finalLayout };
    },
    page,
    /**
   * Sets up the Svelte application.
   */
  setup({ el, App, props }) {
    console.log("[DEBUG] app.ts props", props);
    // Svelte 5's `mount` handles both initial mounting and hydration.
    mount(App, { target: el!, props });
  },
  })
);
