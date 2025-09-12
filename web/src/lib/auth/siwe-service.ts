'use client';

import { createSiweMessage } from 'viem/siwe';


import { createPublicClient } from '@/lib/graphql/client';
import { QUERY_NONCE, LOGIN_MUTATION } from '@/lib/graphql/queries';
import type { NonceVariables, LoginVariables } from '@/lib/graphql/types';

import { tokenManager } from './token-manager';

import type { AuthResult } from './global-auth-manager';
import type { SignMessageParameters } from 'wagmi/actions';

interface AuthenticateWithWalletParams {
  address: `0x${string}`;
  chainId: number;
  signMessageAsync: (args: SignMessageParameters) => Promise<`0x${string}`>;
}

class SiweService {
  private async getNonce(): Promise<string> {
    const client = createPublicClient();
    const variables: NonceVariables = { input: { length: 32 } };
    const data = await client.request<{ nonce: string }>(QUERY_NONCE, variables);
    return data.nonce;
  }

  private createMessage(address: `0x${string}`, nonce: string, chainId: number): string {
    return createSiweMessage({
      domain: typeof window !== 'undefined' ? window.location.host : 'apps.degov.ai',
      address,
      statement: `DeGov.AI wants you to sign in with your Ethereum account: ${address}`,
      uri: typeof window !== 'undefined' ? window.location.origin : 'https://apps.degov.ai',
      version: '1',
      chainId,
      nonce
    });
  }

  async authenticateWithWallet({
    address,
    chainId,
    signMessageAsync
  }: AuthenticateWithWalletParams): Promise<AuthResult> {
    try {
      // Get nonce
      const nonce = await this.getNonce();
      if (!nonce) {
        return { success: false, error: 'Failed to get nonce' };
      }

      // Create SIWE message
      const message = this.createMessage(address, nonce, chainId);

      // Sign message
      const signature = await signMessageAsync({ message });

      // Login with GraphQL
      const client = createPublicClient();
      const variables: LoginVariables = {
        input: {
          message,
          signature
        }
      };

      const data = await client.request<{ login: { token: string } }>(LOGIN_MUTATION, variables);
      const token = data.login.token;

      if (!token) {
        return { success: false, error: 'No token received from login' };
      }

      // Store token
      tokenManager.setToken(token);

      return {
        success: true,
        token
      };
    } catch (error) {
      console.error('Authentication failed:', error);

      let errorMessage = 'Authentication failed';
      if (error instanceof Error) {
        if (error.message.includes('User rejected')) {
          errorMessage = 'Authentication cancelled by user';
        } else {
          errorMessage = error.message;
        }
      }

      return {
        success: false,
        error: errorMessage
      };
    }
  }

  async signOut(): Promise<void> {
    try {
      // Clear both token and address
      tokenManager.setToken(null);
      tokenManager.setAddress(null);
    } catch (error) {
      console.error('Sign out failed:', error);
      throw error;
    }
  }
}

export const siweService = new SiweService();
