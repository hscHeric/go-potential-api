-- Seed: Criar administrador padrão
-- Email: admin@potential-idiomas.com
-- Senha: admin@potential-idiomas.com
-- IMPORTANTE: Troque a senha após primeiro login!

INSERT INTO auth (id, email, password_hash, role, status, created_at, updated_at)
VALUES (
    '92b0e486-c1ff-4bb5-95cb-42e4d01134c0',
    'admin@potential-idiomas.com',
    '$2a$10$R8rsaW3X9DizTCTOTrwL9Oixo7bnTej7PUEd.gJqx9VdSpwfsrc56',
    'admin',
    'active',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO NOTHING;
