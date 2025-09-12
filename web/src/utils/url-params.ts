'use client';

/**
 * Parse URL parameters for token and address
 * Used for supporting external auth flow where token and address are passed via URL
 */
export interface UrlAuthParams {
  token: string | null;
  address: string | null;
}

/**
 * Parse URL search parameters to extract auth-related parameters
 */
export function parseUrlAuthParams(): UrlAuthParams {
  if (typeof window === 'undefined') {
    return { token: null, address: null };
  }

  const searchParams = new URLSearchParams(window.location.search);
  const rawToken = searchParams.get('token');
  const rawAddress = searchParams.get('address');

  // Decode URL parameters
  const token = rawToken ? decodeURIComponent(rawToken) : null;
  const address = rawAddress ? decodeURIComponent(rawAddress) : null;

  // Validate address format if provided
  const validAddress = address && address.startsWith('0x') && address.length === 42 
    ? address as `0x${string}`
    : null;

  return {
    token,
    address: validAddress
  };
}

/**
 * Clear URL auth parameters from the current URL without triggering page reload
 */
export function clearUrlAuthParams(): void {
  if (typeof window === 'undefined') return;

  const url = new URL(window.location.href);
  url.searchParams.delete('token');
  url.searchParams.delete('address');
  
  // Update URL without reload
  window.history.replaceState({}, '', url.toString());
}

/**
 * Check if URL contains auth parameters
 */
export function hasUrlAuthParams(): boolean {
  const params = parseUrlAuthParams();
  return !!(params.token && params.address);
}