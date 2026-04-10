import { create } from "zustand";
import api from "./api";

interface User {
  id: string;
  email: string;
  name: string;
  status: string;
  roles: string[];
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  hydrate: () => void;
  hasRole: (role: string) => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: null,
  isAuthenticated: false,

  login: async (email: string, password: string) => {
    const response = await api.post("/auth/login", { email, password });
    const token = response.data.token;
    localStorage.setItem("token", token);

    const meResponse = await api.get("/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    });
    const user = meResponse.data as User;
    localStorage.setItem("user", JSON.stringify(user));
    set({ user, token, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    set({ user: null, token: null, isAuthenticated: false });
  },

  hydrate: () => {
    const token = localStorage.getItem("token");
    const userJson = localStorage.getItem("user");
    if (token && userJson) {
      try {
        const user = JSON.parse(userJson) as User;
        set({ user, token, isAuthenticated: true });
      } catch {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
      }
    }
  },

  hasRole: (role: string) => {
    const { user } = get();
    return user?.roles?.includes(role) ?? false;
  },
}));
