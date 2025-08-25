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
