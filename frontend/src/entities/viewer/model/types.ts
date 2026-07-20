// The currently signed-in app user. Mirrors auth.User in the Go backend.
export interface Viewer {
  id: number;
  email: string;
  nickname: string;
  avatarUrl: string;
  emailVerified: boolean;
  createdAt: string;
}
