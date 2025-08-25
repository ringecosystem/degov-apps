'use client';

import { useConnectModal } from '@rainbow-me/rainbowkit';
import { useCallback } from 'react';
import { useAccount } from 'wagmi';

import { useAuth } from '@/contexts/auth';
import { useSiweAuth } from '@/hooks/useSiweAuth';


export const useRequireAuth = () => {
  const { isAuthenticated } = useAuth();
  const { isConnected } = useAccount();
  const { openConnectModal } = useConnectModal();
  const { authenticate, isAuthenticating } = useSiweAuth();

  /**
   * Ensure user is connected and authenticated, if not, guide user to complete authentication
   * @returns Promise<boolean> - Whether authentication is successful
   */
  const ensureAuth = useCallback(async (): Promise<boolean> => {
    // 1. Check wallet connection status
    if (!isConnected) {
      openConnectModal?.();
      return false;
    }

    // 2. Check authentication status
    if (!isAuthenticated) {
      const authSuccess = await authenticate();
      return authSuccess;
    }

    return true;
  }, [isConnected, isAuthenticated, openConnectModal, authenticate]);

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