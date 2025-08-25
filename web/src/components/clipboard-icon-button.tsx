'use client';
import { Check } from 'lucide-react';
import Image from 'next/image';
import { useCallback, useEffect, useState, useRef } from 'react';
import { useCopyToClipboard } from 'react-use';

import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

interface ClipboardIconButtonProps {
  text?: string;
  size?: string | number;
  color?: string;
  className?: string;
  strokeWidth?: number;
}

const ClipboardIconButton = ({
  text = '',
  size,
  color,
  className,
  strokeWidth = 1
}: ClipboardIconButtonProps) => {
  const [state, copyToClipboard] = useCopyToClipboard();
  const [copied, setCopied] = useState(false);
  const [open, setOpen] = useState(false);

  const enterTimeout = useRef<NodeJS.Timeout | undefined>(undefined);
  const leaveTimeout = useRef<NodeJS.Timeout | undefined>(undefined);

  const handleCopy = useCallback(() => {
    if (!text) return;
    copyToClipboard(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1000);
  }, [copyToClipboard, text]);

  useEffect(() => {
    if (state.error) {
      console.error('Copy failed:', state.error);
    }
  }, [state]);

  const handleMouseEnter = useCallback(() => {
    clearTimeout(leaveTimeout.current);
    enterTimeout.current = setTimeout(() => {
      setOpen(true);
    }, 300);
  }, []);

  const handleMouseLeave = useCallback(() => {
    clearTimeout(enterTimeout.current);
    leaveTimeout.current = setTimeout(() => {
      setOpen(false);
    }, 300);
  }, []);

  useEffect(() => {
    return () => {
      clearTimeout(enterTimeout.current);
      clearTimeout(leaveTimeout.current);
    };
  }, []);

  if (!text) return null;

  return (
    <Tooltip open={open}>
      <TooltipTrigger asChild>
        <div
          onClick={handleCopy}
          className="flex flex-shrink-0 cursor-pointer items-center"
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          <Check
            strokeWidth={strokeWidth}
            size={size}
            color={color}
            className={cn(
              'text-muted-foreground hover:text-muted-foreground/80',
              className,
              copied ? 'block' : 'hidden'
            )}
          />
          <Image
            src="/copy.svg"
            alt="Copy"
            width={size as number}
            height={size as number}
            className={cn(
              'text-muted-foreground hover:text-muted-foreground/80 flex-shrink-0 opacity-50',
              className,
              copied ? 'hidden' : 'block'
            )}
          />
        </div>
      </TooltipTrigger>
      <TooltipContent>{copied ? 'Copied!' : 'Copy to clipboard'}</TooltipContent>
    </Tooltip>
  );
};

export default ClipboardIconButton;
