import { ExclamationCircleOutline, InfoCircleOutline } from 'flowbite-svelte-icons';
import toast from 'svelte-french-toast';

export const toastStyle = 'background-color: var(--color-primary-50); color: black;';

export const showSuccessToast = (message: string) => {
  toast.success(message, {
    style: toastStyle
  });
};

export const showErrorToast = (message: string) => {
  toast.error(message, {
    style: toastStyle
  });
};

export const showInfoToast = (message: string) => {
  toast(message, {
    style: toastStyle,
    // @ts-expect-error
    icon: InfoCircleOutline,
    className: 'info-toast',
    iconTheme: {
      primary: '#007AFF',
      secondary: '#007AFF'
    }
  });
};

export const showWarningToast = (message: string) => {
  toast(message, {
    // @ts-expect-error
    icon: ExclamationCircleOutline,
    style: toastStyle,
    className: 'warning-toast',
    iconTheme: {
      primary: '#FFB400',
      secondary: '#FFB400'
    }
  });
};