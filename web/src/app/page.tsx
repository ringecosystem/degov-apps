'use client';
import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect, useCallback, useMemo } from 'react';

import type { ColumnType } from '@/components/custom-table';
import { CustomTable } from '@/components/custom-table';
import { SortableCell } from '@/components/sortable-cell';
import { Button } from '@/components/ui/button';
import TagGroup from '@/components/ui/tag-group';
import { useGraphqlDaoData } from '@/hooks/useGraphqlDaoData';
import type { DaoInfo } from '@/utils/config';
import { formatNetworkName } from '@/utils/helper';

import { DaoList } from './_components/daoList';
import { MobileSearchDialog } from './_components/MobileSearchDialog';

type SortState = 'asc' | 'desc';

export default function Home() {
  const { daoData, isLoading } = useGraphqlDaoData();

  const [sortState, setSortState] = useState<SortState | undefined>(undefined);
  const [searchQuery, setSearchQuery] = useState('');
  const [openSearchDialog, setOpenSearchDialog] = useState(false);
  const [selectedNetwork, setSelectedNetwork] = useState<string>('');

  const filteredAndSortedData = useMemo(() => {
    let filtered = daoData;

    if (selectedNetwork) {
      filtered = daoData.filter(
        (dao) => dao.network.toLowerCase() === selectedNetwork.toLowerCase()
      );
    } else if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase().trim();
      filtered = daoData.filter(
        (dao) =>
          dao.name.toLowerCase().includes(query) ||
          dao.network.toLowerCase().includes(query) ||
          dao.code.toLowerCase().includes(query)
      );
    }

    if (sortState) {
      filtered = [...filtered].sort((a, b) => {
        const aProposals = a.proposals || 0;
        const bProposals = b.proposals || 0;
        return sortState === 'asc' ? aProposals - bProposals : bProposals - aProposals;
      });
    }

    return filtered;
  }, [daoData, searchQuery, sortState, selectedNetwork]);

  const columns: ColumnType<DaoInfo>[] = [
    {
      title: 'Name',
      key: 'name',
      className: 'w-[34%] text-left',
      render(value) {
        return (
          <div className="flex items-center gap-[10px]">
            <Link
              href={value?.website}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-[10px] hover:underline"
            >
              <Image
                src={value?.daoIcon}
                alt="dao"
                width={34}
                height={34}
                className="rounded-full"
              />
              <span className="text-[16px]">{value?.name}</span>
            </Link>
            <TagGroup dao={value} />
          </div>
        );
      }
    },
    {
      title: 'Network',
      key: 'network',
      className: 'w-[28%] text-left',
      render(value) {
        return (
          <div
            className="flex cursor-pointer items-center justify-start gap-[10px] transition-opacity hover:opacity-80"
            onClick={() => {
              setSelectedNetwork(value?.network || '');
              setSearchQuery('');
            }}
          >
            <Image
              src={value?.networkIcon}
              alt="network"
              width={24}
              height={24}
              className="rounded-full"
            />
            <span className="text-[16px]">{formatNetworkName(value?.network)}</span>
          </div>
        );
      }
    },
    {
      title: <SortableCell onClick={setSortState} sortState={sortState} />,
      key: 'proposals',
      className: 'w-[28%] text-center',
      render(value) {
        return <span className="text-[16px]">{value?.proposals}</span>;
      }
    }
    // {
    //   title: 'Action',
    //   key: 'action',
    //   className: 'w-[20%] text-right',
    //   render(value) {
    //     return (
    //       <div className="flex items-center justify-end gap-[10px]">
    //         <Link
    //           href={`/setting/${value?.id}`}
    //           className="cursor-pointer transition-opacity hover:opacity-80"
    //         >
    //           <Image src="/setting.svg" alt="setting" width={20} height={20} />
    //         </Link>
    //         <button className="cursor-pointer transition-opacity hover:opacity-80">
    //           <Image src="/favorite.svg" alt="favorite" width={20} height={20} />
    //         </button>
    //       </div>
    //     );
    //   }
    // }
  ];

  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
    setSelectedNetwork('');
    setOpenSearchDialog(false);
  }, []);

  const handleDesktopSearch = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
    setSelectedNetwork('');
  }, []);

  const clearSearch = useCallback(() => {
    setSearchQuery('');
    setSelectedNetwork('');
  }, []);

  const openMobileSearch = useCallback(() => {
    setOpenSearchDialog(true);
  }, []);

  useEffect(() => {
    return () => {
      setSortState(undefined);
    };
  }, []);

  return (
    <div className="container flex flex-col gap-[20px]">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-[10px]">
          <span className="text-[18px] font-semibold">
            All DAOs({filteredAndSortedData.length}
            {(searchQuery || selectedNetwork) &&
              daoData.length !== filteredAndSortedData.length && (
                <span className="text-muted-foreground">/{daoData.length}</span>
              )}
            )
          </span>
          {selectedNetwork && (
            <div className="bg-muted text-muted-foreground hidden items-center gap-[5px] rounded-[12px] px-[8px] py-[4px] text-[12px] font-medium md:flex">
              <span>Network: {formatNetworkName(selectedNetwork)}</span>
              <button
                onClick={clearSearch}
                className="cursor-pointer rounded-full p-[2px] transition-colors hover:opacity-80"
                title="Clear network filter"
              >
                <Image src="/close.svg" alt="clear" width={10} height={10} />
              </button>
            </div>
          )}
        </div>
        <div className="flex items-center gap-[20px]">
          <div className="bg-card flex h-[36px] w-[109px] items-center gap-[13px] rounded-[19px] border px-[17px] py-[9px] md:h-auto md:w-[388px] md:gap-[10px]">
            <Image src="/search.svg" alt="search" width={16} height={16} />
            <input
              className="placeholder:text-muted-foreground hidden h-[17px] w-full outline-none placeholder:text-[14px] md:block"
              placeholder="Search by DAO name or Chain name"
              value={searchQuery}
              onChange={handleDesktopSearch}
            />
            <button
              className="text-muted-foreground block text-[14px] md:hidden"
              onClick={openMobileSearch}
            >
              Search
            </button>
            {(searchQuery || selectedNetwork) && (
              <button
                onClick={clearSearch}
                className="text-muted-foreground hover:text-foreground hidden items-center justify-center md:flex"
                title="Clear search"
              >
                <Image src="/close.svg" alt="clear" width={12} height={12} />
              </button>
            )}
          </div>
          <div className="fixed right-0 bottom-[20px] left-[20px] grid grid-cols-[calc(100%-20px)_calc(50%-20px)] gap-[20px] md:static md:grid-cols-1 md:justify-end">
            <Button variant="outline" className="hidden rounded-[100px]" asChild>
              <Link href="/add/existing">Add Existing DAO</Link>
            </Button>
            <Button
              variant="outline"
              className="!border-foreground rounded-[100px] p-[10px]"
              asChild
            >
              <Link
                href="https://docs.google.com/forms/u/1/d/e/1FAIpQLSdYjX87_xxTQFLl-brEj87vxU3ucH682nYy3bGUNpR4nL9HaQ/viewform"
                target="_blank"
                rel="noopener noreferrer"
              >
                With Assistance
              </Link>
            </Button>
          </div>
        </div>
      </div>

      {
        // <>
        //   {filteredAndSortedData.length === 0 && searchQuery ? (
        //     <div className="text-muted-foreground py-8 text-center">
        //       <div className="mb-4">
        //         <Image
        //           src="/empty.svg"
        //           alt="No results"
        //           width={64}
        //           height={64}
        //           className="mx-auto"
        //         />
        //       </div>
        //       <p className="text-lg">No DAOs found</p>
        //       <p className="text-sm">Try searching with different keywords</p>
        //       <button
        //         onClick={clearSearch}
        //         className="mt-2 text-blue-500 underline hover:text-blue-600"
        //       >
        //         View all DAOs
        //       </button>
        //     </div>
        //   ) : (

        //   )}
        // </>
        <>
          <CustomTable
            columns={columns}
            dataSource={filteredAndSortedData}
            className="hidden md:block"
            rowKey="id"
            isLoading={isLoading}
            emptyText="No DAOs found"
          />
          <DaoList
            daoInfo={filteredAndSortedData}
            isLoading={isLoading}
            onNetworkClick={(network) => {
              setSelectedNetwork(network);
              setSearchQuery('');
            }}
          />
        </>
      }
      <MobileSearchDialog
        open={openSearchDialog}
        onOpenChange={setOpenSearchDialog}
        onConfirm={handleSearch}
        initialQuery={searchQuery}
      />
    </div>
  );
}
