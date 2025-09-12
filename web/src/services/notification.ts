/**
 * NotificationService - Service layer for notification operations
 */

import { getToken } from '@/lib/auth/token-manager';
import { createAuthorizedClient } from '@/lib/graphql/client';
import {
  LIST_NOTIFICATION_CHANNELS,
  SUBSCRIBED_DAOS,
  SUBSCRIBED_PROPOSALS,
  RESEND_OTP,
  VERIFY_NOTIFICATION_CHANNEL,
  SUBSCRIBE_PROPOSAL,
  UNSUBSCRIBE_PROPOSAL,
  SUBSCRIBE_DAO,
  UNSUBSCRIBE_DAO
} from '@/lib/graphql/queries';
import type {
  ListNotificationChannelsResponse,
  SubscribedDaosResponse,
  SubscribedProposalsResponse,
  ResendOTPResponse,
  VerifyNotificationChannelResponse,
  ProposalSubscriptionResponse,
  DaoSubscriptionResponse
} from '@/lib/graphql/types';

export type NotificationChannelType = 'EMAIL';

export interface VerifyNotificationChannelInput {
  type: NotificationChannelType;
  value: string;
  otpCode: string;
}

export interface ProposalSubscriptionInput {
  daoCode: string;
  proposalId: string;
}

export interface DaoSubscriptionInput {
  daoCode: string;
  features?: Array<{
    name: string;
    strategy: string;
  }>;
}

class NotificationServiceClass {
  private getAuthenticatedClient() {
    const token = getToken();
    if (!token) {
      throw new Error('No authentication token available');
    }
    return createAuthorizedClient(token);
  }

  async listNotificationChannels() {
    const client = this.getAuthenticatedClient();
    const data = await client.request<ListNotificationChannelsResponse>(LIST_NOTIFICATION_CHANNELS);
    return data.listNotificationChannels;
  }

  async getSubscribedDaos() {
    const client = this.getAuthenticatedClient();
    const data = await client.request<SubscribedDaosResponse>(SUBSCRIBED_DAOS);
    return data.subscribedDaos;
  }

  async getSubscribedProposals() {
    const client = this.getAuthenticatedClient();
    const data = await client.request<SubscribedProposalsResponse>(SUBSCRIBED_PROPOSALS);
    return data.subscribedProposals;
  }

  async resendOTP(type: NotificationChannelType, value: string) {
    const client = this.getAuthenticatedClient();
    const data = await client.request<{ resendOTP: ResendOTPResponse }>(RESEND_OTP, {
      input: { type, value }
    });
    return data.resendOTP;
  }

  async verifyNotificationChannel(input: VerifyNotificationChannelInput) {
    const client = this.getAuthenticatedClient();
    const data = await client.request<{
      verifyNotificationChannel: VerifyNotificationChannelResponse;
    }>(VERIFY_NOTIFICATION_CHANNEL, { input });
    return data.verifyNotificationChannel;
  }

  async subscribeProposal(input: ProposalSubscriptionInput) {
    const client = this.getAuthenticatedClient();
    const data = await client.request<{ subscribeProposal: ProposalSubscriptionResponse }>(
      SUBSCRIBE_PROPOSAL,
      { input }
    );
    return data.subscribeProposal;
  }

  async unsubscribeProposal(daoCode: string, proposalId: string) {
    const client = this.getAuthenticatedClient();
    const data = await client.request<{ unsubscribeProposal: ProposalSubscriptionResponse }>(
      UNSUBSCRIBE_PROPOSAL,
      { daoCode, proposalId }
    );
    return data.unsubscribeProposal;
  }

  async subscribeDao(input: DaoSubscriptionInput) {
    const client = this.getAuthenticatedClient();
    const data = await client.request<{ subscribeDao: DaoSubscriptionResponse }>(SUBSCRIBE_DAO, {
      input
    });
    return data.subscribeDao;
  }

  async unsubscribeDao(daoCode: string) {
    const client = this.getAuthenticatedClient();
    const data = await client.request<{ unsubscribeDao: DaoSubscriptionResponse }>(
      UNSUBSCRIBE_DAO,
      { daoCode }
    );
    return data.unsubscribeDao;
  }
}

export const NotificationService = new NotificationServiceClass();
