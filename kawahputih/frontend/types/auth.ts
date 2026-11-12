export interface AuthUser {
  id: string;
  name: string;
  email: string;
  phone: string;
  is_verified: boolean;
  roles: string[];
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  user: Omit<AuthUser, "roles">; // backend's auth.UserResponse doesn't include roles; roles come from /me-equivalent or the login usecase's role lookup
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface RegisterPayload {
  name: string;
  email: string;
  phone?: string;
  password: string;
}
