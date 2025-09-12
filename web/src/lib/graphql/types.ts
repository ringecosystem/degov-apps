// GraphQL Types for DeGov API

export interface NonceResponse {
  nonce: string;
}

export interface LoginResponse {
  login: {
    token: string;
  };
}

export interface LoginInput {
  message: string;
  signature: string;
}

export interface ModifyLikeDaoInput {
  daoCode: string;
  action: 'LIKE' | 'UNLIKE';
}

export interface Chip {
  id: string;
  daoCode: string;
  chipCode: string;
  flag: string;
  additional: string;
  ctime: string;
  utime: string;
}

export interface LastProposal {
  chainId: number;
  daoCode: string;
  proposalId: string;
  proposalLink: string;
  proposalCreatedAt: string;
  proposalAtBlock: number;
  state: string;
  timeNextTrack: string;
  timesTrack: number;
}

export interface Dao {
  id: string;
  name: string;
  code: string;
  chainId: number;
  chainName: string;
  seq: number;
  timeSyncd: string;
  ctime: string;
  utime: string;
  state: 'ACTIVE' | 'INACTIVE';
  tags: string[];
  chips: Chip[];
  metricsCountMembers: number;
  metricsCountProposals: number;
  metricsCountVote: number;
  metricsSumPower: string;
  endpoint: string;
  logo: string;
  chainLogo: string;
  liked: boolean;
  lastProposal: LastProposal | null;
}

export interface LikedDao {
  code: string;
}

export interface SubscribedDao {
  code: string;
}

export interface DaosResponse {
  daos: Dao[];
  subscribedDaos: SubscribedDao[];
}

// GraphQL Variables Types
export interface GetNonceInput {
  length: number;
}

export interface NonceVariables {
  input: GetNonceInput;
}

export interface LoginVariables {
  input: LoginInput;
}

export interface ModifyLikeDaoVariables {
  input: ModifyLikeDaoInput;
}

// Notification Types
export type NotificationChannelType = 'EMAIL';

export interface NotificationChannel {
  id: string;
  channelType: NotificationChannelType;
  channelValue: string;
  verified: boolean;
  payload?: string;
  ctime: string;
}

export interface ListNotificationChannelsResponse {
  listNotificationChannels: NotificationChannel[];
}

export interface BindNotificationChannelInput {
  type: NotificationChannelType;
  value: string;
}

export interface BindNotificationChannelResponse {
  id: string;
  code: number;
  expiration: number;
  message?: string;
  rateLimit: number;
}

export interface ResendOTPResponse {
  code: number;
  message?: string;
  rateLimit: number;
}

export interface VerifyNotificationChannelInput {
  id: string;
  otpCode: string;
}

export interface VerifyNotificationChannelResponse {
  code: number;
  message?: string;
}

export interface SubscriptionFeatureInput {
  type: string;
  enabled: boolean;
}

export interface ProposalSubscriptionInput {
  daoCode: string;
  proposalId: string;
  features?: SubscriptionFeatureInput[];
}

export interface ProposalSubscriptionResponse {
  code: number;
  message?: string;
}

export interface SubscribedFeature {
  name: string;
  strategy: string;
}

export interface SubscribedDaoItem extends Record<string, unknown> {
  dao: {
    code: string;
    name: string;
  };
  features: SubscribedFeature[];
}

export interface SubscribedDaosResponse {
  subscribedDaos: SubscribedDaoItem[];
}

export interface SubscribedProposal {
  proposalId: string;
  daoCode: string;
  state: string;
  title: string;
  proposalCreatedAt: string;
}

export interface SubscribedProposalItem extends Record<string, unknown> {
  proposal: SubscribedProposal;
  dao: {
    code: string;
    name: string;
    logo: string;
    chainName: string;
    chainLogo: string;
  };
  features: SubscribedFeature[];
}

export interface SubscribedProposalsResponse {
  subscribedProposals: SubscribedProposalItem[];
}

export interface DaoSubscriptionInput {
  daoCode?: string;
  features?: SubscriptionFeatureInput[];
}

export interface DaoSubscriptionResponse {
  code: number;
  message?: string;
}

// Notification Variables Types
export interface BindNotificationChannelVariables {
  type: NotificationChannelType;
  value: string;
}

export interface VerifyNotificationChannelVariables {
  id: string;
  otpCode: string;
}

export interface SubscribeProposalVariables {
  daoCode: string;
  proposalId: string;
  features?: SubscriptionFeatureInput[];
}

export interface UnsubscribeProposalVariables {
  daoCode: string;
  proposalId: string;
}

export interface SubscribeDaoVariables {
  daoCode: string;
  features?: SubscriptionFeatureInput[];
}

export interface UnsubscribeDaoVariables {
  daoCode: string;
}

export interface UnSubscribeChannelVariables {
  daoCode: string;
  features: SubscriptionFeatureInput[];
}

export interface UnSubscribeChannelResponse {
  subscribeDao: DaoSubscriptionResponse;
  unsubscribeDao: DaoSubscriptionResponse;
}

export interface UnSubscribeProposalVariables {
  daoCode: string;
  proposalId: string;
  features: SubscriptionFeatureInput[];
}

export interface UnSubscribeProposalResponse {
  subscribeProposal: ProposalSubscriptionResponse;
  unsubscribeProposal: ProposalSubscriptionResponse;
}
