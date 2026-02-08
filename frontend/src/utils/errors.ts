import axios from 'axios';

/**
 * Extract a human-readable error message from an unknown error.
 * Works with Axios errors, standard Errors, and unknown types.
 */
export function extractErrorMessage(err: unknown, fallback = 'An unexpected error occurred'): string {
  if (axios.isAxiosError(err)) {
    return err.response?.data?.message
      || err.response?.data?.detail
      || err.message
      || fallback;
  }
  if (err instanceof Error) {
    return err.message;
  }
  return fallback;
}
