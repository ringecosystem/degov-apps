'use client';

import { createContext, useContext } from 'react';

export interface AuthContextType {
  token: string | null;
  setToken: (token: string | null) => void;
  isAuthenticated: boolean;
  // URL auth support
  urlAddress: string | null;
  isUsingUrlAuth: boolean;
  clearUrlAuth: () => void;
  // Clear URL auth on 401/unauthorized errors
  clearUrlAuthOnError: () => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
