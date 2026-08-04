-- Remove só os 11 exercícios que esta migration efetivamente inseriu — não
-- toca em 'おはようございます', que já existia (migration 0010), só foi
-- vinculado; ele volta a ficar solto na linha de baixo.
DELETE FROM exercises WHERE language = 'ja-JP' AND expected_transcript IN (
    'げんきです、ありがとうございます',
    'おげんきですか',
    'きょうもがんばりましょう',
    'パスポートです、どうぞ',
    'とうじょうぐちはどこですか',
    'フライトはおくれていますか',
    'ありがとうございました',
    'いただきます',
    'とてもおいしいです',
    'しおをとってください',
    'ごちそうさまでした'
);

UPDATE exercises
SET scenario_id = NULL, order_in_scenario = NULL
WHERE language = 'ja-JP' AND expected_transcript = 'おはようございます';

DELETE FROM scenarios WHERE title_pt IN (
    'Cumprimentar um colega no trabalho de manhã',
    'Check-in no aeroporto',
    'Jantar em família'
);
