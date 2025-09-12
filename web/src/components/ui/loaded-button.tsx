import { Loader2 } from 'lucide-react';
import React from 'react';

import { cn } from '@/lib/utils';

import { Button } from './button';

interface LoadedButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  isLoading?: boolean;
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
  size?: 'default' | 'sm' | 'lg' | 'icon';
}

export function LoadedButton({
  children,
  isLoading,
  className,
  disabled,
  variant,
  size,
  ...props
}: LoadedButtonProps) {
  return (
    <Button
      className={cn(className)}
      disabled={disabled || isLoading}
      variant={variant}
      size={size}
      {...props}
    >
      {isLoading ? (
        <span className="flex items-center justify-center gap-[5px]">
          <Loader2 className="h-4 w-4 animate-spin" />
          {children}
        </span>
      ) : (
        children
      )}
    </Button>
  );
}
