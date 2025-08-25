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
    subscribedDaos {
      code
    }
  }
`;
