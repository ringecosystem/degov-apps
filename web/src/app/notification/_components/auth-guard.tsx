'use client';

import { ConnectButton } from '@rainbow-me/rainbowkit';

import { NotFoundIcon } from '@/components/icons/not-found-icon';
import { useAccount } from '@/hooks/useAccount';

interface AuthGuardProps {
  children: React.ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { isConnected, isConnecting } = useAccount();

  // Show loading state while connecting
  if (isConnecting) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center">
        <div className="border-foreground h-8 w-8 animate-spin rounded-full border-b-2"></div>
        <p className="text-muted-foreground mt-4 text-sm">Connecting...</p>
      </div>
    );
  }

  if (!isConnected) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center space-y-6">
        <div className="flex flex-col items-center space-y-4">
          <div className="bg-muted flex h-16 w-16 items-center justify-center rounded-full">
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              className="text-muted-foreground"
            >
              <path
                d="M21 18v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v1"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              <polyline
                points="15,10 20,10 20,14 15,14"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              <line
                x1="4"
                y1="12"
                x2="20"
                y2="12"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            </svg>
          </div>

          <div className="space-y-2 text-center">
            <h3 className="text-foreground text-lg font-semibold">Connect Wallet</h3>
            <p className="text-muted-foreground max-w-sm text-sm">
              You need to connect your wallet to access notification settings and manage your
              subscriptions.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
