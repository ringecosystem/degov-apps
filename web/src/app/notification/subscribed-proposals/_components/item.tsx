// id: 3,
// name: 'Proposal 3: Community Initiative',
// daoName: 'DAO Name 1',
// daoLogo: '/example/dao1.svg',
// status: ProposalState.Succeeded

import Image from 'next/image';

import { ProposalState } from '@/components/proposal-status';
import { ProposalStatus } from '@/components/proposal-status';

type ItemProps = {
  id: string;
  name: string;
  daoName: string;
  daoLogo: string;
  status: string | ProposalState;
  onRemove: () => void;
};

export const Item = ({ id: _id, name, daoName, daoLogo, status, onRemove }: ItemProps) => {
  const normalizedStatus = typeof status === 'string' 
    ? (ProposalState[status as keyof typeof ProposalState] || ProposalState.Pending)
    : status;

  return (
    <div className="flex items-center justify-between gap-[10px]">
      <div className="bg-card flex flex-1 flex-col gap-[10px] rounded-[14px] p-[10px]">
        <h3 className="text-[14px]">{name}</h3>
        <div className="flex items-center justify-between gap-[10px]">
          <div className="flex items-center gap-[5px]">
            <Image src={daoLogo} alt={daoName} width={16} height={16} />
            <span className="text-muted-foreground text-[12px]">{daoName}</span>
          </div>
          <ProposalStatus status={normalizedStatus} />
        </div>
      </div>
      <button className="cursor-pointer transition-opacity hover:opacity-80" onClick={onRemove}>
        <Image
          src="/unsubscribed.svg"
          alt="unsubscribed"
          width={20}
          height={20}
          className="h-[20px] w-[20px]"
        />
      </button>
    </div>
  );
};
