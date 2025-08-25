import { useMemo } from 'react';

import { useAuth } from '@/contexts/auth';
import { useQueryDaosPublic, useQueryDaos } from '@/lib/graphql';
import type { Dao } from '@/lib/graphql/types';
import type { DaoInfo } from '@/utils/config';

function transformDaoData(dao: Dao, index: number): DaoInfo {

  return {
    id: dao.id,
    name: dao.name,
    code: dao.code,
    daoIcon: dao.logo,
    network: dao.chainName || `Chain ${dao.chainId}`,
    networkIcon: dao.chainLogo,
    proposals: dao.metricsCountProposals,
    favorite: dao.liked,
    settable: true,
    website: dao.endpoint || '',
    indexer: '',
    chainId: dao.chainId.toString(),
    chips: dao.chips,
    lastProposal: dao.lastProposal
  };
}

export function useGraphqlDaoData() {
  const { isAuthenticated } = useAuth();
  
  const publicQuery = useQueryDaosPublic();
  const authQuery = useQueryDaos();
  
  const graphqlData = (isAuthenticated && authQuery.data) ? authQuery.data : publicQuery.data;
  const isLoading = isAuthenticated ? authQuery.isLoading : publicQuery.isLoading;
  const error = isAuthenticated ? authQuery.error : publicQuery.error;

  const daoData = useMemo(() => {
    if (!graphqlData?.daos) return [];

    return graphqlData.daos
      ?.filter((dao) => !dao.tags?.includes('demo'))
      .map((dao, index) => transformDaoData(dao, index));
  }, [graphqlData]);


  return {
    daoData,
    isLoading,
    error: error?.message || null,
    subscribedDaos: isAuthenticated ? (graphqlData?.subscribedDaos || []) : []
  };
}
