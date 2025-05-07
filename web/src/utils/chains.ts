import { darwinia, mainnet } from 'viem/chains';

import type { Chain } from '@rainbow-me/rainbowkit';

const chainConfigMap: Record<number, Chain> = {
  [mainnet.id]: mainnet,
  [darwinia.id]: darwinia
};

// Returns an array of all chain configurations, filtering based on deployment mode
export function getChains(): [Chain, ...Chain[]] {
  const filteredChains: Chain[] = Object.values(chainConfigMap);
  if (filteredChains.length === 0) {
    throw new Error('No suitable chain configurations are available.');
  }
  return filteredChains as [Chain, ...Chain[]];
}

// Returns the chain by its id
export function getChainById(id?: number): Chain | undefined {
  return id ? chainConfigMap[id] : undefined;
}

// Returns the default chain configuration based on deployment mode
export function getDefaultChain(): Chain {
  const filteredChains = Object.values(chainConfigMap);
  if (filteredChains.length === 0) {
    throw new Error(
      'No suitable chain configurations are available for the current deployment mode.'
    );
  }

  const defaultChainId = mainnet.id;
  const defaultChain = filteredChains.find((chain) => chain.id === defaultChainId);

  return defaultChain || filteredChains[0];
}

// Returns the default chain id based on the default chain
export function getDefaultChainId(): number {
  const defaultChain = getDefaultChain();
  return defaultChain.id;
}

// return if the chain is supported
export function isSupportedChain(chainId: number): boolean {
  const filteredChains = Object.values(chainConfigMap);
  return filteredChains.some((chain) => chain.id === chainId);
}
