'use client';
import { RainbowKitProvider } from '@rainbow-me/rainbowkit';
import { QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';
import { WagmiProvider } from 'wagmi';

import { APP_NAME } from '@/config/base';
import { config, queryClient } from '@/config/wagmi';
import { useRainbowKitTheme } from '@/hooks/useRainbowKitTheme';
import { getDefaultChain, getDefaultChainId } from '@/utils/chains';
import '@rainbow-me/rainbowkit/styles.css';

export function DAppProvider({ children }: React.PropsWithChildren<unknown>) {
  const rainbowKitTheme = useRainbowKitTheme();
  
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider
          theme={rainbowKitTheme}
          locale="en-US"
          appInfo={{ appName: APP_NAME }}
          initialChain={getDefaultChain()}
          id={getDefaultChainId() ? String(getDefaultChainId()) : undefined}
        >
          {children}
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
}
