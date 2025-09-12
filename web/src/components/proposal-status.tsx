import { cn } from '@/lib/utils';

export enum ProposalState {
  Pending = 'PENDING',
  Active = 'ACTIVE',
  Canceled = 'CANCELED',
  Defeated = 'DEFEATED',
  Succeeded = 'SUCCEEDED',
  Queued = 'QUEUED',
  Expired = 'EXPIRED',
  Executed = 'EXECUTED'
}
export const getDisplayText = (status: ProposalState) => {
  switch (status) {
    case ProposalState.Pending:
      return 'Pending';
    case ProposalState.Active:
      return 'Active';
    case ProposalState.Canceled:
      return 'Canceled';
    case ProposalState.Defeated:
      return 'Defeated';
    case ProposalState.Succeeded:
      return 'Succeeded';
    case ProposalState.Queued:
      return 'Queued';
    case ProposalState.Expired:
      return 'Expired';
    case ProposalState.Executed:
      return 'Executed';
    default:
      return '-';
  }
};

export const getStatusColor = (status: ProposalState) => {
  switch (status) {
    case ProposalState.Pending:
      return {
        bg: 'bg-pending/10',
        text: 'text-pending'
      };
    case ProposalState.Active:
      return {
        bg: 'bg-active/10',
        text: 'text-active'
      };
    case ProposalState.Canceled:
      return {
        bg: 'bg-canceled/10',
        text: 'text-canceled'
      };
    case ProposalState.Defeated:
      return {
        bg: 'bg-defeated/10',
        text: 'text-defeated'
      };
    case ProposalState.Succeeded:
      return {
        bg: 'bg-succeeded/10',
        text: 'text-succeeded'
      };
    case ProposalState.Queued:
      return {
        bg: 'bg-succeeded/10',
        text: 'text-succeeded'
      };
    case ProposalState.Expired:
      return {
        bg: 'bg-succeeded/10',
        text: 'text-succeeded'
      };
    case ProposalState.Executed:
      return {
        bg: 'bg-executed/10',
        text: 'text-executed'
      };
    default:
      return {
        bg: '',
        text: ''
      };
  }
};

interface ProposalStatusProps {
  status: ProposalState;
}
export function ProposalStatus({ status }: ProposalStatusProps) {
  return (
    <span
      className={cn(
        'inline-block rounded-[14px] px-[18px] py-[4px] text-[14px] font-normal',
        getStatusColor(status).bg,
        getStatusColor(status).text
      )}
    >
      {getDisplayText(status)}
    </span>
  );
}
