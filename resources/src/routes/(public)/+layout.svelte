<script lang="ts">
  import PublicFooter from "$lib/components/PublicFooter.svelte";
  import PublicHeader from "$lib/components/PublicHeader.svelte";
  import RootLayout from "$lib/components/RootLayout.svelte";

  let {
    children,
    flash,
  }: {
    children: any;
    flash: any;
  } = $props();
</script>

<RootLayout {flash}>
  <PublicHeader />
  {@render children()}
  <PublicFooter />
</RootLayout>
