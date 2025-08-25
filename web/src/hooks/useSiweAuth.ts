'use client';

import { useCallback, useState } from 'react';
import { toast } from 'react-toastify';
import { createSiweMessage } from 'viem/siwe';
import { useAccount, useSignMessage, useChainId } from 'wagmi';

import { useAuth } from '@/contexts/auth';
import { createPublicClient } from '@/lib/graphql/client';
import { useLogin } from '@/lib/graphql/hooks';
import { QUERY_NONCE } from '@/lib/graphql/queries';
import type { NonceVariables } from '@/lib/graphql/types';


export const useSiweAuth = () => {
  const { address, isConnected } = useAccount();
  const chainId = useChainId();
  const { setToken } = useAuth();
  const [isAuthenticating, setIsAuthenticating] = useState(false);

  const { signMessageAsync } = useSignMessage();
  const loginMutation = useLogin();

  const getNonce = useCallback(async (): Promise<string> => {
    const client = createPublicClient();
    const variables: NonceVariables = { input: { length: 32 } };
    const data = await client.request<{ nonce: string }>(QUERY_NONCE, variables);
    return data.nonce;
  }, []);

  const createMessage = useCallback((address: `0x${string}`, nonce: string) => {
    return createSiweMessage({
      domain: typeof window !== 'undefined' ? window.location.host : 'apps.degov.ai',
      address,
      statement: `DeGov.AI wants you to sign in with your Ethereum account: ${address}`,
      uri: typeof window !== 'undefined' ? window.location.origin : 'https://apps.degov.ai',
      version: '1',
      chainId,
      nonce,
    });
  }, [chainId]);

  const authenticate = useCallback(async (): Promise<boolean> => {
    if (!isConnected || !address) {
      toast.error('Please connect your wallet first');
      return false;
    }

    setIsAuthenticating(true);

    try {
      const nonce = await getNonce();
      if (!nonce) {
        throw new Error('Failed to get nonce');
      }

      const message = createMessage(address, nonce);

      const signature = await signMessageAsync({ message });

      const token = await loginMutation.mutateAsync({
        input: {
          message,
          signature
        }
      });

      setToken(token);
      return true;

    } catch (error) {
      console.error('Authentication failed:', error);
      if (error instanceof Error) {
        if (error.message.includes('User rejected')) {
          toast.error('Authentication cancelled');
        } else {
          toast.error(`Authentication failed: ${error.message}`);
        }
      } else {
        toast.error('Authentication failed. Please try again.');
      }
      return false;
    } finally {
      setIsAuthenticating(false);
    }
  }, [isConnected, address, getNonce, createMessage, signMessageAsync, loginMutation, setToken]);

  return {
    authenticate,
    isAuthenticating,
    canAuthenticate: isConnected && !!address
  };
};