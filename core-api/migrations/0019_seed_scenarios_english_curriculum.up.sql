-- Lacuna de seed: a migration 0014 criou 3 scenarios ja-JP, mas nunca existiu
-- equivalente pra en-US — os 30 exercícios da migration 0012 (categorias
-- saudacoes/compras/restaurante/direcoes/emergencia_saude/trabalho_social,
-- 5 por categoria, sprint_day_ref 1-30) ficaram soltos (scenario_id NULL)
-- desde então. Como a UI navega por scenario, esse conteúdo nunca aparecia
-- (confirmado por teste real: login como Vitória, zero exercícios em inglês
-- visíveis). Não é dado corrompido, é seed que faltou.
--
-- 1 scenario por categoria (6 no total), order_index 1-6 seguindo a ordem
-- crescente de sprint_day_ref (progressão de dificuldade já implícita no
-- currículo). Diferente da 0014 (que linkava por expected_transcript exato,
-- viável só porque cada exercício tinha texto único), aqui o link é por
-- category + sprint_day_ref via ROW_NUMBER() — mais robusto porque não exige
-- hardcodar UUID nem transcript de cada exercício, e a partição por categoria
-- com exatamente 5 linhas cada garante order_in_scenario 1-5 sem ambiguidade.
INSERT INTO scenarios (language, title_pt, context_description_pt, order_index) VALUES
('en-US', 'Conhecer alguém pela primeira vez', 'Você está puxando conversa com alguém que acabou de conhecer — se apresentando, perguntando como a pessoa está e se despedindo educadamente.', 1),
('en-US', 'Fazer compras em uma loja', 'Você está numa loja física perguntando preço, forma de pagamento, trocando um item com defeito ou procurando o provador.', 2),
('en-US', 'Pedir comida em um restaurante', 'Você acabou de sentar num restaurante e precisa pedir mesa, perguntar recomendação, avisar de uma alergia e fechar a conta.', 3),
('en-US', 'Pedir informações na rua', 'Você está perdido ou precisa de indicação — perguntando caminho, distância, se pode chamar um táxi.', 4),
('en-US', 'Lidar com uma emergência de saúde ou perda de documento', 'Situação urgente: você está passando mal, precisa de um médico, perdeu o passaporte ou precisa acionar a polícia.', 5),
('en-US', 'Situações profissionais no trabalho', 'Contexto de trabalho: se apresentar profissionalmente, se desculpar por atraso, remarcar reunião ou responder a uma proposta.', 6);

WITH scenario_map (category, title_pt) AS (
    VALUES
    ('saudacoes', 'Conhecer alguém pela primeira vez'),
    ('compras', 'Fazer compras em uma loja'),
    ('restaurante', 'Pedir comida em um restaurante'),
    ('direcoes', 'Pedir informações na rua'),
    ('emergencia_saude', 'Lidar com uma emergência de saúde ou perda de documento'),
    ('trabalho_social', 'Situações profissionais no trabalho')
),
ranked AS (
    SELECT
        e.id AS exercise_id,
        s.id AS scenario_id,
        ROW_NUMBER() OVER (PARTITION BY e.category ORDER BY e.sprint_day_ref) AS rn
    FROM exercises e
    JOIN scenario_map sm ON sm.category = e.category
    JOIN scenarios s ON s.language = 'en-US' AND s.title_pt = sm.title_pt
    WHERE e.language = 'en-US' AND e.scenario_id IS NULL
)
UPDATE exercises
SET scenario_id = ranked.scenario_id, order_in_scenario = ranked.rn
FROM ranked
WHERE exercises.id = ranked.exercise_id;
