import yaml from 'js-yaml';

import type { Chip, LastProposal } from '@/lib/graphql/types';

export interface DaoConfig {
  name: string;
  code: string;
  xprofile?: string;
  links: {
    website: string;
    config: string;
    indexer: string;
  };
  chain: {
    id: string;
    name: string;
  };
}

export interface ConfigData {
  daos: DaoConfig[];
}

export interface DaoInfo extends Record<string, unknown> {
  id: string;
  name: string;
  code: string;
  daoIcon: string;
  network: string;
  networkIcon: string;
  proposals: number;
  favorite: boolean;
  settable: boolean;
  website: string;
  indexer: string;
  chainId?: string;
  chips?: Chip[];
  lastProposal?: LastProposal | null;
}

export async function loadConfig(): Promise<ConfigData> {
  try {
    const response = await fetch('/config.yml');
    const yamlText = await response.text();
    const config = yaml.load(yamlText) as ConfigData;
    return config;
  } catch (error) {
    console.error('Error loading config:', error);
    throw error;
  }
}

export async function getProposalsCount(indexerUrl: string): Promise<number> {
  try {
    const response = await fetch(indexerUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        query: `
          query MyQuery {
            dataMetrics {
              proposalsCount
            }
          }
        `
      })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    if (data.errors) {
      console.error('GraphQL errors:', data.errors);
      return 0;
    }

    const proposalsCount = data.data?.dataMetrics?.[0]?.proposalsCount || 0;
    return proposalsCount;
  } catch (error) {
    console.error(`Error fetching proposals count from ${indexerUrl}:`, error);
    return 0;
  }
}
