'use client';

const TOKEN_KEY = 'degov_auth_token';
const REMOTE_TOKEN_KEY = 'degov_remote_auth_token';

class TokenManager {
  // 本地token管理
  getToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(TOKEN_KEY);
  }

  setToken(token: string | null): void {
    if (typeof window === 'undefined') return;
    
    if (token) {
      localStorage.setItem(TOKEN_KEY, token);
    } else {
      localStorage.removeItem(TOKEN_KEY);
    }
    
    // Dispatch event for auth context to listen to
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('auth-token-change', {
        detail: { token }
      }));
    }
  }

  clearToken(): void {
    this.setToken(null);
  }

  // 远程token管理
  getRemoteToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(REMOTE_TOKEN_KEY);
  }

  setRemoteToken(token: string | null): void {
    if (typeof window === 'undefined') return;
    
    if (token) {
      localStorage.setItem(REMOTE_TOKEN_KEY, token);
    } else {
      localStorage.removeItem(REMOTE_TOKEN_KEY);
    }
  }

  clearRemoteToken(): void {
    this.setRemoteToken(null);
  }

  // 清理所有token
  clearAllTokens(): void {
    this.clearToken();
    this.clearRemoteToken();
  }

  // 检查token是否存在且格式正确
  hasValidFormat(): boolean {
    const token = this.getToken();
    return !!(token && token.length > 10); // 简单的格式检查
  }

  hasValidRemoteFormat(): boolean {
    const token = this.getRemoteToken();
    return !!(token && token.length > 10); // 简单的格式检查
  }

  // 获取当前应该使用的token（优先使用本地token，然后是远程token）
  getCurrentToken(): string | null {
    return this.getToken() || this.getRemoteToken();
  }
}

// 创建单例实例
export const tokenManager = new TokenManager();

// 导出常用方法，保持向后兼容
export const getToken = () => tokenManager.getCurrentToken();
export const setToken = (token: string | null) => tokenManager.setToken(token);
export const clearToken = () => tokenManager.clearToken();

// 导出远程token管理方法
export const getRemoteToken = () => tokenManager.getRemoteToken();
export const setRemoteToken = (token: string | null) => tokenManager.setRemoteToken(token);
export const clearRemoteToken = () => tokenManager.clearRemoteToken();