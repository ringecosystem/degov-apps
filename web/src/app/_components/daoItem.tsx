import Image from 'next/image';
import Link from 'next/link';

import { Separator } from '@/components/ui/separator';

interface DaoItemProps {
  name: string;
  daoIcon: string;
  network: string;
  proposals: number;
  id: string;
}
export const DaoItem = ({ name, daoIcon, network, proposals, id }: DaoItemProps) => {
  return (
    <div className="bg-card flex flex-col gap-[10px] rounded-[14px] p-[10px]">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-[10px]">
          <Image src={daoIcon} alt={name} width={32} height={32} />
          <p className="text-[18px] font-semibold">{name}</p>
        </div>

        <p className="text-muted-foreground text-[14px]">{network}</p>
      </div>
      <Separator className="my-0" />
      <div className="flex items-center justify-between">
        <p className="text-[14px]">{proposals ?? 0} Proposals</p>
        {/* <div className="flex items-center justify-end gap-[10px]">
          <Link
            href={`/setting/${id}`}
            className="cursor-pointer transition-opacity hover:opacity-80"
          >
            <Image src="/setting.svg" alt="setting" width={20} height={20} />
          </Link>
          <button className="cursor-pointer transition-opacity hover:opacity-80">
            <Image src="/favorite.svg" alt="favorite" width={20} height={20} />
          </button>
        </div> */}
      </div>
    </div>
  );
};
