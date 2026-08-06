import { api } from "../../lib/apiClient";

export type ExerciseType =
  | "audio_pronunciation"
  | "multiple_choice_translation"
  | "word_order"
  | "verb_conjugation"
  | "dictation"
  | "free_translation"
  | "matching_pairs"
  | "true_false";

export interface Exercise {
  id: string;
  category: string;
  difficulty: number;
  prompt_pt: string;
  expected_transcript: string;
  expected_romaji?: string;
  sprint_day_ref: number;
  language: string;
  scenario_id?: string;
  order_in_scenario?: number;
  exercise_type: ExerciseType;
  type_data?: unknown;
}

export interface Scenario {
  id: string;
  language: string;
  title_pt: string;
  context_description_pt: string;
  order_index: number;
}

export interface ScenarioDetail extends Scenario {
  exercises: Exercise[];
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

export function listExercises(sprintDay?: number, language?: string): Promise<Exercise[]> {
  const params = new URLSearchParams();
  if (sprintDay !== undefined) params.set("sprint_day", String(sprintDay));
  if (language) params.set("language", language);
  const query = params.toString();
  return api.get<Exercise[]>(`/exercises${query ? `?${query}` : ""}`);
}

export function getExercise(id: string): Promise<Exercise> {
  return api.get<Exercise>(`/exercises/${id}`);
}

export function getScenario(id: string): Promise<ScenarioDetail> {
  return api.get<ScenarioDetail>(`/scenarios/${id}`);
}

export function submitAttempt(exerciseId: string, audioBlob: Blob): Promise<AttemptResult> {
  const formData = new FormData();
  formData.append("audio", audioBlob, "attempt.webm");
  return api.post<AttemptResult>(`/exercises/${exerciseId}/attempts`, formData);
}

export function getAttemptHistory(exerciseId: string): Promise<Attempt[]> {
  return api.get<Attempt[]>(`/exercises/${exerciseId}/attempts`);
}

// ---------- Fase A: os 5 tipos binários (POST /answer) ----------

export interface MatchingPair {
  left: string;
  right: string;
}

// Só os campos relevantes ao exercise_type do exercício vão preenchidos —
// espelha validation.AnswerRequest em core-api/internal/exercises/validation.
export interface AnswerRequest {
  selected_index?: number;
  submitted_order?: string[];
  submitted_pairs?: MatchingPair[];
  answer?: boolean;
}

export interface AnswerResult {
  correct: boolean;
  correct_index?: number;
  correct_order?: string[];
  correct_pairs?: MatchingPair[];
  correct_answer?: boolean;
}

export function submitAnswer(exerciseId: string, req: AnswerRequest): Promise<AnswerResult> {
  return api.post<AnswerResult>(`/exercises/${exerciseId}/answer`, req);
}

// ---------- Fase A: dictation e free_translation (POST /text-attempt) ----------

export interface TextAttemptResult {
  transcript: string;
  // Texto contra o qual o backend de fato comparou — pra dictation é
  // sempre exercise.expected_transcript, pra free_translation é qual das
  // várias acceptable_answers teve o melhor score (o cliente não tem como
  // saber isso sozinho, só type_data.acceptable_answers com todas as opções).
  expected: string;
  score: number;
  verdict: Verdict;
  diff: DiffEntry[];
}

export function submitTextAttempt(exerciseId: string, transcript: string): Promise<TextAttemptResult> {
  return api.post<TextAttemptResult>(`/exercises/${exerciseId}/text-attempt`, { transcript });
}

// ---------- type_data por exercise_type (espelha core-api/internal/exercises/validation) ----------

export interface MultipleChoiceTranslationData {
  options: string[];
  correct_index: number;
}

export interface WordOrderData {
  shuffled_words: string[];
  correct_order: string[];
}

export interface VerbConjugationData {
  sentence_template: string;
  verb_infinitive: string;
  options: string[];
  correct_index: number;
}

export interface FreeTranslationData {
  acceptable_answers: string[];
}

export interface MatchingPairsData {
  pairs: MatchingPair[];
}

export interface TrueFalseData {
  statement: string;
  correct_answer: boolean;
}
