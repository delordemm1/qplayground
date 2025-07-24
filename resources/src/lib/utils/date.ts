export const formatDate = (dateString: string) => {
  if (!dateString) return "N/A";
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
};

export function toRFC3339(date: Date | string) {
  if (typeof date === "string") {
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
  if (!dateValue) return "";

  const date = new Date(dateValue);

  // Check if the created date is valid
  if (isNaN(date.getTime())) {
    return "";
  }

  // Get local date and time components
  const year = date.getFullYear();
  // Pad month, day, hours, and minutes with a leading zero if they are single-digit
  const month = String(date.getMonth() + 1).padStart(2, "0"); // getMonth() is 0-indexed
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");

  // Return the correctly formatted string
  return `${year}-${month}-${day}T${hours}:${minutes}`;
}

// Helper function to format duration
export function formatDuration(durationMs: number): string {
  if (durationMs < 1000) {
    return `${durationMs}ms`;
  } else if (durationMs < 60000) {
    return `${(durationMs / 1000).toFixed(2)}s`;
  } else {
    const minutes = Math.floor(durationMs / 60000);
    const seconds = ((durationMs % 60000) / 1000).toFixed(2);
    return `${minutes}m ${seconds}s`;
  }
}

// Calculate percentiles from an array of numbers
export function calculatePercentile(values: number[], percentile: number): number {
  if (values.length === 0) return 0;
  
  const sorted = [...values].sort((a, b) => a - b);
  const index = (percentile / 100) * (sorted.length - 1);
  
  if (Number.isInteger(index)) {
    return sorted[index];
  } else {
    const lower = Math.floor(index);
    const upper = Math.ceil(index);
    const weight = index - lower;
    return sorted[lower] * (1 - weight) + sorted[upper] * weight;
  }
}

// Calculate statistical metrics from an array of durations
export function calculateDurationStats(durations: number[]) {
  if (durations.length === 0) {
    return {
      avg: 0,
      min: 0,
      max: 0,
      p50: 0,
      p95: 0,
      count: 0
    };
  }

  const sorted = [...durations].sort((a, b) => a - b);
  const sum = durations.reduce((acc, val) => acc + val, 0);

  return {
    avg: sum / durations.length,
    min: sorted[0],
    max: sorted[sorted.length - 1],
    p50: calculatePercentile(durations, 50),
    p95: calculatePercentile(durations, 95),
    count: durations.length
  };
}
