UPDATE exercises
SET scenario_id = NULL, order_in_scenario = NULL
WHERE language = 'en-US' AND scenario_id IN (
    SELECT id FROM scenarios WHERE language = 'en-US' AND title_pt IN (
        'Conhecer alguém pela primeira vez',
        'Fazer compras em uma loja',
        'Pedir comida em um restaurante',
        'Pedir informações na rua',
        'Lidar com uma emergência de saúde ou perda de documento',
        'Situações profissionais no trabalho'
    )
);

DELETE FROM scenarios WHERE language = 'en-US' AND title_pt IN (
    'Conhecer alguém pela primeira vez',
    'Fazer compras em uma loja',
    'Pedir comida em um restaurante',
    'Pedir informações na rua',
    'Lidar com uma emergência de saúde ou perda de documento',
    'Situações profissionais no trabalho'
);
