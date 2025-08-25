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
    .replace(/\bnetwork\b/gi, '') // Remove 'network' case insensitive
    .trim(); // Remove leading/trailing whitespace
}

/**
 * Format timestamp to custom time ago format based on client requirements
 * @param timestamp Unix timestamp (as string) or ISO 8601 date string
 * @returns Custom formatted time ago string
 */
export function formatTimeAgo(timestamp: string) {
  if (!timestamp) return "";

  let date: Date;
  
  // Try to parse as ISO 8601 date string first
  if (timestamp.includes('T') || timestamp.includes('-')) {
    date = new Date(timestamp);
  } else {
    // Parse as numeric timestamp
    const timestampNum = Number(timestamp);
    if (isNaN(timestampNum)) {
      console.error(`Invalid timestamp: "${timestamp}"`);
      return "";
    }
    
    // Handle both seconds and milliseconds timestamps
    const isMilliseconds = timestampNum > 10000000000;
    date = new Date(isMilliseconds ? timestampNum : timestampNum * 1000);
  }
  
  const now = new Date();
  
  if (!date || isNaN(date.getTime())) {
    console.error(`Invalid date from timestamp: "${timestamp}"`);
    return "";
  }

  if (date.getTime() > now.getTime()) {
    // If date is in the future, return formatted date
    if (date.getFullYear() === now.getFullYear()) {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    } else {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    }
  }

  // Calculate differences in milliseconds
  const diffMs = now.getTime() - date.getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const diffWeeks = Math.floor(diffDays / 7);
  const diffMonths = Math.floor(diffDays / 30);

  // Different year - show "MMM D, YYYY" format (e.g., "Mar 20, 2024")
  if (date.getFullYear() !== now.getFullYear()) {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  // More than 1 month - show "MMM D" format for current year (e.g., "Apr 20")
  if (diffMonths >= 1) {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  // 3 weeks ago
  if (diffWeeks === 3) {
    return "three weeks ago";
  }

  // 2 weeks ago
  if (diffWeeks === 2) {
    return "two weeks ago";
  }

  // 1 week ago
  if (diffWeeks === 1) {
    return "one week ago";
  }

  // More than 3 weeks but less than 1 month
  if (diffDays >= 21) {
    return "last month";
  }

  // Yesterday
  if (diffDays === 1) {
    return "yesterday";
  }

  // More than 1 day but less than 1 week - show days
  if (diffDays >= 2) {
    return `${diffDays} days ago`;
  }

  // Less than 24 hours - show hours
  if (diffHours >= 1) {
    return `${diffHours} ${diffHours === 1 ? "hour" : "hours"} ago`;
  }

  // Less than 1 hour - show minutes
  if (diffMinutes >= 1) {
    return `${diffMinutes} ${diffMinutes === 1 ? "minute" : "minutes"} ago`;
  }

  return "just now";
}
