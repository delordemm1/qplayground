<script lang="ts">
  import { inertia } from "@inertiajs/svelte";
  import { showErrorToast, showSuccessToast } from "$lib/utils/toast";

  let email = $state("");
  let otp = $state("");
  let isOtpSent = $state(false);
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  async function requestOTP(e) {
    e.preventDefault();
    if (!email) {
      errors.email = "Email is required";
      return;
    }

    isLoading = true;
    errors = {};

    try {
      const response = await fetch("/auth/request-otp", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email }),
      });

      const data = await response.json();

      if (response.ok) {
        isOtpSent = true;
        showSuccessToast("OTP sent to your email address");
      } else {
        if (data.errors) {
          errors = data.errors;
        } else {
          showErrorToast(data.error || "Failed to send OTP");
        }
      }
    } catch (error) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isLoading = false;
    }
  }

  async function verifyOTP(e) {
    e.preventDefault();
    if (!otp) {
      errors.otp = "OTP is required";
      return;
    }

    isLoading = true;
    errors = {};

    try {
      const response = await fetch("/auth/verify-otp", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, otp }),
      });

      const data = await response.json();

      if (response.ok) {
        showSuccessToast("Login successful!");
        // Redirect to dashboard
        window.location.href = "/dashboard";
      } else {
        if (data.errors) {
          errors = data.errors;
        } else {
          showErrorToast(data.error || "Invalid OTP");
        }
      }
    } catch (error) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isLoading = false;
    }
  }

  function resetForm() {
    isOtpSent = false;
    otp = "";
    errors = {};
  }
</script>

<svelte:head>
  <title>QPlayground - Authentication</title>
</svelte:head>


<main class="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
  <div class="sm:mx-auto sm:w-full sm:max-w-md">
    <div class="flex justify-center">
      <svg
        class="h-12 w-12 text-primary-600"
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
    </div>
    <h2 class="mt-6 text-center text-3xl font-bold tracking-tight text-gray-900">
      Sign in to QPlayground
    </h2>
    <p class="mt-2 text-center text-sm text-gray-600">
      Enter your email to receive a one-time password
    </p>
  </div>

  <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
    <div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
      {#if !isOtpSent}
        <!-- Email Input Form -->
        <form onsubmit={requestOTP} class="space-y-6">
          <div>
            <label for="email" class="block text-sm font-medium text-gray-700">
              Email address
            </label>
            <div class="mt-1">
              <input
                id="email"
                name="email"
                type="email"
                autocomplete="email"
                required
                bind:value={email}
                class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-primary-500 sm:text-sm"
                class:border-red-300={errors.email}
                class:focus:border-red-500={errors.email}
                class:focus:ring-red-500={errors.email}
                placeholder="Enter your email"
              />
              {#if errors.email}
                <p class="mt-2 text-sm text-red-600">{errors.email}</p>
              {/if}
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={isLoading}
              class="flex w-full justify-center rounded-md border border-transparent bg-primary-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {#if isLoading}
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Sending...
              {:else}
                Send OTP
              {/if}
            </button>
          </div>
        </form>
      {:else}
        <!-- OTP Verification Form -->
        <form onsubmit={verifyOTP} class="space-y-6">
          <div>
            <label for="otp" class="block text-sm font-medium text-gray-700">
              One-Time Password
            </label>
            <div class="mt-1">
              <input
                id="otp"
                name="otp"
                type="text"
                autocomplete="one-time-code"
                required
                maxlength="6"
                bind:value={otp}
                class="block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-primary-500 sm:text-sm text-center text-lg tracking-widest"
                class:border-red-300={errors.otp}
                class:focus:border-red-500={errors.otp}
                class:focus:ring-red-500={errors.otp}
                placeholder="000000"
              />
              {#if errors.otp}
                <p class="mt-2 text-sm text-red-600">{errors.otp}</p>
              {/if}
            </div>
            <p class="mt-2 text-sm text-gray-500">
              Enter the 6-digit code sent to <strong>{email}</strong>
            </p>
          </div>

          <div class="space-y-3">
            <button
              type="submit"
              disabled={isLoading}
              class="flex w-full justify-center rounded-md border border-transparent bg-primary-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {#if isLoading}
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Verifying...
              {:else}
                Verify & Sign In
              {/if}
            </button>

            <button
              type="button"
              onclick={resetForm}
              class="flex w-full justify-center rounded-md border border-gray-300 bg-white py-2 px-4 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
            >
              Use different email
            </button>
          </div>
        </form>
      {/if}

      <div class="mt-6">
        <div class="relative">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-300" />
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="bg-white px-2 text-gray-500">Secure authentication</span>
          </div>
        </div>
      </div>
    </div>
  </div>
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