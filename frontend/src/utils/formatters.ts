/**
 * Return a TailwindCSS text color class based on the sentiment score.
 */
export function getSentimentColor(sentimentScore?: number | null): string {
  if (sentimentScore === undefined || sentimentScore === null) {
    return 'text-gray-500';
  }
  if (sentimentScore > 0) {
    return 'text-green-600';
  }
  if (sentimentScore < 0) {
    return 'text-red-600';
  }
  return 'text-gray-500';
}

/**
 * Format a date/time string into a relative "X ago" label.
 */
export function formatTime(timeString: string): string {
  if (!timeString) return '';
  const date = new Date(timeString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

/**
 * Return a human-readable sentiment label.
 */
export function getSentimentLabel(sentimentScore?: number | null): string {
  if (sentimentScore === undefined || sentimentScore === null) return '';
  if (sentimentScore > 0) return 'Positive';
  if (sentimentScore < 0) return 'Negative';
  return 'Neutral';
}

/**
 * Return a sentiment arrow character.
 */
export function getSentimentArrow(sentimentScore?: number | null): string {
  if (sentimentScore === undefined || sentimentScore === null) return '';
  if (sentimentScore > 0) return '↑';
  if (sentimentScore < 0) return '↓';
  return '→';
}
