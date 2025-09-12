'use client';

const TOKEN_KEY = 'degov_auth_token';
const ADDRESS_KEY = 'degov_auth_address';

class TokenManager {
  private get storage() {
    return typeof window !== 'undefined' ? sessionStorage : null;
  }

  getToken(): string | null {
    if (!this.storage) return null;
    return this.storage.getItem(TOKEN_KEY);
  }

  getAddress(): string | null {
    if (!this.storage) return null;
    return this.storage.getItem(ADDRESS_KEY);
  }

  setToken(token: string | null): void {
    if (!this.storage) return;

    if (token) {
      this.storage.setItem(TOKEN_KEY, token);
    } else {
      this.storage.removeItem(TOKEN_KEY);
    }
  }

  setAddress(address: string | null): void {
    if (!this.storage) return;

    if (address) {
      this.storage.setItem(ADDRESS_KEY, address);
    } else {
      this.storage.removeItem(ADDRESS_KEY);
    }
  }

  getAuthData(): { token: string | null; address: string | null } {
    return {
      token: this.getToken(),
      address: this.getAddress()
    };
  }
}

export const tokenManager = new TokenManager();

export const getToken = () => tokenManager.getToken();
