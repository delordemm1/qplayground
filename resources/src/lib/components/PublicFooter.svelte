<script lang="ts">
  import { inertia } from "@inertiajs/svelte";
</script>

<footer class="bg-gray-800 text-gray-300">
  <div class="container mx-auto px-6 py-12">
    <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-8">
      <div class="col-span-2 md:col-span-4 lg:col-span-1">
        <h3 class="text-xl font-bold text-white mb-4">QPayground</h3>
        <p class="text-sm">Your Stage for Creativity & Competition.</p>
      </div>
      <div>
        <h4 class="font-semibold text-white mb-4">Company</h4>
        <ul class="space-y-2 text-sm">
          <li><a class="hover:text-white" href="/about" use:inertia>About Us</a></li>
          <li><a class="hover:text-white" href="/blog" use:inertia>Blog</a></li>
          <li><a class="hover:text-white" href="/contact" use:inertia>Contact Us</a></li>
        </ul>
      </div>
      <div>
        <h4 class="font-semibold text-white mb-4">For Qreators</h4>
        <ul class="space-y-2 text-sm">
          <li><a class="hover:text-white" href="#how-to-host">How to Host</a></li>
          <li><a class="hover:text-white" href="#features">Features</a></li>
          <li><a class="hover:text-white" href="#pricing">Pricing</a></li>
        </ul>
      </div>
      <div>
        <h4 class="font-semibold text-white mb-4">For Qrew</h4>
        <ul class="space-y-2 text-sm">
          <li><a class="hover:text-white" href="#explore">Explore Contests</a></li>
          <li><a class="hover:text-white" href="#how-to-join">How to Join</a></li>
          <li><a class="hover:text-white" href="#faqs">FAQs</a></li>
        </ul>
      </div>
      <div>
        <h4 class="font-semibold text-white mb-4">Legal</h4>
        <ul class="space-y-2 text-sm">
          <li><a class="hover:text-white" href="/terms" use:inertia>Terms of Service</a></li>
          <li><a class="hover:text-white" href="#privacy">Privacy Policy</a></li>
        </ul>
      </div>
    </div>
    <div
      class="mt-12 border-t border-gray-700 pt-8 flex flex-col md:flex-row justify-between items-center text-sm"
    >
      <p>© 2024 QPayground. All Rights Reserved.</p>
      <div class="flex space-x-4 mt-4 md:mt-0">
        <a class="hover:text-white" href="#facebook">
          <svg
            aria-hidden="true"
            class="w-6 h-6"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              clip-rule="evenodd"
              d="M22 12c0-5.523-4.477-10-10-10S2 6.477 2 12c0 4.991 3.657 9.128 8.438 9.878v-6.987h-2.54V12h2.54V9.797c0-2.506 1.492-3.89 3.777-3.89 1.094 0 2.238.195 2.238.195v2.46h-1.26c-1.243 0-1.63.771-1.63 1.562V12h2.773l-.443 2.89h-2.33v6.988C18.343 21.128 22 16.991 22 12z"
              fill-rule="evenodd"
            ></path>
          </svg>
        </a>
        <a class="hover:text-white" href="#twitter">
          <svg
            aria-hidden="true"
            class="w-6 h-6"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              d="M8.29 20.251c7.547 0 11.675-6.253 11.675-11.675 0-.178 0-.355-.012-.53A8.348 8.348 0 0022 5.92a8.19 8.19 0 01-2.357.646 4.118 4.118 0 001.804-2.27 8.224 8.224 0 01-2.605.996 4.107 4.107 0 00-6.993 3.743 11.65 11.65 0 01-8.457-4.287 4.106 4.106 0 001.27 5.477A4.072 4.072 0 012.8 9.71v.052a4.105 4.105 0 003.292 4.022 4.095 4.095 0 01-1.853.07 4.108 4.108 0 003.834 2.85A8.233 8.233 0 012 18.407a11.616 11.616 0 006.29 1.84"
            ></path>
          </svg>
        </a>
        <a class="hover:text-white" href="#instagram">
          <svg
            aria-hidden="true"
            class="w-6 h-6"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              clip-rule="evenodd"
              d="M12.315 2c2.43 0 2.784.013 3.808.06 1.064.049 1.791.218 2.427.465a4.902 4.902 0 011.772 1.153 4.902 4.902 0 011.153 1.772c.247.636.416 1.363.465 2.427.048 1.024.06 1.378.06 3.808s-.012 2.784-.06 3.808c-.049 1.064-.218 1.791-.465 2.427a4.902 4.902 0 01-1.153 1.772 4.902 4.902 0 01-1.772 1.153c-.636.247-1.363.416-2.427.465-1.024.048-1.378.06-3.808.06s-2.784-.013-3.808-.06c-1.064-.049-1.791-.218-2.427-.465a4.902 4.902 0 01-1.772-1.153 4.902 4.902 0 01-1.153-1.772c-.247-.636-.416-1.363-.465-2.427-.048-1.024-.06-1.378-.06-3.808s.012-2.784.06-3.808c.049-1.064.218-1.791.465-2.427a4.902 4.902 0 011.153-1.772A4.902 4.902 0 016.08 2.525c.636-.247 1.363-.416 2.427-.465C9.53 2.013 9.884 2 12.315 2zM12 8.118c-2.136 0-3.863 1.727-3.863 3.863s1.727 3.863 3.863 3.863 3.863-1.727 3.863-3.863S14.136 8.118 12 8.118zM12 14.172c-1.18 0-2.143-.963-2.143-2.143s.963-2.143 2.143-2.143 2.143.963 2.143 2.143S13.18 14.172 12 14.172zM16.838 7.022a1.212 1.212 0 110-2.424 1.212 1.212 0 010 2.424z"
              fill-rule="evenodd"
            ></path>
          </svg>
        </a>
      </div>
    </div>
  </div>
</footer>