import { ChevronDown, Power } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useCallback, useMemo } from 'react';
import { useAccount } from 'wagmi';

import { AddressAvatar } from '@/components/address-avatar';
import { AddressResolver } from '@/components/address-resolver';
import ClipboardIconButton from '@/components/clipboard-icon-button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useDisconnectWallet } from '@/hooks/useDisconnectWallet';
import { formatShortAddress } from '@/utils';

import { Button } from '../ui/button';
interface ConnectedProps {
  address: `0x${string}`;
}

export const Connected = ({ address }: ConnectedProps) => {
  const { disconnectWallet } = useDisconnectWallet();
  const { chain } = useAccount();

  const getBlockExplorerUrl = useMemo(() => {
    return `${chain?.blockExplorers?.default?.url}/address/${address}`;
  }, [chain, address]);

  const handleDisconnect = useCallback(() => {
    disconnectWallet(address);
  }, [disconnectWallet, address]);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger>
        <AddressResolver address={address} showShortAddress>
          {(value) => (
            <div className="border-foreground/20 flex cursor-pointer items-center gap-[5px] rounded-[10px] border p-[5px] md:gap-[10px]">
              <AddressAvatar address={address} className="size-[24px] rounded-full" size={24} />
              <span className="text-foreground hidden text-[14px] font-medium md:block">
                {value}
              </span>
              <ChevronDown size={16} className="text-muted-foreground" />
            </div>
          )}
        </AddressResolver>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="bg-card rounded-[26px] p-[20px] shadow-2xl" align="end">
        <div className="flex items-center gap-[10px]">
          <AddressAvatar address={address} className="rounded-full" />
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="text-[18px] font-extrabold text-white/80">
                {formatShortAddress(address)}
              </span>
            </TooltipTrigger>
            <TooltipContent>{address}</TooltipContent>
          </Tooltip>
          <ClipboardIconButton text={address} size={16} className="size-[16px] flex-shrink-0" />
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                href={getBlockExplorerUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="flex cursor-pointer items-center justify-center transition-colors hover:opacity-80"
              >
                <Image
                  src="/external-link.svg"
                  alt="Open in block explorer"
                  width={16}
                  height={16}
                  className="size-[16px] flex-shrink-0 opacity-50"
                />
              </Link>
            </TooltipTrigger>
            <TooltipContent>View on block explorer</TooltipContent>
          </Tooltip>
        </div>
        <DropdownMenuSeparator className="my-[20px]" />
        <div className="flex flex-col justify-center gap-[20px]">
          {/* <Button asChild className="h-[40px] w-full gap-[5px] rounded-[100px]" variant="outline">
            <Link href="/notification">
              <Image
                src="/bell.svg"
                alt="bell"
                width={20}
                height={20}
                className="size-[20px] flex-shrink-0"
              />
              <span className="text-[14px]">Notification</span>
            </Link>
          </Button> */}
          <Button
            onClick={handleDisconnect}
            className="h-[40px] w-full gap-[5px] rounded-[100px]"
            variant="outline"
          >
            <Power size={20} className="text-white" strokeWidth={2} />
            <span className="text-[14px]">Disconnect</span>
          </Button>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
