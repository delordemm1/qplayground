<script>
  import { inertia } from "@inertiajs/svelte";
  
  let y = $state(0);
  let mobileMenuOpen = $state(false);
  
  let links = $state([
    { name: "Home", href: "/", isButton: false, isActive: false },
    { name: "Products", href: "/products", isButton: false, isActive: false },
    { name: "Projects", href: "/projects", isButton: false, isActive: false },
    { name: "About", href: "/about", isButton: false, isActive: false },
    { name: "Blog", href: "/blog", isButton: false, isActive: false },
    { name: "Contact", href: "/contact", isButton: false, isActive: false },
    { name: "Get Started", href: "/auth", isButton: true, isActive: false },
  ]);

  $effect(() => {
    // Set the active link based on current pathname
    const currentPath = window.location.pathname;
    links.forEach((link) => {
      if (link.href === currentPath || (link.href === "/" && currentPath === "/")) {
        link.isActive = true;
      } else {
        link.isActive = false;
      }
    });
  });

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  function closeMenu() {
    mobileMenuOpen = false;
  }

  $effect(() => {
    document.body.style.overflow = mobileMenuOpen ? "hidden" : "auto";
  });
</script>

<svelte:window bind:scrollY={y} />

<header
  class="sticky top-0 z-50 w-full border-b border-solid border-[var(--border-color)] bg-[var(--background-dark)]/80 backdrop-blur-md"
>
  <div
    class="container mx-auto flex items-center justify-between whitespace-nowrap px-6 py-4"
  >
    <!-- Logo -->
    <div class="flex items-center gap-3">
      <svg
        class="h-8 w-8 text-[var(--primary-color)]"
        fill="none"
        viewBox="0 0 48 48"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          clip-rule="evenodd"
          d="M39.475 21.6262C40.358 21.4363 40.6863 21.5589 40.7581 21.5934C40.7876 21.655 40.8547 21.857 40.8082 22.3336C40.7408 23.0255 40.4502 24.0046 39.8572 25.2301C38.6799 27.6631 36.5085 30.6631 33.5858 33.5858C30.6631 36.5085 27.6632 38.6799 25.2301 39.8572C24.0046 40.4502 23.0255 40.7407 22.3336 40.8082C21.8571 40.8547 21.6551 40.7875 21.5934 40.7581C21.5589 40.6863 21.4363 40.358 21.6262 39.475C21.8562 38.4054 22.4689 36.9657 23.5038 35.2817C24.7575 33.2417 26.5497 30.9744 28.7621 28.762C30.9744 26.5497 33.2417 24.7574 35.2817 23.5037C36.9657 22.4689 38.4054 21.8562 39.475 21.6262ZM4.41189 29.2403L18.7597 43.5881C19.8813 44.7097 21.4027 44.9179 22.7217 44.7893C24.0585 44.659 25.5148 44.1631 26.9723 43.4579C29.9052 42.0387 33.2618 39.5667 36.4142 36.4142C39.5667 33.2618 42.0387 29.9052 43.4579 26.9723C44.1631 25.5148 44.659 24.0585 44.7893 22.7217C44.9179 21.4027 44.7097 19.8813 43.5881 18.7597L29.2403 4.41187C27.8527 3.02428 25.8765 3.02573 24.2861 3.36776C22.6081 3.72863 20.7334 4.58419 18.8396 5.74801C16.4978 7.18716 13.9881 9.18353 11.5858 11.5858C9.18354 13.988 7.18717 16.4978 5.74802 18.8396C4.58421 20.7334 3.72865 22.6081 3.36778 24.2861C3.02574 25.8765 3.02429 27.8527 4.41189 29.2403Z"
          fill="currentColor"
          fill-rule="evenodd"
        ></path>
      </svg>
      <a href="/" use:inertia>
        <h2 class="text-[var(--text-primary)] text-xl font-bold">
          QPayground
        </h2>
      </a>
    </div>

    <!-- Desktop Navigation -->
    <nav class="hidden md:flex items-center gap-8">
      {#each links as link}
        {#if !link.isButton}
          <a
            class="text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors text-sm font-medium {link.isActive ? 'active-link' : ''}"
            href={link.href}
            use:inertia
          >
            {link.name}
          </a>
        {/if}
      {/each}
    </nav>

    <!-- Desktop CTA Button -->
    <div class="hidden md:flex">
      {#each links as link}
        {#if link.isButton}
          <a
            href={link.href}
            use:inertia
            class="min-w-[100px] flex items-center justify-center rounded-lg h-10 px-5 bg-[var(--primary-color)] hover:bg-blue-600 transition-colors text-white text-sm font-bold shadow-lg glow-effect"
          >
            {link.name}
          </a>
        {/if}
      {/each}
    </div>

    <!-- Mobile Menu Button -->
    <button class="md:hidden text-[var(--text-primary)]" onclick={toggleMobileMenu}>
      <svg
        class="h-6 w-6"
        fill="none"
        height="24"
        stroke="currentColor"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        viewBox="0 0 24 24"
        width="24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <line x1="4" x2="20" y1="12" y2="12"></line>
        <line x1="4" x2="20" y1="6" y2="6"></line>
        <line x1="4" x2="20" y1="18" y2="18"></line>
      </svg>
    </button>
  </div>

  <!-- Mobile Menu -->
  {#if mobileMenuOpen}
    <div
      class="md:hidden absolute top-full left-0 w-full bg-[var(--background-dark)] border-t border-[var(--border-color)] shadow-lg"
    >
      <nav class="flex flex-col items-center py-4 space-y-4">
        {#each links as link}
          {#if !link.isButton}
            <a
              class="text-[var(--text-primary)] hover:text-[var(--primary-color)] transition-colors text-base font-medium {link.isActive ? 'active-link-mobile' : ''}"
              href={link.href}
              use:inertia
              onclick={closeMenu}
            >
              {link.name}
            </a>
          {:else}
            <a
              href={link.href}
              use:inertia
              class="min-w-[100px] items-center justify-center rounded-lg h-10 px-5 bg-[var(--primary-color)] hover:bg-blue-600 transition-colors text-white text-base font-bold shadow-lg glow-effect"
              onclick={closeMenu}
            >
              {link.name}
            </a>
          {/if}
        {/each}
      </nav>
    </div>
  {/if}
</header>

<style>
  .active-link {
    color: var(--primary-color);
    font-weight: 600;
    position: relative;
  }

  .active-link::after {
    content: '';
    position: absolute;
    bottom: -4px;
    left: 0;
    right: 0;
    height: 2px;
    background-color: var(--primary-color);
    border-radius: 1px;
  }

  .active-link-mobile {
    color: var(--primary-color);
    font-weight: 600;
  }
</style>