'use client';

import Image from 'next/image';
import Link from 'next/link';

import { ConnectButton } from '@/components/connect-button';
import { useAccount } from '@/hooks/useAccount';

import { NotificationButton } from '../notification-button';
export function Header() {
  // import { ThemeButton } from '../theme-button';
  const { isConnected } = useAccount();
  return (
    <header className="border-border border-b py-[10px] md:py-[20px]">
      <div className="mx-auto flex items-center justify-between px-[10px] md:container">
        <Link href="/">
          <Image
            src="/logo.svg"
            alt="DeGov.AI"
            width={117.2}
            height={24}
            className="block md:hidden"
          />
          <Image
            src="/logo.svg"
            alt="DeGov.AI"
            width={127}
            height={26}
            className="hidden md:block"
          />
        </Link>
        <div className="flex items-center gap-[5px] md:gap-[10px]">
          <ConnectButton />
          {isConnected && <NotificationButton />}
          {/* <ThemeButton /> */}
        </div>
      </div>
    </header>
  );
}
