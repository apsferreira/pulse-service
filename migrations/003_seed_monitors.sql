-- Seed monitors for all IIT ecosystem services
INSERT INTO monitors (name, url, method, interval_seconds, timeout_seconds, is_active) VALUES
    ('Auth Service',              'https://auth.institutoitinerante.com.br/health',         'GET', 60, 10, true),
    ('Libri API',                 'https://libri.institutoitinerante.com.br/health',         'GET', 60, 10, true),
    ('Nitro API',                 'https://nitro.institutoitinerante.com.br/health',         'GET', 60, 10, true),
    ('Customer Service',          'https://customer.institutoitinerante.com.br/health',      'GET', 60, 10, true),
    ('Catalog API',               'https://catalog-api.institutoitinerante.com.br/health',   'GET', 60, 10, true),
    ('Checkout API',              'https://checkout-api.institutoitinerante.com.br/health',  'GET', 60, 10, true),
    ('Contracts Service',         'https://contracts.institutoitinerante.com.br/health',     'GET', 60, 10, true),
    ('Notification Service',      'https://notification.institutoitinerante.com.br/health',  'GET', 60, 10, true),
    ('Pulse API',                 'https://pulse-api.institutoitinerante.com.br/health',     'GET', 60, 10, true),
    ('Catalog Frontend',          'https://catalog.institutoitinerante.com.br',              'GET', 60, 10, true),
    ('Libri Frontend',            'https://libri.institutoitinerante.com.br',                'GET', 60, 10, true),
    ('Nitro Frontend',            'https://nitro.institutoitinerante.com.br',                'GET', 60, 10, true),
    ('Antonio Pedro Site',        'https://antoniopedro.com.br',                            'GET', 300, 15, true),
    ('Instituto Itinerante Site', 'https://institutoitinerante.com.br',                     'GET', 300, 15, true)
ON CONFLICT DO NOTHING;
