'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useCallback, useMemo } from 'react';
import { toast } from 'react-toastify';

import { useIsMobileAndSubSection } from '@/app/notification/_hooks/isMobileAndSubSection';
import { CustomTable } from '@/components/custom-table';
import type { ColumnType } from '@/components/custom-table';
import { useConfirm } from '@/contexts/confirm-context';
import { useSubscribedDaos, useUnsubscribeDao } from '@/hooks/useNotification';
import { useQueryDaos } from '@/lib/graphql/hooks';
import type { SubscribedDaoItem, Dao } from '@/lib/graphql/types';
import { extractErrorMessage } from '@/utils/graphql-error-handler';

import { Item } from './_components/item';

type EnhancedSubscribedDaoItem = SubscribedDaoItem &
  Record<string, unknown> & {
    enhancedDao?: Dao;
  };

type ColumnProps = {
  onRemove: (daoCode: string) => void;
};

const columns = ({ onRemove }: ColumnProps): ColumnType<EnhancedSubscribedDaoItem>[] => [
  {
    title: 'Name',
    key: 'dao',
    width: 375,
    render: (value) => (
      <div className="flex items-center gap-2">
        <div className="bg-background flex h-8 w-8 items-center justify-center rounded-full">
          <Image
            src={value.enhancedDao?.logo ?? ''}
            alt={value.dao?.name ?? ''}
            width={24}
            height={24}
            className="rounded-full"
          />
        </div>
        <span>{value.dao?.name || 'Unknown DAO'}</span>
      </div>
    )
  },
  {
    title: 'Network',
    key: 'network',
    width: 375,
    className: 'text-center',
    render(value) {
      return (
        <div className="flex items-center justify-center gap-[10px]">
          <Image
            src={value.enhancedDao?.chainLogo ?? ''}
            alt={value.enhancedDao?.chainName ?? ''}
            width={16}
            height={17}
            className="rounded-full"
          />
          <span className="text-[16px]">{value.enhancedDao?.chainName || 'Unknown Network'}</span>
        </div>
      );
    }
  },
  {
    title: 'Action',
    key: 'action',
    width: 60,
    className: 'text-right',
    render(value) {
      return (
        <button
          className="cursor-pointer transition-opacity hover:opacity-80"
          onClick={() => onRemove(value.dao?.code || '')}
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

export default function SubscribedDAOsPage() {
  const { data: subscriptions, isLoading, refetch } = useSubscribedDaos();
  const { data: daosData, isLoading: isDaosLoading } = useQueryDaos();
  const unsubscribeMutation = useUnsubscribeDao();
  const { confirm } = useConfirm();
  const isMobileAndSubSection = useIsMobileAndSubSection();

  const enhancedSubscriptions = useMemo(() => {
    if (!subscriptions || !daosData?.daos) return [];

    return subscriptions.map((subscription): EnhancedSubscribedDaoItem => {
      const enhancedDao = daosData.daos.find((dao) => dao.code === subscription.dao.code);
      return {
        ...subscription,
        enhancedDao
      };
    });
  }, [subscriptions, daosData?.daos]);

  const handleUnsubscribe = useCallback(
    (daoCode: string) => {
      if (!daoCode) return;

      confirm({
        title: 'Unsubscribe',
        description: 'Are you sure you want to unsubscribe from this DAO?',
        cancelText: 'Cancel',
        confirmText: 'Confirm',
        onConfirm: () => {
          return unsubscribeMutation.mutateAsync(daoCode, {
            onSuccess: () => {
              refetch();
            },
            onError: (error: any) => {
              const errorMessage = extractErrorMessage(error) || 'Failed to unsubscribe DAO';
              toast.error(errorMessage);
            }
          });
        }
      });
    },
    [confirm, unsubscribeMutation, refetch]
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
          <h1 className="text-[18px] font-semibold">Subscribed DAOs</h1>
        </Link>
      )}
      <CustomTable<EnhancedSubscribedDaoItem>
        columns={columns({ onRemove: handleUnsubscribe })}
        dataSource={(enhancedSubscriptions as EnhancedSubscribedDaoItem[]) ?? []}
        isLoading={isLoading || isDaosLoading}
        rowKey={(record) => record.dao?.code || ''}
        className="hidden md:block"
      />

      <div className="mt-[15px] flex flex-col gap-[15px] md:hidden">
        {enhancedSubscriptions.map((subscription) => (
          <Item
            key={subscription.dao?.code}
            id={subscription.dao?.code || ''}
            name={subscription.dao?.name || 'Unknown DAO'}
            logo={subscription.enhancedDao?.logo || ''}
            network={subscription.enhancedDao?.chainName || 'Unknown Network'}
            networkLogo={subscription.enhancedDao?.chainLogo || ''}
            onRemove={() => handleUnsubscribe(subscription.dao?.code || '')}
          />
        ))}
      </div>
    </>
  );
}
