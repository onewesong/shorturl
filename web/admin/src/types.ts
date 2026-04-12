export type Link = {
  id: number;
  code: string;
  target_url: string;
  enabled: boolean;
  click_count: number;
  created_at?: string;
  updated_at?: string;
};

export type CreateLinkInput = {
  code: string;
  target_url: string;
};

export type UpdateLinkInput = {
  code: string;
  target_url: string;
  enabled: boolean;
};

export type AuthSession = {
  authenticated: boolean;
  username: string;
};
