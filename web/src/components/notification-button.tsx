'use client';

import { useRouter } from 'next/navigation';

import { NotificationIcon } from './icons/notification-icon';
import { Button } from './ui/button';
export function NotificationButton() {
  const router = useRouter();

  const handleClick = () => {
    router.push('/notification/subscription');
  };

  return (
    <Button
      className="border-border bg-background hidden h-[36px] w-[36px] items-center justify-center rounded-[10px] border lg:flex"
      variant="outline"
      onClick={handleClick}
    >
      <NotificationIcon className="h-[20px] w-[20px]" />
    </Button>
  );
}
