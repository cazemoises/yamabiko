import { api } from "../../lib/apiClient";

export interface PinProfile {
  id: string;
  display_name: string;
  accent_color?: string;
}

export function fetchPinProfiles(): Promise<PinProfile[]> {
  return api.get<PinProfile[]>("/auth/profiles");
}
