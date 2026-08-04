-- Importação de seed/japanese_scenarios_pilot.json: 3 cenários ja-JP com 4
-- exercícios encadeados cada. Dedupe verificado contra os seeds anteriores
-- (migrations 0004 e 0010) por expected_transcript exato: só 1 colisão real
-- — 'おはようございます' (cenário 1, exercício 1) já existia solto desde a
-- migration 0010 ('saudacoes_apresentacao', sprint_day_ref=1) — vinculada ao
-- cenário via UPDATE em vez de duplicada. Os outros 11 exercícios são novos.
-- sprint_day_ref 51-53 (1 por cenário, todos os exercícios do mesmo cenário
-- compartilham o dia): trilha própria, não colide com a numeração 1-50 do
-- currículo principal nem com os dias 1-30 do currículo en-US.
DO $$
DECLARE
    v_scenario_id UUID;
BEGIN
    -- Cenário 1: Cumprimentar um colega no trabalho de manhã
    INSERT INTO scenarios (language, title_pt, context_description_pt, order_index)
    VALUES (
        'ja-JP',
        'Cumprimentar um colega no trabalho de manhã',
        'Você chega no escritório e cruza com um colega. É de manhã, ambiente informal mas profissional.',
        1
    )
    RETURNING id INTO v_scenario_id;

    UPDATE exercises
    SET scenario_id = v_scenario_id, order_in_scenario = 1
    WHERE language = 'ja-JP' AND scenario_id IS NULL AND expected_transcript = 'おはようございます';

    INSERT INTO exercises (category, difficulty, prompt_pt, expected_transcript, expected_romaji, sprint_day_ref, language, scenario_id, order_in_scenario) VALUES
    ('cenario_trabalho_manha', 2, 'O colega pergunta como você está. Responda que está bem.', 'げんきです、ありがとうございます', 'genki desu, arigatou gozaimasu', 51, 'ja-JP', v_scenario_id, 2),
    ('cenario_trabalho_manha', 2, 'Pergunte de volta como ele está.', 'おげんきですか', 'o genki desu ka', 51, 'ja-JP', v_scenario_id, 3),
    ('cenario_trabalho_manha', 3, 'Se despeça dizendo que vai trabalhar bem / se esforçar (expressão comum ao começar o dia de trabalho).', 'きょうもがんばりましょう', 'kyou mo ganbarimashou', 51, 'ja-JP', v_scenario_id, 4);

    -- Cenário 2: Check-in no aeroporto
    INSERT INTO scenarios (language, title_pt, context_description_pt, order_index)
    VALUES (
        'ja-JP',
        'Check-in no aeroporto',
        'Você está no balcão de check-in de uma companhia aérea no Japão.',
        2
    )
    RETURNING id INTO v_scenario_id;

    INSERT INTO exercises (category, difficulty, prompt_pt, expected_transcript, expected_romaji, sprint_day_ref, language, scenario_id, order_in_scenario) VALUES
    ('cenario_aeroporto', 1, 'Entregue seu passaporte dizendo ''aqui está, por favor''.', 'パスポートです、どうぞ', 'pasupooto desu, douzo', 52, 'ja-JP', v_scenario_id, 1),
    ('cenario_aeroporto', 2, 'Pergunte qual é o portão de embarque.', 'とうじょうぐちはどこですか', 'toujouguchi wa doko desu ka', 52, 'ja-JP', v_scenario_id, 2),
    ('cenario_aeroporto', 3, 'Pergunte se o voo está atrasado.', 'フライトはおくれていますか', 'furaito wa okurete imasu ka', 52, 'ja-JP', v_scenario_id, 3),
    ('cenario_aeroporto', 1, 'Agradeça a atendente ao final.', 'ありがとうございました', 'arigatou gozaimashita', 52, 'ja-JP', v_scenario_id, 4);

    -- Cenário 3: Jantar em família
    INSERT INTO scenarios (language, title_pt, context_description_pt, order_index)
    VALUES (
        'ja-JP',
        'Jantar em família',
        'Você está jantando na casa da família da sua namorada/esposa no Japão.',
        3
    )
    RETURNING id INTO v_scenario_id;

    INSERT INTO exercises (category, difficulty, prompt_pt, expected_transcript, expected_romaji, sprint_day_ref, language, scenario_id, order_in_scenario) VALUES
    ('cenario_jantar_familia', 1, 'Antes de comer, diga a expressão tradicional de agradecimento pela comida.', 'いただきます', 'itadakimasu', 53, 'ja-JP', v_scenario_id, 1),
    ('cenario_jantar_familia', 1, 'Elogie a comida dizendo que está deliciosa.', 'とてもおいしいです', 'totemo oishii desu', 53, 'ja-JP', v_scenario_id, 2),
    ('cenario_jantar_familia', 2, 'Peça para passarem o sal.', 'しおをとってください', 'shio wo totte kudasai', 53, 'ja-JP', v_scenario_id, 3),
    ('cenario_jantar_familia', 1, 'Ao terminar de comer, diga a expressão tradicional de agradecimento final.', 'ごちそうさまでした', 'gochisousama deshita', 53, 'ja-JP', v_scenario_id, 4);
END $$;
