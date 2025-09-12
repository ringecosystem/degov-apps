import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/contexts/auth';
import { useAuthenticatedRequest } from '@/hooks/useAuthenticatedRequest';
import { getToken } from '@/lib/auth/token-manager';

import { createAuthorizedClient, createPublicClient } from './client';
import {
  QUERY_NONCE,
  LOGIN_MUTATION,
  MODIFY_LIKE_DAO_MUTATION,
  QUERY_DAOS,
  LIST_NOTIFICATION_CHANNELS,
  SUBSCRIBED_DAOS,
  SUBSCRIBED_PROPOSALS,
  BIND_NOTIFICATION_CHANNEL,
  RESEND_OTP,
  VERIFY_NOTIFICATION_CHANNEL,
  SUBSCRIBE_PROPOSAL,
  UNSUBSCRIBE_PROPOSAL,
  SUBSCRIBE_DAO,
  UNSUBSCRIBE_DAO,
  UN_SUBSCRIBE_CHANNEL,
  UN_SUBSCRIBE_PROPOSAL
} from './queries';

import type {
  LoginResponse,
  LoginVariables,
  ModifyLikeDaoVariables,
  DaosResponse,
  NonceVariables,
  ListNotificationChannelsResponse,
  BindNotificationChannelVariables,
  BindNotificationChannelResponse,
  VerifyNotificationChannelVariables,
  VerifyNotificationChannelResponse,
  SubscribedDaosResponse,
  SubscribedProposalsResponse,
  SubscribeProposalVariables,
  ProposalSubscriptionResponse,
  UnsubscribeProposalVariables,
  SubscribeDaoVariables,
  DaoSubscriptionResponse,
  UnsubscribeDaoVariables,
  UnSubscribeChannelVariables,
  UnSubscribeChannelResponse,
  UnSubscribeProposalVariables,
  UnSubscribeProposalResponse
} from './types';

export const QUERY_KEYS = {
  nonce: (length: number) => ['nonce', length] as const,
  daos: () => ['daos'] as const,
  notificationChannels: () => ['notificationChannels'] as const,
  subscribedDaos: () => ['subscribedDaos'] as const,
  subscribedProposals: () => ['subscribedProposals'] as const
} as const;

export const useQueryNonce = (length: number = 10) => {
  return useQuery({
    queryKey: QUERY_KEYS.nonce(length),
    queryFn: async () => {
      const client = createPublicClient();
      const variables: NonceVariables = { input: { length } };
      const data = await client.request<{ nonce: string }>(QUERY_NONCE, variables);
      return data.nonce;
    },
    staleTime: 0,
    gcTime: 0
  });
};

export const useLogin = () => {
  const { setToken } = useAuth();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (variables: LoginVariables) => {
      const client = createPublicClient();
      const data = await client.request<LoginResponse>(LOGIN_MUTATION, variables);
      return data.login.token;
    },
    onSuccess: (token) => {
      setToken(token);
      queryClient.invalidateQueries({
        predicate: (query) => query.queryKey[0] === 'daos'
      });
    },
    onError: (error) => {
      console.error('Login failed:', error);
    }
  });
};

export const useQueryDaos = () => {
  const { token: contextToken } = useAuth();
  const token = getToken() || contextToken;

  return useQuery({
    queryKey: [...QUERY_KEYS.daos(), token],
    queryFn: async () => {
      if (!token) throw new Error('No authentication token available');
      const client = createAuthorizedClient(token);
      const data = await client.request<DaosResponse>(QUERY_DAOS);
      return data;
    },
    enabled: !!token, // Only run query when token is available
    staleTime: 1000 * 60 * 5
  });
};

export const useQueryDaosPublic = () => {
  return useQuery({
    queryKey: ['daos-public'],
    queryFn: async () => {
      const client = createPublicClient();
      const data = await client.request<DaosResponse>(QUERY_DAOS);
      return data;
    },
    staleTime: 1000 * 60 * 5
  });
};

export const useModifyLikeDao = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: ModifyLikeDaoVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{ modifyLikeDao: boolean }>(
          MODIFY_LIKE_DAO_MUTATION,
          variables
        );
        return data.modifyLikeDao;
      });
    },
    onMutate: async (variables) => {
      // Cancel any outgoing refetches (so they don't overwrite our optimistic update)
      await queryClient.cancelQueries({ queryKey: QUERY_KEYS.daos() });
      await queryClient.cancelQueries({ queryKey: ['daos-public'] });

      // Snapshot the previous values
      const previousAuthData = queryClient.getQueryData([...QUERY_KEYS.daos(), token]);
      const previousPublicData = queryClient.getQueryData(['daos-public']);

      const isLikeAction = variables.input.action === 'LIKE';

      // Update function to apply the optimistic update
      const updateData = (old: DaosResponse | undefined) => {
        if (!old) return old;

        // Update the liked field in the corresponding DAO object
        const updatedDaos = old.daos.map((dao) => {
          if (dao.code === variables.input.daoCode) {
            return { ...dao, liked: isLikeAction };
          }
          return dao;
        });

        return {
          ...old,
          daos: updatedDaos
        };
      };

      // Optimistically update both authenticated and public caches
      queryClient.setQueryData([...QUERY_KEYS.daos(), token], updateData);
      queryClient.setQueryData(['daos-public'], updateData);

      return { previousAuthData, previousPublicData };
    },
    onError: (error, _variables, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      if (context?.previousAuthData) {
        queryClient.setQueryData([...QUERY_KEYS.daos(), token], context.previousAuthData);
      }
      if (context?.previousPublicData) {
        queryClient.setQueryData(['daos-public'], context.previousPublicData);
      }

      console.error('Failed to modify DAO like status:', error);
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({
        predicate: (query) => query.queryKey[0] === 'daos' || query.queryKey[0] === 'daos-public'
      });
    }
  });
};

export const useLikeDao = () => {
  const modifyLikeMutation = useModifyLikeDao();

  return {
    likeDao: (daoCode: string) => modifyLikeMutation.mutate({ input: { daoCode, action: 'LIKE' } }),
    unlikeDao: (daoCode: string) =>
      modifyLikeMutation.mutate({ input: { daoCode, action: 'UNLIKE' } }),
    ...modifyLikeMutation
  };
};

export const useLogout = () => {
  const { setToken } = useAuth();
  const queryClient = useQueryClient();

  return () => {
    setToken(null);
    queryClient.clear();
  };
};

// Notification Hooks
export const useListNotificationChannels = () => {
  const { token: contextToken } = useAuth();
  const token = getToken() || contextToken;

  return useQuery({
    queryKey: [...QUERY_KEYS.notificationChannels(), token],
    queryFn: async () => {
      if (!token) throw new Error('No authentication token available');
      const client = createAuthorizedClient(token);
      const data = await client.request<ListNotificationChannelsResponse>(
        LIST_NOTIFICATION_CHANNELS
      );
      return data.listNotificationChannels;
    },
    enabled: !!token,
    staleTime: 1000 * 60 * 5
  });
};

export const useSubscribedDaos = () => {
  const { token: contextToken } = useAuth();
  const token = getToken() || contextToken;

  return useQuery({
    queryKey: [...QUERY_KEYS.subscribedDaos(), token],
    queryFn: async () => {
      if (!token) throw new Error('No authentication token available');
      const client = createAuthorizedClient(token);
      const data = await client.request<SubscribedDaosResponse>(SUBSCRIBED_DAOS);
      return data.subscribedDaos;
    },
    enabled: !!token,
    retry: 1,
    staleTime: 1000 * 60 * 5
  });
};

export const useSubscribedProposals = () => {
  const { token: contextToken } = useAuth();
  const token = getToken() || contextToken;

  return useQuery({
    queryKey: [...QUERY_KEYS.subscribedProposals(), token],
    queryFn: async () => {
      if (!token) throw new Error('No authentication token available');
      const client = createAuthorizedClient(token);
      const data = await client.request<SubscribedProposalsResponse>(SUBSCRIBED_PROPOSALS);
      return data.subscribedProposals;
    },
    enabled: !!token,
    retry: 1,
    staleTime: 1000 * 60 * 5
  });
};

export const useBindNotificationChannel = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: BindNotificationChannelVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{
          bindNotificationChannel: BindNotificationChannelResponse;
        }>(BIND_NOTIFICATION_CHANNEL, variables);
        return data.bindNotificationChannel;
      });
    },
    onError: (error) => {
      console.error('Failed to bind notification channel:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.notificationChannels() });
    }
  });
};

export const useResendOTP = () => {
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: BindNotificationChannelVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{ resendOTP: BindNotificationChannelResponse }>(
          RESEND_OTP,
          variables
        );
        return data.resendOTP;
      });
    },
    onError: (error) => {
      console.error('Failed to resend OTP:', error);
    }
  });
};

export const useVerifyNotificationChannel = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: VerifyNotificationChannelVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{
          verifyNotificationChannel: VerifyNotificationChannelResponse;
        }>(VERIFY_NOTIFICATION_CHANNEL, variables);
        return data.verifyNotificationChannel;
      });
    },
    onError: (error) => {
      console.error('Failed to verify notification channel:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.notificationChannels() });
    }
  });
};

export const useSubscribeProposal = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: SubscribeProposalVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{ subscribeProposal: ProposalSubscriptionResponse }>(
          SUBSCRIBE_PROPOSAL,
          variables
        );
        return data.subscribeProposal;
      });
    },
    onError: (error) => {
      console.error('Failed to subscribe to proposal:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.subscribedDaos() });
    }
  });
};

export const useUnsubscribeProposal = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: UnsubscribeProposalVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{ unsubscribeProposal: ProposalSubscriptionResponse }>(
          UNSUBSCRIBE_PROPOSAL,
          variables
        );
        return data.unsubscribeProposal;
      });
    },
    onError: (error) => {
      console.error('Failed to unsubscribe from proposal:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.subscribedDaos() });
    }
  });
};

export const useSubscribeDao = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: SubscribeDaoVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{ subscribeDao: DaoSubscriptionResponse }>(
          SUBSCRIBE_DAO,
          variables
        );
        return data.subscribeDao;
      });
    },
    onError: (error) => {
      console.error('Failed to subscribe to DAO:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.subscribedDaos() });
    }
  });
};

export const useUnsubscribeDao = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: UnsubscribeDaoVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<{ unsubscribeDao: DaoSubscriptionResponse }>(
          UNSUBSCRIBE_DAO,
          variables
        );
        return data.unsubscribeDao;
      });
    },
    onError: (error) => {
      console.error('Failed to unsubscribe from DAO:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.subscribedDaos() });
    }
  });
};

export const useUnSubscribeChannel = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: UnSubscribeChannelVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<UnSubscribeChannelResponse>(
          UN_SUBSCRIBE_CHANNEL,
          variables
        );
        return data;
      });
    },
    onError: (error) => {
      console.error('Failed to subscribe/unsubscribe channel:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.subscribedDaos() });
    }
  });
};

export const useUnSubscribeProposal = () => {
  const { token: contextToken, setToken } = useAuth();
  const token = getToken() || contextToken;
  const queryClient = useQueryClient();
  const { executeWithAuth } = useAuthenticatedRequest();

  return useMutation({
    mutationFn: async (variables: UnSubscribeProposalVariables) => {
      return await executeWithAuth(async (authToken) => {
        const client = createAuthorizedClient(authToken);
        const data = await client.request<UnSubscribeProposalResponse>(
          UN_SUBSCRIBE_PROPOSAL,
          variables
        );
        return data;
      });
    },
    onError: (error) => {
      console.error('Failed to subscribe/unsubscribe proposal:', error);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.subscribedDaos() });
    }
  });
};
