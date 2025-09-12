'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useCallback } from 'react';
import { toast } from 'react-toastify';

import { useIsMobileAndSubSection } from '@/app/notification/_hooks/isMobileAndSubSection';
import { CustomTable } from '@/components/custom-table';
import type { ColumnType } from '@/components/custom-table';
import { ProposalStatus } from '@/components/proposal-status';
import { useConfirm } from '@/contexts/confirm-context';
import { useSubscribedProposals, useUnsubscribeProposal } from '@/hooks/useNotification';
import type { SubscribedProposalItem } from '@/lib/graphql/types';
import { extractErrorMessage } from '@/utils/graphql-error-handler';

import { Item } from './_components/item';

type ColumnProps = {
  onRemove: (daoCode: string, proposalId: string) => void;
};

const columns = ({ onRemove }: ColumnProps): ColumnType<SubscribedProposalItem>[] => [
  {
    title: 'DAO',
    key: 'daoName',
    width: '20.9%',
    className: 'text-left',
    render: (value: SubscribedProposalItem) => {
      return <div className="truncate">{value.dao?.chainName || 'Unknown Chain'}</div>;
    }
  },
  {
    title: 'Proposal',
    key: 'proposalTitle',
    width: '40.7%',
    className: 'text-left',
    style: { maxWidth: '0' },
    render: (value: SubscribedProposalItem) => {
      return (
        <div title={value.proposal?.title || 'Proposal'} className="truncate">
          {value.proposal?.title || 'Untitled Proposal'}
        </div>
      );
    }
  },

  {
    title: 'Status',
    key: 'proposalStatus',
    width: '28.1%',
    className: 'text-center',
    render: (value: SubscribedProposalItem) => {
      return (
        <div className="flex justify-center">
          <ProposalStatus status={value.proposal?.state as any} />
        </div>
      );
    }
  },
  {
    title: 'Action',
    key: 'action',
    width: '9.6%',
    className: 'text-right',
    render(value: SubscribedProposalItem) {
      return (
        <button
          className="cursor-pointer transition-opacity hover:opacity-80"
          onClick={() => onRemove(value.proposal?.daoCode || '', value.proposal?.proposalId || '')}
        >
          <Image
            src="/unsubscribed.svg"
            alt="unsubscribed"
            width={20}
            height={20}
            className="h-[20px] w-[20px]"
          />
        </button>
      );
    }
  }
];

export default function SubscribedProposalsPage() {
  const isMobileAndSubSection = useIsMobileAndSubSection();
  const { data: subscriptions, isLoading, refetch } = useSubscribedProposals();

  const unsubscribeProposalMutation = useUnsubscribeProposal();
  const { confirm } = useConfirm();

  const handleUnsubscribe = useCallback(
    (daoCode: string, proposalId: string) => {
      if (!daoCode || !proposalId) return;

      confirm({
        title: 'Unsubscribe',
        description: 'Are you sure you want to unsubscribe notifications for this proposal??',
        cancelText: 'Cancel',
        confirmText: 'Confirm',
        onConfirm: () => {
          return unsubscribeProposalMutation.mutateAsync(
            { daoCode, proposalId },
            {
              onSuccess: () => {
                refetch();
              },
              onError: (error: any) => {
                const errorMessage = extractErrorMessage(error) || 'Failed to unsubscribe proposal';
                toast.error(errorMessage);
              }
            }
          );
        }
      });
    },
    [confirm, unsubscribeProposalMutation, refetch]
  );

  return (
    <>
      {isMobileAndSubSection && (
        <Link href={`/notification`} className="flex items-center gap-[5px] md:gap-[10px]">
          <Image
            src="/back.svg"
            alt="back"
            width={32}
            height={32}
            className="size-[32px] flex-shrink-0"
          />
          <h1 className="text-[18px] font-semibold">Subscribed Proposals</h1>
        </Link>
      )}
      <CustomTable<SubscribedProposalItem>
        columns={columns({ onRemove: handleUnsubscribe })}
        dataSource={(subscriptions as SubscribedProposalItem[]) ?? []}
        isLoading={isLoading}
        rowKey={(record) => `${record.proposal?.daoCode}-${record.proposal?.proposalId}` || ''}
        className="hidden md:block"
      />

      <div className="mt-[15px] flex flex-col gap-[15px] md:hidden">
        {(subscriptions || []).map((subscription) => (
          <Item
            key={`${subscription.proposal?.daoCode}-${subscription.proposal?.proposalId}`}
            id={subscription.proposal?.proposalId || ''}
            name={subscription.proposal?.title || 'Untitled Proposal'}
            daoName={subscription.dao?.name || 'Unknown DAO'}
            daoLogo={subscription.dao?.logo || '/example/dao-placeholder.svg'}
            status={subscription.proposal?.state}
            onRemove={() =>
              handleUnsubscribe(
                subscription.proposal?.daoCode || '',
                subscription.proposal?.proposalId || ''
              )
            }
          />
        ))}
      </div>
    </>
  );
}
