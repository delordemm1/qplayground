import { createInertiaApp } from "@inertiajs/svelte";
import { mount } from "svelte";
import RootLayout from "./lib/components/RootLayout.svelte"; // A fallback root layout
import ErrorPage from "./routes/+error.svelte"; // A fallback error page
import "../css/app.css";

createInertiaApp({
  /**
   * Resolves the page component and its layout based on the page name,
   * mimicking SvelteKit's file-based routing structure.
   *
   * @param name The name of the page from Inertia (e.g., "dashboard").
   */
  resolve: (name) => {
    // Use eager importing for synchronous resolution.
    // @ts-expect-error
    const pages = import.meta.glob("./routes/**/+page.svelte", { eager: true });
    // @ts-expect-error
    const layouts = import.meta.glob("./routes/**/+layout.svelte", {
      eager: true,
    });

    // Normalize name: remove leading/trailing slashes.
    const normalizedName = name.replace(/^\/+|\/+$/g, "");

    let foundPagePath = null;

    // Find the correct page file by matching the Inertia name against
    // file paths after stripping out route groups `(...)`.
    for (const path in pages) {
      // Create a "routable" path by removing route group segments.
      // e.g., './routes/(app)/dashboard/+page.svelte' becomes './routes/dashboard/+page.svelte'
      const routablePath = path.replace(/\/\([^)]+\)\//g, "/");

      if (
        `./routes/${normalizedName ? normalizedName + "/" : ""}+page.svelte` ===
        routablePath
      ) {
        foundPagePath = path;
        break;
      }
    }

    // If the page component doesn't exist, return a dedicated error page.
    if (!foundPagePath) {
      console.error(`Page component not found for name: ${name}`);
      return { default: ErrorPage, layout: RootLayout };
    }

    const pageModule = pages[foundPagePath] as any;

    // --- Layout Resolution ---
    // Start searching for the nearest `+layout.svelte` file from the page's
    // actual file path, traversing upwards.
    let layoutComponent: any = null;
    let currentPath = foundPagePath.substring(
      0,
      foundPagePath.lastIndexOf("/")
    );

    while (currentPath.startsWith("./routes")) {
      const layoutPath = `${currentPath}/+layout.svelte`;
      if (layouts[layoutPath]) {
        const layoutModule = layouts[layoutPath] as any;
        layoutComponent = layoutModule.default;
        break; // Found the nearest layout, so we stop searching.
      }
      // Move up one directory.
      const parentPath = currentPath.substring(0, currentPath.lastIndexOf("/"));
      if (parentPath === currentPath) break; // Reached the top without finding anything.
      currentPath = parentPath;
    }

    // Determine the final layout.
    // If the page exports `noLayout: true`, bypass all layout logic.
    const finalLayout = pageModule.noLayout
      ? null
      : pageModule.layout || layoutComponent || RootLayout;

    return { default: pageModule.default, layout: finalLayout };
  },

  /**
   * Sets up the Svelte application.
   */
  setup({ el, App, props }) {
    console.log("[App] setup", { props });
    // Svelte 5's `mount` handles both initial mounting and hydration.
    mount(App, { target: el!, props });
  },

  /**
   * Configures the progress indicator.
   */
  progress: {
    delay: 250,
    color: "#29d",
    includeCSS: true,
    showSpinner: false,
  },
});
