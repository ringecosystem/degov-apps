'use client';

import { useState, useEffect, useCallback } from 'react';

import { tokenManager } from '@/lib/auth/token-manager';
import { parseUrlAuthParams, clearUrlAuthParams, hasUrlAuthParams } from '@/utils/url-params';

import { AuthContext } from './context';

import type { ReactNode } from 'react';

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [token, setTokenState] = useState<string | null>(null);
  const [urlAddress, setUrlAddress] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  const isUsingUrlAuth = !!(token && urlAddress);
  const isAuthenticated = !!token;

  // Initialize from storage and URL parameters on mount
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const urlParams = parseUrlAuthParams();
      const storedAuthData = tokenManager.getAuthData();

      // Clear URL params first
      if (urlParams.token || urlParams.address) {
        clearUrlAuthParams();
      }

      // Simple priority: stored > URL params
      if (storedAuthData.token) {
        setTokenState(storedAuthData.token);
        setUrlAddress(storedAuthData.address);
      } else if (urlParams.token && urlParams.address) {
        // Save URL params and set state
        tokenManager.setToken(urlParams.token);
        tokenManager.setAddress(urlParams.address);
        setTokenState(urlParams.token);
        setUrlAddress(urlParams.address);
      }

      setIsInitialized(true);
    }
  }, []);

  const setToken = (newToken: string | null) => {
    setTokenState(newToken);
    if (newToken) {
      setUrlAddress(null);
      tokenManager.setToken(newToken);
      tokenManager.setAddress(null);
    } else {
      setUrlAddress(null);
      tokenManager.setToken(null);
      tokenManager.setAddress(null);
    }
  };

  const clearAuth = () => {
    setTokenState(null);
    setUrlAddress(null);
    tokenManager.setToken(null);
    tokenManager.setAddress(null);
  };

  // Don't render children until token is initialized from localStorage
  if (!isInitialized) {
    return null;
  }

  return (
    <AuthContext.Provider
      value={{
        token,
        setToken,
        isAuthenticated,
        urlAddress,
        isUsingUrlAuth,
        clearUrlAuth: clearAuth,
        clearUrlAuthOnError: clearAuth
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
