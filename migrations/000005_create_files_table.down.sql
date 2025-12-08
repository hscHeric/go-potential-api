-- Tabela centralizada de arquivos
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    metadata JSONB DEFAULT '{}',
    uploaded_by UUID REFERENCES auth(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices para performance
CREATE INDEX idx_files_entity ON files(entity_type, entity_id);
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by);
CREATE INDEX idx_files_created_at ON files(created_at DESC);

-- Trigger para updated_at
CREATE TRIGGER update_files_updated_at
    BEFORE UPDATE ON files
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comentários para documentação
COMMENT ON TABLE files IS 'Tabela centralizada para armazenamento de todos os arquivos do sistema';
COMMENT ON COLUMN files.entity_type IS 'Tipo de entidade: user_profile, user_document, activity_attachment, class_material, etc';
COMMENT ON COLUMN files.entity_id IS 'ID da entidade relacionada';
COMMENT ON COLUMN files.metadata IS 'Metadados adicionais em formato JSON (ex: dimensões de imagem, duração de vídeo, etc)';
