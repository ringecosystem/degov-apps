import { gql } from 'graphql-request';

// Query for getting nonce
export const QUERY_NONCE = gql`
  query QueryNonce($input: GetNonceInput!) {
    nonce(input: $input)
  }
`;

// Mutation for login
export const LOGIN_MUTATION = gql`
  mutation LoginMutation($input: LoginInput!) {
    login(input: $input) {
      token
    }
  }
`;

// Mutation for modifying like status of DAO
export const MODIFY_LIKE_DAO_MUTATION = gql`
  mutation ModifyLikeMutation($input: ModifyLikeDaoInput!) {
    modifyLikeDao(input: $input)
  }
`;

// Query for getting DAOs list
export const QUERY_DAOS = gql`
  query QueryDaos {
    daos {
      id
      name
      code
      chainId
      chainName
      seq
      timeSyncd
      ctime
      utime
      state
      tags
      chips {
        id
        daoCode
        chipCode
        flag
        additional
        ctime
        utime
      }
      metricsCountMembers
      metricsCountProposals
      metricsCountVote
      metricsSumPower
      endpoint
      logo
      chainLogo
      liked
      lastProposal {
        chainId
        daoCode
        proposalId
        proposalLink
        proposalCreatedAt
        proposalAtBlock
        state
        timeNextTrack
        timesTrack
      }
    }
  }
`;

// Notification Queries
export const LIST_NOTIFICATION_CHANNELS = gql`
  query ListNotificationChannels {
    listNotificationChannels {
      id
      channelType
      channelValue
      verified
      payload
      ctime
    }
  }
`;

export const SUBSCRIBED_DAOS = gql`
  query SubscribedDaos {
    subscribedDaos {
      dao {
        code
        name
      }
      features {
        name
        strategy
      }
    }
  }
`;

export const SUBSCRIBED_PROPOSALS = gql`
  query SubscribedProposals {
    subscribedProposals {
      proposal {
        proposalId
        daoCode
        state
        title
        proposalCreatedAt
      }
      dao {
        code
        name
        logo
        chainName
        chainLogo
      }
      features {
        name
        strategy
      }
    }
  }
`;

// Notification Mutations
export const BIND_NOTIFICATION_CHANNEL = gql`
  mutation BindNotificationsChannel($type: NotificationChannelType!, $value: String!) {
    bindNotificationChannel(input: { type: $type, value: $value }) {
      id
      code
      expiration
      message
      rateLimit
    }
  }
`;

export const RESEND_OTP = gql`
  mutation ResendOTP($input: BaseNotificationChannelInput!) {
    resendOTP(input: $input) {
      code
      message
      rateLimit
    }
  }
`;

export const VERIFY_NOTIFICATION_CHANNEL = gql`
  mutation VerifyNotificationChannel($input: VerifyNotificationChannelInput!) {
    verifyNotificationChannel(input: $input) {
      code
      message
    }
  }
`;

export const SUBSCRIBE_PROPOSAL = gql`
  mutation SubscribeProposal($input: SubscribeProposalInput!) {
    subscribeProposal(input: $input) {
      code
      message
    }
  }
`;

export const UNSUBSCRIBE_PROPOSAL = gql`
  mutation UnsubscribeProposal($daoCode: String!, $proposalId: String!) {
    unsubscribeProposal(input: { daoCode: $daoCode, proposalId: $proposalId }) {
      daoCode
      proposalId
      state
    }
  }
`;

export const SUBSCRIBE_DAO = gql`
  mutation SubscribeDao($input: SubscribeDaoInput!) {
    subscribeDao(input: $input) {
      code
      message
    }
  }
`;

export const UNSUBSCRIBE_DAO = gql`
  mutation UnsubscribeDao($daoCode: String!) {
    unsubscribeDao(input: { daoCode: $daoCode }) {
      daoCode
      state
    }
  }
`;

export const UN_SUBSCRIBE_CHANNEL = gql`
  mutation UN_SubscribeChannel($daoCode: String!, $features: [SubscriptionFeatureInput!]) {
    subscribeDao(input: { daoCode: $daoCode, features: $features }) {
      daoCode
      state
    }
    unsubscribeDao(input: { daoCode: $daoCode }) {
      daoCode
      state
    }
  }
`;

export const UN_SUBSCRIBE_PROPOSAL = gql`
  mutation UN_SubscribeProposal(
    $daoCode: String!
    $proposalId: String!
    $features: [SubscriptionFeatureInput!]
  ) {
    subscribeProposal(input: { daoCode: $daoCode, proposalId: $proposalId, features: $features }) {
      state
      proposalId
      daoCode
    }
    unsubscribeProposal(input: { daoCode: $daoCode, proposalId: $proposalId }) {
      state
      proposalId
      daoCode
    }
  }
`;
