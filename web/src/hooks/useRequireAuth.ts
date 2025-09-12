'use client';

import { useConnectModal } from '@rainbow-me/rainbowkit';
import { useCallback } from 'react';

import { useAuth } from '@/contexts/auth';
import { useAccount } from '@/hooks/useAccount';
import { useSiweAuth } from '@/hooks/useSiweAuth';


export const useRequireAuth = () => {
  const { isAuthenticated } = useAuth();
  const { isConnected, authSource } = useAccount();
  const { openConnectModal } = useConnectModal();
  const { authenticate, isAuthenticating } = useSiweAuth();

  /**
   * Ensure user is connected and authenticated, if not, guide user to complete authentication
   * @returns Promise<boolean> - Whether authentication is successful
   */
  const ensureAuth = useCallback(async (): Promise<boolean> => {
    // For URL auth, skip wallet connection modal and go directly to auth check
    if (!isConnected) {
      // Only try to open wallet modal if not using URL auth
      if (authSource !== 'url') {
        openConnectModal?.();
        return false;
      }
    }

    // Check authentication status
    if (!isAuthenticated) {
      const authResult = await authenticate();
      return authResult.success;
    }

    return true;
  }, [isConnected, authSource, isAuthenticated, openConnectModal, authenticate]);

  /**
    * Wrap operations that require authentication
   * @param action - The operation function that requires authentication
   * @returns The wrapped function, which will first check the authentication status
   */
  const withAuth = useCallback(
    <T extends any[], R>(action: (...args: T) => Promise<R> | R) => {
      return async (...args: T): Promise<R | null> => {
        const isAuthed = await ensureAuth();
        if (!isAuthed) {
          return null;
        }
        return action(...args);
      };
    },
    [ensureAuth]
  );

  return {
    ensureAuth,
    withAuth,
    isAuthenticating,
    canAuthenticate: isConnected,
    isAuthenticated: isConnected && isAuthenticated,
  };
};