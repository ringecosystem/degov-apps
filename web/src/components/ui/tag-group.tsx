import { cn } from '@/lib/utils';
import type { DaoInfo } from '@/utils/config';

import ActiveTag from './active-tag';
import AgentEnabled from './agent-enabled';

const activeTagList = ['ACTIVE', 'PENDING', 'SUCCEEDED', 'QUEUED'];

interface TagGroupProps {
  dao?: DaoInfo;
  className?: string;
}
const TagGroup = ({ dao, className }: TagGroupProps) => {
  const chips = dao?.chips;
  const activeTag = chips?.find(
    (chip) => chip.chipCode === 'METRICS_STATE' && activeTagList.includes(chip.flag)
  );
  const agentEnabledTag = chips?.find((chip) => chip.chipCode === 'AGENT');

  return activeTag || agentEnabledTag ? (
    <div className={cn('flex items-center gap-[10px]', className)}>
      {agentEnabledTag && <AgentEnabled />}
      {activeTag && <ActiveTag />}
    </div>
  ) : null;
};

export default TagGroup;
