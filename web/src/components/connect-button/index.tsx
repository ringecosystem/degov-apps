'use client';
import { useConnectModal } from '@rainbow-me/rainbowkit';

import { useAuth } from '@/contexts/auth';
import { useAccount } from '@/hooks/useAccount';
import { isSupportedChainById } from '@/utils/chains';

import { Button } from '../ui/button';

import { Connected } from './connected';

export const ConnectButton = () => {
  const { openConnectModal } = useConnectModal();
  const { chainId, address, isConnected, isConnecting, isReconnecting, authSource } = useAccount();
  const { clearUrlAuth } = useAuth();

  if (isConnecting || isReconnecting) {
    return null;
  }

  if (!isConnected && openConnectModal) {
    return (
      <Button onClick={openConnectModal} className="rounded-[100px]">
        Connect Wallet
      </Button>
    );
  }

  // For URL auth, chainId might be undefined, so only check if it exists
  if (chainId && !isSupportedChainById(Number(chainId))) {
    return (
      <Button variant="destructive" className="cursor-auto rounded-[100px]">
        Error Chain
      </Button>
    );
  }

  if (address) {
    return <Connected address={address} authSource={authSource} />;
  }

  return null;
};
