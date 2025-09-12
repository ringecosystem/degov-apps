'use client';

import { useAccount as useWagmiAccount } from 'wagmi';

import { useAuth } from '@/contexts/auth';

/**
 * Enhanced useAccount hook that supports both wagmi wallet connection and URL-based auth
 * 
 * Priority:
 * 1. If URL auth is active, ALWAYS use URL auth (disable wagmi completely)
 * 2. If wallet is connected and no URL auth, use wallet's address and connection status  
 * 3. Otherwise, return disconnected state
 */
export const useAccount = () => {
  const wagmiAccount = useWagmiAccount();
  const { isUsingUrlAuth, urlAddress } = useAuth();

  // URL auth takes COMPLETE priority - when active, wagmi is completely disabled
  if (isUsingUrlAuth && urlAddress) {
    return {
      address: urlAddress as `0x${string}`,
      isConnected: true,
      isConnecting: false,
      isReconnecting: false,
      chainId: undefined, // URL auth doesn't provide chainId
      // Additional fields to indicate auth source
      authSource: 'url' as const,
      isUsingUrlAuth: true,
    };
  }

  // Only use wallet when URL auth is NOT active
  if (!isUsingUrlAuth && wagmiAccount.isConnected && wagmiAccount.address) {
    return {
      address: wagmiAccount.address,
      isConnected: true,
      isConnecting: wagmiAccount.isConnecting,
      isReconnecting: wagmiAccount.isReconnecting,
      chainId: wagmiAccount.chainId,
      // Additional fields to indicate auth source
      authSource: 'wallet' as const,
      isUsingUrlAuth: false,
    };
  }

  // Default disconnected state
  return {
    address: undefined,
    isConnected: false,
    isConnecting: isUsingUrlAuth ? false : wagmiAccount.isConnecting, // Disable wagmi loading states during URL auth
    isReconnecting: isUsingUrlAuth ? false : wagmiAccount.isReconnecting,
    chainId: wagmiAccount.chainId,
    // Additional fields to indicate auth source
    authSource: 'none' as const,
    isUsingUrlAuth: false,
  };
};