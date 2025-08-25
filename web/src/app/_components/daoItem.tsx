import Image from 'next/image';
import Link from 'next/link';

import { LikeButton } from '@/components/like-button';
import { Separator } from '@/components/ui/separator';
import type { DaoInfo } from '@/utils/config';
import { formatNetworkName, formatTimeAgo } from '@/utils/helper';

type DaoItemProps = DaoInfo;

export const DaoItem = (dao: DaoItemProps) => {
  const { name, daoIcon, network, proposals, website, favorite, lastProposal } = dao;

  return (
    <div className="bg-card flex flex-col gap-[10px] rounded-[14px] p-[10px]">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-[10px]">
          <Link
            href={website}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-[10px] hover:underline"
          >
            <Image src={daoIcon} alt={name} width={32} height={32} className="rounded-full" />
            <p className="text-[18px] font-semibold">{name}</p>
          </Link>
        </div>

        <p className="text-muted-foreground text-[14px]">{formatNetworkName(network)}</p>
      </div>
      <Separator className="my-0" />
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-[4px]">
          <p className="text-[14px]">{proposals ?? 0} Proposals</p>
        </div>
        <div className="flex items-center justify-end gap-[10px]">
          {/* <Link
            href={`/setting/${id}`}
            className="cursor-pointer transition-opacity hover:opacity-80"
          >
            <Image src="/setting.svg" alt="setting" width={20} height={20} />
          </Link> */}
          <LikeButton dao={dao} isLiked={favorite} />
        </div>
      </div>
    </div>
  );
};
