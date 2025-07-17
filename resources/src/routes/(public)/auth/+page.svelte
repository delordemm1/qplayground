<script lang="ts">
  import { inertia } from "@inertiajs/svelte";
</script>

<svelte:head>
  <title>QPlayground - Authentication</title>
</svelte:head>


<main>
</main>

<style>
  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .animate-fadeIn {
    animation: fadeIn 1s ease-in-out forwards;
  }
  .stagger-animation {
    animation: fadeIn 0.8s ease-in-out both;
  }
</style>