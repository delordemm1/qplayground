<script lang="ts">
    import type { FlashMessageType } from '$lib/types/index'; // Assuming you define your types
  
    type Props = {
      type: FlashMessageType;
      message: string;
    };
  
    let { type, message }: Props = $props();
  
    const typeStyles = {
      success: {
        icon: 'fa-solid fa-circle-check',
        barColor: 'var(--color-success-500, #22c55e)', // Fallback green
      },
      error: {
        icon: 'fa-solid fa-circle-xmark',
        barColor: 'var(--color-error-500, #ef4444)', // Fallback red
      },
      warning: {
        icon: 'fa-solid fa-triangle-exclamation',
        barColor: 'var(--color-warning-500, #f59e0b)', // Fallback amber
      },
      info: {
        icon: 'fa-solid fa-circle-info',
        barColor: 'var(--color-info-500, #3b82f6)', // Fallback blue
      },
    };
  
    const style = typeStyles[type] ?? typeStyles.info;
  </script>
  
  <div
    class="toast-container"
    style:--bar-color={style.barColor}
  >
    <div class="icon-wrapper">
      <i class={style.icon}></i>
    </div>
  
    <div class="message-wrapper">
      <p>{message}</p>
    </div>
  </div>
  
  <style>
    .toast-container {
      display: flex;
      align-items: center;
      background-color: white;
      color: #1f2937; /* gray-800 */
      border-radius: 8px;
      padding: 12px 16px;
      min-width: 300px;
      box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
      border-left: 5px solid var(--bar-color);
    }
  
    .icon-wrapper {
      font-size: 1.25rem; /* 20px */
      color: var(--bar-color);
      flex-shrink: 0;
    }
  
    .message-wrapper {
      margin-left: 12px;
      font-size: 0.875rem; /* 14px */
      line-height: 1.25rem;
      font-weight: 500;
    }
  </style>