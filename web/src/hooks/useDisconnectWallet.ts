import { useCallback } from 'react';
import { useDisconnect } from 'wagmi';

import { useAuth } from '@/contexts/auth';
import { useAccount } from '@/hooks/useAccount';

export const useDisconnectWallet = () => {
  const { disconnect } = useDisconnect();
  const { clearUrlAuth, isUsingUrlAuth } = useAuth();
  const { authSource } = useAccount();

  const disconnectWallet = useCallback(
    async (address: `0x${string}`) => {
      // Handle URL auth disconnection - clear everything
      if (authSource === 'url' || isUsingUrlAuth) {
        clearUrlAuth();
        return;
      }

      // Handle wallet disconnection - also clear any existing auth states
      try {
        // Clear any tokens first
        clearUrlAuth();

        // Try to revoke wallet permissions
        if (typeof window !== 'undefined' && window?.ethereum?.request) {
          try {
            await window.ethereum.request({
              method: 'wallet_revokePermissions',
              params: [{ eth_accounts: address }]
            });
            console.log('Wallet permissions revoked successfully');
          } catch (error) {
            console.error('Error revoking wallet permissions:', error);
          }
        }

        // Disconnect wallet
        disconnect();
      } catch (error) {
        console.error('Error during wallet disconnection:', error);
      }
    },
    [disconnect, clearUrlAuth, authSource, isUsingUrlAuth]
  );
  return { disconnectWallet };
};
