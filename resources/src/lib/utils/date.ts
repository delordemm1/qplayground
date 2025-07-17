export const formatDate = (dateString: string) => {
  if (!dateString) return "N/A";
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
};

export function toRFC3339(date: Date|string) {
    if (typeof date === 'string') {
        date = new Date(date);
    }
  const pad = (n: number) => (n < 10 ? "0" + n : n);

  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  const seconds = pad(date.getSeconds());

  const timezoneOffset = -date.getTimezoneOffset();
  const sign = timezoneOffset >= 0 ? "+" : "-";
  const offsetHours = pad(Math.floor(Math.abs(timezoneOffset) / 60));
  const offsetMinutes = pad(Math.abs(timezoneOffset) % 60);
  const timezoneString = `${sign}${offsetHours}:${offsetMinutes}`;

  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}${timezoneString}`;
}

/**
   * Formats a date string or Date object for a datetime-local input.
   * @param {string | Date} dateValue - The date to format.
   * @returns {string} The formatted date string or an empty string.
   */
export function formatDateForInput(dateValue: string | Date): string {
  // Return an empty string if the input is null, undefined, or empty
  if (!dateValue) return '';

  const date = new Date(dateValue);

  // Check if the created date is valid
  if (isNaN(date.getTime())) {
    return '';
  }

  // Get local date and time components
  const year = date.getFullYear();
  // Pad month, day, hours, and minutes with a leading zero if they are single-digit
  const month = String(date.getMonth() + 1).padStart(2, '0'); // getMonth() is 0-indexed
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');

  // Return the correctly formatted string
  return `${year}-${month}-${day}T${hours}:${minutes}`;
}