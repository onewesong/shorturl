export type Link = {
  id: number;
  code: string;
  target_url: string;
  remark: string;
  tags: string[];
  enabled: boolean;
  click_count: number;
  created_at?: string;
  updated_at?: string;
};

export type CreateLinkInput = {
  code: string;
  target_url: string;
  remark: string;
  tags: string[];
};

export type UpdateLinkInput = {
  code: string;
  target_url: string;
  remark: string;
  tags: string[];
  enabled: boolean;
};

export type AuthSession = {
  authenticated: boolean;
  username: string;
};

export type VisitPoint = {
  bucket: string;
  clicks: number;
};

export type VisitBreakdown = {
  name: string;
  count: number;
};

export type VisitRecord = {
  visited_at: string;
  ip_masked: string;
  referer: string;
  referer_host: string;
  user_agent: string;
  client_name: string;
  client_type: string;
  device_type: string;
  os: string;
};

export type LinkAnalytics = {
  link: Link;
  range_days: number;
  recent_clicks: number;
  unique_ips: number;
  last_visited_at?: string;
  time_series: VisitPoint[];
  top_referrers: VisitBreakdown[];
  top_clients: VisitBreakdown[];
  recent_visits: VisitRecord[];
};
