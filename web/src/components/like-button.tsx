'use client';

import Image from 'next/image';
import { useOptimistic, useTransition, useState } from 'react';
import { toast } from 'react-toastify';

import { useRequireAuth } from '@/hooks/useRequireAuth';
import { useLikeDao } from '@/lib/graphql/hooks';
import { cn } from '@/lib/utils';
import type { DaoInfo } from '@/utils/config';

interface LikeButtonProps {
  dao: DaoInfo;
  isLiked: boolean;
  className?: string;
}

export const LikeButton = ({ dao, isLiked, className }: LikeButtonProps) => {
  const { withAuth, isAuthenticating } = useRequireAuth();
  const [, startTransition] = useTransition();
  const [isLiking, setIsLiking] = useState(false);

  const [optimisticLiked, setOptimisticLiked] = useOptimistic(
    isLiked,
    (_currentState, newState: boolean) => newState
  );

  const { likeDao, unlikeDao } = useLikeDao();

  const handleLike = async () => {
    if (isAuthenticating || isLiking) {
      return;
    }

    const newLikeState = !optimisticLiked;

    startTransition(() => {
      setOptimisticLiked(newLikeState);
    });

    try {
      setIsLiking(true);
      const result = await withAuth(async () => {
        if (newLikeState) {
          return likeDao(dao.code);
        } else {
          return unlikeDao(dao.code);
        }
      })();

      if (result === null) {
        startTransition(() => {
          setOptimisticLiked(!newLikeState);
        });
      }
    } catch (error) {
      startTransition(() => {
        setOptimisticLiked(!newLikeState);
      });
      console.error('Like operation failed:', error);
      toast.error('Like operation failed. Please try again.');
    } finally {
      setIsLiking(false);
    }
  };

  const isLoading = isAuthenticating || isLiking;

  return (
    <button
      onClick={handleLike}
      disabled={isLoading}
      className={cn(
        'cursor-pointer transition-opacity hover:opacity-80 disabled:opacity-50',
        isAuthenticating && 'animate-pulse',
        className
      )}
      title={optimisticLiked ? 'Remove from favorites' : 'Add to favorites'}
    >
      <Image
        src={optimisticLiked ? '/favorited.svg' : '/favorite.svg'}
        alt={optimisticLiked ? 'liked' : 'not liked'}
        width={20}
        height={20}
        priority
      />
    </button>
  );
};
