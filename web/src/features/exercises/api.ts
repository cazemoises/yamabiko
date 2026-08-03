import { api } from "../../lib/apiClient";

export interface Exercise {
  id: string;
  category: string;
  difficulty: number;
  prompt_pt: string;
  expected_transcript: string;
  expected_romaji?: string;
  sprint_day_ref: number;
}

export interface DiffEntry {
  op: string;
  position: number;
  expected?: string;
  actual?: string;
  pattern?: string;
}

export type Verdict = "PASS" | "PARTIAL" | "FAIL";

export interface AttemptResult {
  transcript: string;
  score: number;
  verdict: Verdict;
  diff: DiffEntry[];
  xp_gained: number;
}

export interface Attempt {
  id: string;
  user_id: string;
  exercise_id: string;
  audio_transcript: string;
  similarity_score: number;
  verdict: Verdict;
  phonetic_diff: DiffEntry[];
  created_at: string;
}

export function listExercises(sprintDay?: number): Promise<Exercise[]> {
  const query = sprintDay !== undefined ? `?sprint_day=${sprintDay}` : "";
  return api.get<Exercise[]>(`/exercises${query}`);
}

export function getExercise(id: string): Promise<Exercise> {
  return api.get<Exercise>(`/exercises/${id}`);
}

export function submitAttempt(exerciseId: string, audioBlob: Blob): Promise<AttemptResult> {
  const formData = new FormData();
  formData.append("audio", audioBlob, "attempt.webm");
  return api.post<AttemptResult>(`/exercises/${exerciseId}/attempts`, formData);
}

export function getAttemptHistory(exerciseId: string): Promise<Attempt[]> {
  return api.get<Attempt[]>(`/exercises/${exerciseId}/attempts`);
}
