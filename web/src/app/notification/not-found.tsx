import Link from 'next/link';

import { NotFoundIcon } from '@/components/icons/not-found-icon';

export default function NotFound() {
  return (
    <div className="flex h-full flex-col items-center justify-center space-y-6 text-center">
      <NotFoundIcon className="text-muted-foreground h-16 w-auto" />

      <div className="space-y-2">
        <h2 className="text-foreground text-xl font-semibold">Page Not Found</h2>
        <p className="text-muted-foreground">
          The page you are looking for doesn&apos;t exist or has been moved.
        </p>
      </div>

      <Link
        href="/notification"
        className="bg-primary text-primary-foreground hover:bg-primary/90 inline-flex items-center rounded-md px-4 py-2 text-sm font-medium transition-colors"
      >
        Back to Notifications
      </Link>
    </div>
  );
}
