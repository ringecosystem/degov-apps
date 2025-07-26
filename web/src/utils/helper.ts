export type GetPathLevelType = {
  pathname: string;
  moduleName: string;
};

export type PathLevelType = {
  isFirstLevel: boolean;
  isSecondLevel: boolean;
  section?: string;
};

export function getPathLevel({ pathname, moduleName }: GetPathLevelType): PathLevelType {
  const parts = pathname.split('/').filter(Boolean);
  const isFirstLevel = parts.length === 2 && parts[0] === moduleName;
  const isSecondLevel = parts.length === 3 && parts[0] === moduleName;
  const section = isSecondLevel ? parts[2] : undefined;
  return { isFirstLevel, isSecondLevel, section };
}

/**
 * Format network name by removing 'Network' (case insensitive) and trimming whitespace
 * @param networkName - The network name to format
 * @returns Formatted network name
 */
export function formatNetworkName(networkName: string | undefined): string {
  if (!networkName) return 'Unknown';

  return networkName
    .replace(/network/gi, '') // Remove 'network' case insensitive
    .trim(); // Remove leading/trailing whitespace
}
